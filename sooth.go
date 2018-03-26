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
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"time"
)

var nameWidth = 25

func init() {
	rand.Seed(time.Now().Unix())
}

// pingResponse records the results of one ping attempt.
type pingResponse struct {
	Target    string
	Time      time.Time
	LastReply time.Time
	Raw       []byte
	Pings     int
	Pongs     int
	Loss      int
	Min       float64
	Avg       float64
	Max       float64
	Dev       float64
}

type configuration struct {
	Verbose bool `json:"verbose"`
	Web     struct {
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
	v := flag.Bool("v", false, "Turn on verbose output.")
	flag.Parse()

	conf := configuration{}
	conf.Verbose = false
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

	widestName := 0
	for _, v := range conf.Targets {
		n := len(v)
		if widestName < n {
			widestName = n
		}
	}
	if widestName < nameWidth {
		nameWidth = widestName
	}

	if *v {
		conf.Verbose = *v
		fmt.Println("Sooth Copyright 2018 Paul Gorman. Released under the Simplified BSD License.")
		fmt.Println("Starting Sooth with verbose output.")
		fmt.Println(time.Now().Format(time.Stamp))
		fmt.Println("Using configuration file", *c)
		fmt.Printf("Monitoring %v targets.\n", len(conf.Targets))
		fmt.Printf("Press ENTER for a summary of results.\n")
		fmt.Printf("Serving web API at http://%s:%s/api/v1/\n\n", conf.Web.IP, conf.Web.Port)
	}

	return conf
}

// historian brokers access to the history of pingResponses for all targests.
func historian(conf *configuration, output chan string, h chan pingResponse, report chan string, web chan []pingResponse) {
	hist := make(map[string]*ring.Ring, len(conf.Targets))
	resultString := `%-` + strconv.Itoa(nameWidth) + `s %6v/%-6v %3v%% loss %8.2f ms avg, %8.2f ms mdev`
	var r pingResponse
	var q string
	for {
		select {
		case r = <-h:
			if _, ok := hist[r.Target]; !ok {
				hist[r.Target] = ring.New(conf.Ping.HistoryLength)
			}
			hist[r.Target].Value = r
			hist[r.Target] = hist[r.Target].Next()
		case q = <-report:
			switch q {
			case "":
				results := make([]string, 0, len(conf.Targets))
				for _, v := range hist {
					r := tally(v)
					results = append(results, fmt.Sprintf(resultString, r.Target, r.Pongs, r.Pings, r.Loss, r.Avg, r.Dev))
				}
				sort.Strings(results)
				for _, v := range results {
					output <- v
				}
			default:
				r := tally(hist[q])
				output <- fmt.Sprintf("                ↳ %s %v/%v %v%% loss, %.2f ms avg, %.2f ms mdev", q, r.Pongs, r.Pings, r.Loss, r.Avg, r.Dev)
				if r.Pongs < 1 && !r.LastReply.IsZero() {
					output <- fmt.Sprintf("                  Last reply %v ago.", time.Now().Sub(r.LastReply).Round(time.Second))
				}
			}
		case <-web:
			results := make([]pingResponse, 0, len(conf.Targets))
			for _, v := range hist {
				results = append(results, tally(v))
			}
			web <- results
		}
	}
}

// ping runs system pings against a target, and reports the results.
func ping(t string, conf *configuration, output chan string, history chan pingResponse, report chan string) {
	var err error
	var r pingResponse
	r.Target = t
	lr := regexp.MustCompile(conf.Ping.LossReportRE)
	rr := regexp.MustCompile(conf.Ping.RTTReportRE)
	for {
		// Stagger the start times of the pings.
		time.Sleep(time.Duration(rand.Intn(250*len(conf.Targets))) * time.Millisecond)
		r.Time = time.Now()
		r.Raw, err = exec.Command("ping", "-c", strconv.Itoa(conf.Ping.PacketCount), "-i", strconv.FormatFloat(conf.Ping.PacketInterval, 'f', -1, 64), t).Output()
		if err != nil && conf.Verbose {
			log.Println(t, "ping failed:", err)
		}
		sp := bytes.Split(r.Raw, []byte("\n"))
		if len(sp) > 3 {
			// Check latency.
			if m := lr.FindSubmatch(sp[len(sp)-3]); m != nil {
				r.Pongs, err = strconv.Atoi(string(m[2]))
				if err != nil {
					log.Println(err)
				}
				r.Pings, err = strconv.Atoi(string(m[1]))
				if err != nil {
					log.Println(err)
				}
			}
			// Check packet loss.
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
			}
		}
		history <- r
		if (r.Pings - r.Pongs) > conf.Ping.PacketThreshold {
			if conf.Verbose {
				output <- fmt.Sprintf("\n%v\n    ↳----------↴", string(r.Raw))
			}
			output <- fmt.Sprintf("%v %v %v", time.Now().Format(time.Stamp), t, string(sp[len(sp)-3]))
			report <- r.Target
		}
		if r.Dev > (r.Avg * conf.Ping.JitterMultiple) {
			if conf.Verbose {
				output <- fmt.Sprintf("\n%v\n    ↳----------↴", string(r.Raw))
			}
			output <- fmt.Sprintf("%v %v %v", time.Now().Format(time.Stamp), t, string(sp[len(sp)-2]))
			report <- r.Target
		}
		time.Sleep(time.Second * time.Duration(conf.Ping.CheckInterval))
	}
}

