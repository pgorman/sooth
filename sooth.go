// Sooth is a simple network monitor.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
)

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
		CheckInterval  string `json:"checkInterval"`
		HistoryLength  string `json:"historyLength"`
		PacketCount    string `json:"packetCount"`
		PacketInterval string `json:"packetInterval"`
		LossReportRE   string `json:"lossReportRE"`
		RTTReportRE    string `json:"rttReportRE"`
	} `json:"ping"`
	Targets []target `json:"targets"`
}

type pingResponse struct {
	exitStatus string
	raw        []byte
}

func configure() configuration {
	c := flag.String("c", "/etc/sooth.conf", "Full path to Sooth configuration file.")
	d := flag.Bool("d", false, "Turn on debuggin messages.")
	flag.Parse()

	conf := configuration{}
	conf.Debug = false
	conf.Web.IP = "127.0.0.1"
	conf.Web.Port = "9444"
	conf.Ping.CheckInterval = "60"
	conf.Ping.HistoryLength = "100"
	conf.Ping.PacketCount = "5"
	conf.Ping.PacketInterval = "0.3"
	conf.Ping.LossReportRE = `^\d+ packets transmitted, (\d+) .+ (\d+)% packet loss.*`
	conf.Ping.RTTReportRE = `^r.+ (\d+\.\d+)/(\d+\.\d+)/(\d+\.\d+)/(\d+\.\d+) ms$`

	f, err := os.Open(*c)
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

func ping(t *target, conf *configuration) {
	var err error
	var r pingResponse
	lr := regexp.MustCompile(conf.Ping.LossReportRE)
	rr := regexp.MustCompile(conf.Ping.RTTReportRE)

	r.raw, err = exec.Command("ping", "-c", conf.Ping.PacketCount, "-i", conf.Ping.PacketInterval, t.Address).Output()
	if err != nil && conf.Debug {
		log.Println("ping failed:", err)
	}
	//	fmt.Println(t.Address, t.Name, t.ID)
	sp := bytes.Split(r.raw, []byte("\n"))
	if len(sp) > 3 {
		if m := lr.FindSubmatch(sp[len(sp)-3]); m != nil {
			loss := string(m[2])
			pkts := string(m[1])
			if pkts != conf.Ping.PacketCount || loss != "0" {
				fmt.Println(t.Name, string(sp[len(sp)-3]))
			}
		}
		if m := rr.FindSubmatch(sp[len(sp)-2]); m != nil {
			// min := string(m[1])
			avg, err := strconv.ParseFloat(string(m[2]), 32)
			if err != nil {
				log.Println(err)
			}
			// max := string(m[3])
			dev, err := strconv.ParseFloat(string(m[4]), 32)
			if err != nil {
				log.Println(err)
			}
			if dev > (avg * 0.3) {
				fmt.Println(t.Name, string(sp[len(sp)-2]))
			}
		}
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
	for _, t := range conf.Targets {
		ping(&t, &conf)
	}
}
