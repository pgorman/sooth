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
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// pingResponse records the results of one ping attempt.
type pingResponse struct {
	Target string
	Time   time.Time
	Raw    []byte
	Loss   int
	Pkts   int
	Min    float64
	Avg    float64
	Max    float64
	Dev    float64
}

// target is a host we ping. Its address can be a hostname or IP address.
type target struct {
	ID      string
	Name    string `json:"name"`
	Address string `json:"address"`
}

type configuration struct {
	Debug bool `json:"debug"`
	Web   struct {
		IP   string `json:"ip"`
		Port string `json:"port"`
	} `json:"web"`
	Ping struct {
		CheckInterval  int    `json:"checkInterval"`
		HistoryLength  int    `json:"historyLength"`
		PacketCount    int    `json:"packetCount"`
		PacketInterval string `json:"packetInterval"`
		LossReportRE   string `json:"lossReportRE"`
		RTTReportRE    string `json:"rttReportRE"`
	} `json:"ping"`
	Targets []target `json:"targets"`
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
	conf.Ping.PacketInterval = "1.0"
	conf.Ping.LossReportRE = `^\d+ packets transmitted, (\d+) .+ (\d+)% packet loss.*`
	conf.Ping.RTTReportRE = `^r.+ (\d+\.\d+)/(\d+\.\d+)/(\d+\.\d+)/(\d+\.\d+) ms$`

	f, err := os.Open(os.ExpandEnv(*c))
	if err != nil {
		log.Fatal("error opening config file: ", err)
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&conf)
	if err != nil {
		log.Fatal("error decodng config JSON: ", err)
	}

	if *d {
		conf.Debug = *d
		fmt.Println("Starting Sooth with debugging.")
		fmt.Println("Using configuration file", *c)
		fmt.Println(time.Now().Format(time.RFC1123))
		fmt.Printf("Monitoring %v targets.\n", len(conf.Targets))
	}

	return conf
}

// historian brokers access to the history of pingResponses for all targests.
func historian(conf *configuration, h chan pingResponse) {
	hist := make(map[string]*ring.Ring, len(conf.Targets))
	var r pingResponse
	for {
		r = <-h
		if _, ok := hist[r.Target]; !ok {
			hist[r.Target] = ring.New(conf.Ping.HistoryLength)
		}
		hist[r.Target].Value = r
		hist[r.Target].Next()
	}
}

// ping runs system pings against a target, and reports the results.
func ping(t target, conf *configuration, console chan string, h chan pingResponse) {
	var err error
	var r pingResponse
	r.Target = t.Address
	lr := regexp.MustCompile(conf.Ping.LossReportRE)
	rr := regexp.MustCompile(conf.Ping.RTTReportRE)
	for {
		time.Sleep((time.Second * time.Duration(conf.Ping.CheckInterval)) + (time.Duration(rand.Intn(2000)) * time.Millisecond))
		r.Time = time.Now()
		r.Raw, err = exec.Command("ping", "-c", strconv.Itoa(conf.Ping.PacketCount), "-i", conf.Ping.PacketInterval, t.Address).Output()
		if err != nil && conf.Debug {
			log.Println(t.Address, "ping failed:", err)
		}
		sp := bytes.Split(r.Raw, []byte("\n"))
		if len(sp) > 3 {
			if m := lr.FindSubmatch(sp[len(sp)-3]); m != nil {
				r.Loss, err = strconv.Atoi(string(m[2]))
				if err != nil {
					log.Println(err)
				}
				r.Pkts, err = strconv.Atoi(string(m[1]))
				if err != nil {
					log.Println(err)
				}
				if (conf.Ping.PacketCount - r.Pkts) > 1 {
					console <- fmt.Sprint(t.Name, " ", string(sp[len(sp)-3]))
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
				if r.Dev > (r.Min * 2.0) {
					console <- fmt.Sprint(t.Name, " ", string(sp[len(sp)-2]))
				}
			}
		}
		h <- r
	}
}

func main() {
	if runtime.GOOS != "linux" {
		log.Println("linux-like platform expected but not detected")
	}
	conf := configure()
	console := make(chan string)
	history := make(chan pingResponse)
	go historian(&conf, history)
	for _, t := range conf.Targets {
		go ping(t, &conf, console, history)
	}
	go func() {
		var in string
		r := bufio.NewReader(os.Stdin)
		fmt.Printf("> ")
		for {
			in, _ = r.ReadString('\n')
			switch in {
			case "ls\n":
				fmt.Println("list!")
			default:
			}
			fmt.Printf("> ")
		}
	}()
	for {
		log.Println(<-console)
	}
}
