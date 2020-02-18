// Sooth monitors ICMP ping responses from a group of network hosts.
// Copyright (C) 2020 Paul Gorman
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sparrc/go-ping"
)

var (
	checkInterval    int
	historyLength    int
	infoFmt          string
	jitterThreshold  int64
	latencyThreshold int64
	lossTolerance    int
	nameWidth        int
	pendFmt          string
	pingCount        int
	pingInterval     time.Duration
	pingSize         int
	pingTimeout      time.Duration
	printMu          sync.Mutex
	quiet            bool
	raw              bool
	startTime        time.Time
	syncPings        bool
	verbose          bool
	warnFmt          string
)

// host holds basic info and history for a target we monitor.
type host struct {
	Name       string
	LastReply  time.Time
	LastRTTs   []time.Duration
	Stats      []*ping.Statistics
	StatsIndex int
}

// monitor collects a set of ping responses to from a host.
func monitor(h *host, wg *sync.WaitGroup, l int) {
	defer wg.Done()
	if !syncPings {
		// Stagger the start times of the pings.
		// This is generally desirable, but can hide correspondences in missing responses.
		time.Sleep(time.Duration(rand.Intn(250*l)) * time.Millisecond)
	}

	pinger, err := ping.NewPinger(h.Name)
	if err != nil {
		log.Fatal(err)
	}

	pinger.Count = pingCount
	pinger.Interval = pingInterval
	pinger.Size = pingSize
	pinger.Timeout = pingTimeout
	pinger.SetPrivileged(false)
	if raw {
		pinger.SetPrivileged(true)
	}

	for i := range h.LastRTTs {
		h.LastRTTs[i] = 0
	}

	pinger.OnRecv = func(pkt *ping.Packet) {
		h.LastRTTs[pkt.Seq] = pkt.Rtt
		h.LastReply = time.Now()
	}

	pinger.OnFinish = func(stats *ping.Statistics) {
		if h.StatsIndex >= historyLength-1 {
			h.StatsIndex = 0
		}
		h.Stats[h.StatsIndex] = stats
		h.StatsIndex++
		warn(h, quiet)
	}
	pinger.Run()
	time.Sleep(time.Second * time.Duration(checkInterval))
}

// info generates a summary of a host's history and current status.
func info(h *host) string {
	if h.Stats[0] == nil {
		return fmt.Sprintf(pendFmt, h.Name)
	}
	var sdev int
	var pongs int
	var pings int
	var rtt int
	var l int

	for _, s := range h.Stats {
		if s == nil {
			break
		}
		sdev += int(s.StdDevRtt)
		pongs += s.PacketsRecv
		pings += s.PacketsSent
		rtt += int(s.AvgRtt)
		l++
	}

	return fmt.Sprintf(infoFmt, h.Name, pongs, pings, 100-pongs*100/pings*100/100,
		//ms := (h.LastRTTs[i] + time.Millisecond).Milliseconds()
		(time.Duration(rtt/l) + time.Millisecond).Milliseconds(),
		(time.Duration(sdev/l) + time.Millisecond).Milliseconds())
}

// warn prints a detailed trouble message when a host fails a test.
func warn(h *host, quiet bool) {
	if h.StatsIndex == 0 {
		return
	}
	s := h.Stats[h.StatsIndex-1]
	loss := false
	jitter := false
	var jc int
	var lastReply string
	woes := make([]string, 0, 3)
	rtts := "    seq:ms"

	if !h.LastReply.IsZero() {
		lastReply = h.LastReply.Format(time.Stamp)
	}

	if s.PacketsSent-s.PacketsRecv > lossTolerance {
		woes = append(woes, "Packet Loss")
		loss = true
	}
	if s.AvgRtt.Milliseconds() > latencyThreshold {
		woes = append(woes, "Latency")
	}
	for i := 0; i < pingCount; i++ {
		ms := (h.LastRTTs[i] + time.Millisecond).Milliseconds()
		if h.LastRTTs[i] == 0 {
			rtts += fmt.Sprintf("%3d:__ ", i)
		} else {
			rtts += fmt.Sprintf("%3d:%-3d", i, ms)
		}
		if !jitter && i > 0 && ms-h.LastRTTs[i-1].Milliseconds() > jitterThreshold {
			jc++
			if jc > 1 {
				woes = append(woes, "Jitter")
				jitter = true
			}
		}
	}
	if len(woes) == 0 {
		return
	}

	if quiet {
		return
	}

	printMu.Lock()

	fmt.Printf(warnFmt, h.Name, strings.Join(woes, ", "), lastReply)
	if loss {
		fmt.Printf("    %v%% loss (%d/%d replies received)\n",
			s.PacketLoss, s.PacketsRecv, s.PacketsSent)
		if s.PacketsRecv == 0 && !h.LastReply.IsZero() {
			fmt.Printf("    Last reply %s (%v ago)\n", lastReply,
				time.Now().Sub(h.LastReply).Round(time.Second))
		}
	}

	if s.PacketLoss < 100 && len(woes) > 0 {
		fmt.Printf("    RTTs %v min, %v avg, %v max, %v stddev\n",
			s.MinRtt.Round(time.Millisecond), s.AvgRtt.Round(time.Millisecond),
			s.MaxRtt.Round(time.Millisecond), s.StdDevRtt.Round(time.Millisecond))
		fmt.Println(rtts)
	}

	fmt.Println("    History ", info(h)[nameWidth:])

	printMu.Unlock()
}

