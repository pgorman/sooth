// Sooth is a simple network monitor.
package main

import (
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

type pingResponse struct {
	Time time.Time
	Raw  []byte
	Loss int
	Pkts int
	Min  float64
	Avg  float64
	Max  float64
	Dev  float64
}

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
		PacketCount    string `json:"packetCount"`
		PacketInterval string `json:"packetInterval"`
		LossReportRE   string `json:"lossReportRE"`
		RTTReportRE    string `json:"rttReportRE"`
	} `json:"ping"`
	Targets []target `json:"targets"`
}

func configure() configuration {
	c := flag.String("c", "${XDG_CONFIG_HOME}/sooth.conf", "Full path to Sooth configuration file.")
	d := flag.Bool("d", false, "Turn on debuggin messages.")
	flag.Parse()

	conf := configuration{}
	conf.Debug = false
	conf.Web.IP = "127.0.0.1"
	conf.Web.Port = "9444"
	conf.Ping.CheckInterval = 60
	conf.Ping.HistoryLength = 100
	conf.Ping.PacketCount = "5"
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

	for k, v := range conf.Targets {
		conf.Targets[k].ID = targetID(v.Name)
	}

	if *d {
		conf.Debug = *d
		fmt.Println("Starting Sooth with debugging.")
		fmt.Println("Using configuration file", *c)
		fmt.Printf("Monitoring %v targets.\n", len(conf.Targets))
	}

	return conf
}

func ping(t target, conf *configuration, console chan string) {
	var err error
	var r pingResponse
	lr := regexp.MustCompile(conf.Ping.LossReportRE)
	rr := regexp.MustCompile(conf.Ping.RTTReportRE)
	h := ring.New(conf.Ping.HistoryLength)
	for {
		time.Sleep((time.Second * time.Duration(conf.Ping.CheckInterval)) + (time.Duration(rand.Intn(2000)) * time.Millisecond))
		r.Time = time.Now()
		r.Raw, err = exec.Command("ping", "-c", conf.Ping.PacketCount, "-i", conf.Ping.PacketInterval, t.Address).Output()
		if err != nil && conf.Debug {
			log.Println(t.Address, "ping failed:", err)
		}
		sp := bytes.Split(r.Raw, []byte("\n"))
		if len(sp) > 3 {
			if m := lr.FindSubmatch(sp[len(sp)-3]); m != nil {
				loss := string(m[2])
				pkts := string(m[1])
				if pkts != conf.Ping.PacketCount || loss != "0" {
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
		h.Value = r
		h = h.Next()
	}
}

func targetID(s string) string {
	return fmt.Sprintf("%08x", s)
}

func main() {
	if runtime.GOOS != "linux" {
		log.Println("linux-like platform expected but not detected")
	}
	conf := configure()
	console := make(chan string)
	for _, t := range conf.Targets {
		go ping(t, &conf, console)
	}
	for {
		log.Println(<-console)
	}
}
