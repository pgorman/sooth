// Sooth is a simple network monitor.
package main

import (
	"bufio"
	"bytes"
	"container/ring"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"time"
)

var prompt = "> "

func init() {
	rand.Seed(time.Now().Unix())
}

// pingResponse records the results of one ping attempt.
type pingResponse struct {
	Target string
	Time   time.Time
	Raw    []byte
	Pings  int
	Pongs  int
	Loss   int
	Min    float64
	Avg    float64
	Max    float64
	Dev    float64
}

type configuration struct {
	Debug bool `json:"debug"`
	Web   struct {
		IP   string `json:"ip"`
		Port string `json:"port"`
	} `json:"web"`
	Ping struct {
		CheckInterval   int     `json:"checkInterval"`
		HistoryLength   int     `json:"historyLength"`
		PacketCount     int     `json:"packetCount"`
		PacketInterval  float64 `json:"packetInterval"`
		JitterMultiple  float64 `json:"jitterMultiple"`
		PacketThreshold int     `json:"packetThreshold"`
		LossReportRE    string  `json:"lossReportRE"`
		RTTReportRE     string  `json:"rttReportRE"`
	} `json:"ping"`
	Targets []string `json:"targets"`
}

// configure sets configuration defaults, then overrides them with values from the config file and command line arguments.
func configure() configuration {
	c := flag.String("c", "${XDG_CONFIG_HOME}/sooth.conf", "Full path to Sooth configuration file.")
	d := flag.Bool("d", false, "Turn on debuggin messages.")
	flag.Parse()

	conf := configuration{}
	conf.Debug = false
	conf.Web.IP = "127.0.0.1"
	conf.Web.Port = "9444"
	conf.Ping.CheckInterval = 50
	conf.Ping.HistoryLength = 100
	conf.Ping.PacketCount = 10
	conf.Ping.PacketInterval = 1.0
	conf.Ping.JitterMultiple = 2.0
	conf.Ping.PacketThreshold = 1
	conf.Ping.LossReportRE = `^(\d+) packets transmitted, (\d+) .+ (\d+)% packet loss.*`
	conf.Ping.RTTReportRE = `^r.+ (\d+\.\d+)/(\d+\.\d+)/(\d+\.\d+)/(\d+\.\d+) ms$`

	f, err := os.Open(os.ExpandEnv(*c))
	if err != nil {
		log.Fatal("error opening config file: ", err)
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&conf)
	if err != nil {
		log.Fatal("error decoding config JSON: ", err)
	}

	if *d {
		conf.Debug = *d
		fmt.Println("Sooth Copyright 2018 Paul Gorman. Released under the Simplified BSD License.")
		fmt.Println(time.Now().Format(time.Stamp), "Starting Sooth with debugging...")
		fmt.Println("Using configuration file", *c)
		fmt.Printf("Monitoring %v targets.\n", len(conf.Targets))
	}

	return conf
}

// cui provides the interactive command line interface.
func cui(report chan string) {
	var in string
	r := bufio.NewReader(os.Stdin)
	fmt.Printf(prompt)
	for {
		in, _ = r.ReadString('\n')
		switch in {
		case "ls\n":
			fmt.Println("list!")
			fmt.Printf(prompt)
		default:
			report <- ""
		}
	}
}

// historian brokers access to the history of pingResponses for all targests.
func historian(conf *configuration, h chan pingResponse, report chan string) {
	hist := make(map[string]*ring.Ring, len(conf.Targets))
	var r pingResponse
	for {
		select {
		case r = <-h:
			if _, ok := hist[r.Target]; !ok {
				hist[r.Target] = ring.New(conf.Ping.HistoryLength)
			}
			hist[r.Target].Value = r
			hist[r.Target] = hist[r.Target].Next()
		case <-report:
			results := make([]string, 0, len(conf.Targets))
			for _, v := range hist {
				r := tally(v)
				// TODO Set the printf target name field with according to the longest name?
				results = append(results, fmt.Sprintf("%-12s %6v/%-6v %3v%% loss", r.Target, r.Pongs, r.Pings, r.Loss))
			}
			sort.Strings(results)
			for _, v := range results {
				fmt.Println(v)
			}
			fmt.Printf(prompt)
		}
	}
}

// ping runs system pings against a target, and reports the results.
func ping(t string, conf *configuration, console chan string, h chan pingResponse) {
	var err error
	var r pingResponse
	r.Target = t
	lr := regexp.MustCompile(conf.Ping.LossReportRE)
	rr := regexp.MustCompile(conf.Ping.RTTReportRE)
	for {
		time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)
		r.Time = time.Now()
		r.Raw, err = exec.Command("ping", "-c", strconv.Itoa(conf.Ping.PacketCount), "-i", strconv.FormatFloat(conf.Ping.PacketInterval, 'f', -1, 64), t).Output()
		if err != nil && conf.Debug {
			log.Println(t, "ping failed:", err)
		}
		sp := bytes.Split(r.Raw, []byte("\n"))
		if len(sp) > 3 {
			if m := lr.FindSubmatch(sp[len(sp)-3]); m != nil {
				r.Pongs, err = strconv.Atoi(string(m[2]))
				if err != nil {
					log.Println(err)
				}
				r.Pings, err = strconv.Atoi(string(m[1]))
				if err != nil {
					log.Println(err)
				}
				if (r.Pings - r.Pongs) > conf.Ping.PacketThreshold {
					console <- fmt.Sprint(t, " ", string(sp[len(sp)-3]))
				}
			}
			if m := rr.FindSubmatch(sp[len(sp)-2]); m != nil {
				r.Min, err = strconv.ParseFloat(string(m[1]), 64)
				if err != nil {
					log.Println(err)
				}
				r.Avg, err = strconv.ParseFloat(string(m[2]), 64)
				if err != nil {
					log.Println(err)
				}
				r.Max, err = strconv.ParseFloat(string(m[3]), 64)
				if err != nil {
					log.Println(err)
				}
				r.Dev, err = strconv.ParseFloat(string(m[4]), 64)
				if err != nil {
					log.Println(err)
				}
				if r.Dev > (r.Avg * conf.Ping.JitterMultiple) {
					console <- fmt.Sprint(t, " ", string(sp[len(sp)-2]))
				}
			}
		}
		h <- r
		time.Sleep(time.Second * time.Duration(conf.Ping.CheckInterval))
	}
}

// tally summarizes ping responses.
func tally(hist *ring.Ring) pingResponse {
	var r pingResponse
	hist.Do(func(v interface{}) {
		if v != nil {
			r.Target = v.(pingResponse).Target
			r.Pings += v.(pingResponse).Pings
			r.Pongs += v.(pingResponse).Pongs
		}
	})
	r.Loss = 100 - int(math.Round((float64(r.Pongs)/float64(r.Pings))*100.0))
	return r
}

func main() {
	if runtime.GOOS != "linux" {
		log.Println("linux-like platform expected but not detected")
	}
	conf := configure()
	console := make(chan string)
	history := make(chan pingResponse)
	report := make(chan string)
	go historian(&conf, history, report)
	for _, t := range conf.Targets {
		go ping(t, &conf, console, history)
	}
	go cui(report)
	for {
		log.Println(<-console)
		fmt.Printf(prompt)
	}
}