// tally summarizes ping responses.
func tally(hist *ring.Ring) pingResponse {
	var r pingResponse
	var rtt float64
	var min float64
	var max float64
	var i float64
	d := make([]float64, 0, hist.Len())

	hist.Do(func(v interface{}) {
		if v != nil {
			r.Target = v.(pingResponse).Target
			r.Time = v.(pingResponse).Time
			r.Pings += v.(pingResponse).Pings
			r.Pongs += v.(pingResponse).Pongs
			rtt += v.(pingResponse).Avg
			if v.(pingResponse).Min < min || min == 0 {
				r.Min = v.(pingResponse).Min
			}
			if v.(pingResponse).Max > max {
				r.Max = v.(pingResponse).Max
			}
			if v.(pingResponse).Pongs > 0 && v.(pingResponse).Time.After(r.LastReply) {
				r.LastReply = v.(pingResponse).Time
			}
			d = append(d, v.(pingResponse).Dev)
			i++
		}
	})

	r.Avg = rtt / i
	if math.IsNaN(r.Avg) {
		r.Avg = -1
	}

	r.Loss = 100 - int(math.Round((float64(r.Pongs)/float64(r.Pings))*100.0))
	if r.Loss < 0 || r.Loss > 100 {
		r.Loss = 100
	}

	sort.Float64s(d)
	if len(d) > 0 {
		r.Dev = d[len(d)/2]
	} else {
		r.Dev = -1
	}

	return r
}

func main() {
	conf := configure()
	output := make(chan string)
	history := make(chan pingResponse)
	report := make(chan string)
	web := make(chan []pingResponse)
	go historian(&conf, output, history, report, web)
	for _, t := range conf.Targets {
		go ping(t, &conf, output, history, report)
	}

	go func() {
		for {
			fmt.Println(<-output)
		}
	}()

	go func() {
		var input string
		r := bufio.NewReader(os.Stdin)
		for {
			input, _ = r.ReadString('\n')
			switch input {
			// TODO Add other commands?
			default:
				report <- ""
			}
		}
	}()

	http.HandleFunc("/api/v1/conf", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(conf)
	})
	http.HandleFunc("/api/v1/history", func(w http.ResponseWriter, r *http.Request) {
		web <- nil
		h := <-web
		json.NewEncoder(w).Encode(h)
	})
	log.Fatal(http.ListenAndServe(conf.Web.IP+":"+conf.Web.Port, nil))
}