func init() {
	startTime = time.Now()
	rand.Seed(startTime.Unix())
}

func main() {
	var pi = flag.Int("i", 1, "interval between sending each ping in seconds")
	var pt = flag.Int("W", 1, "timeout in seconds for ping replies")
	var hostsFile string
	flag.IntVar(&pingCount, "c", 10, "number of pings to send per round")
	flag.IntVar(&checkInterval, "check-interval", 50, "seconds to wait between rounds of pings")
	flag.StringVar(&hostsFile, "f", "", "path to text file listing hosts to ping, one per line")
	flag.IntVar(&historyLength, "history-lenth", 60, "number of rounds of pings to keep")
	flag.Int64Var(&jitterThreshold, "jitter-threshold", 50, "alet on jitter over this in ms")
	flag.Int64Var(&latencyThreshold, "latency-threshold", 150, "alert for avg latency over this in millisenconds")
	flag.IntVar(&lossTolerance, "loss-tolerance", 1, "number of lost packets per round to ignore without printing a warning")
	flag.BoolVar(&quiet, "q", false, "produce output only upon the user's request")
	flag.IntVar(&pingSize, "s", 56, "bytes of data in each packet")
	flag.BoolVar(&syncPings, "sync", false, "syncronize the start times of the pingers")
	flag.BoolVar(&raw, "raw", false, "use priviledged raw ICMP sockets")
	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.Parse()
	pingTimeout = time.Duration(*pt*pingCount) * time.Second
	pingInterval = time.Duration(*pi) * time.Second

	hosts := make([]host, 0, 10)
	if hostsFile != "" {
		f, err := os.Open(hostsFile)
		if err != nil {
			log.Fatal("error opening hosts file: ", err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var h host
			h.Name = scanner.Text()
			if h.Name == "" || h.Name[0] == '#' || h.Name[0:2] == "//" {
				continue
			}
			h.LastRTTs = make([]time.Duration, pingCount)
			h.Stats = make([]*ping.Statistics, historyLength)
			hosts = append(hosts, h)
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
	for _, a := range flag.Args() {
		var h host
		h.Name = a
		h.LastRTTs = make([]time.Duration, pingCount)
		h.Stats = make([]*ping.Statistics, historyLength)
		hosts = append(hosts, h)
	}

	for _, h := range hosts {
		n := len(h.Name)
		if nameWidth < n {
			nameWidth = n
		}
	}
	infoFmt = "%-" + strconv.Itoa(nameWidth) + "s %6v/%-6v %3v%% loss %6dms avg rtt %6dms mdev"
	pendFmt = "%-" + strconv.Itoa(nameWidth) + "s results pending..."
	warnFmt = "%-" + strconv.Itoa(nameWidth) + "s  %-30s  %v\n"

	if verbose && !quiet {
		fmt.Println("Sooth Copyright 2018 Paul Gorman. Released under the GPLv3 License.")
		fmt.Println("Starting Sooth with verbose output.")
		fmt.Println(time.Now().Format(time.Stamp))
		fmt.Println("Using hosts file", hostsFile)
		fmt.Printf("Monitoring %v target(s).\n", len(hosts))
		fmt.Printf("Enter ? for help.\n")
	}

	go func() {
		var input string
		r := bufio.NewReader(os.Stdin)
		for {
			input, _ = r.ReadString('\n')
			switch strings.TrimSpace(input) {
			case "?":
				printMu.Lock()
				fmt.Println(`
Sooth is silent unless it either detects a problem with a host
or the user enters one of these:

  ?      Print this help.
  ENTER  Report on recently troubled hosts.
  a      Report on all hosts.
  q      Quit Sooth.

Run Sooth like 'sooth --help' for a list of command-line flags.
See also https://github.com/pgorman/sooth.
			`)
				printMu.Unlock()
			case "a":
				printMu.Lock()
				for _, h := range hosts {
					fmt.Println(info(&h))
				}
				fmt.Println()
				printMu.Unlock()
			case "q":
				os.Exit(0)
			default:
				for _, h := range hosts {
					warn(&h, false)
				}
				fmt.Println()
			}
		}
	}()

	for {
		var wg sync.WaitGroup
		for i := range hosts {
			wg.Add(1)
			go monitor(&hosts[i], &wg, len(hosts))
		}
		wg.Wait()
	}

}
