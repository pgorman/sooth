// graph prints an ASCII graph of problems per hour in a Sooth log.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
)

func main() {
	var maxProblems int
	flag.IntVar(&maxProblems, "x", 0, "set graph upper bound for problems per hour")
	flag.Parse()
	reHour := regexp.MustCompile(`(?i)^(?:[a-z0-9\-\.]+)\s+(?:Packet Loss|Latency|Jitter)+,*\s+[a-z]{3}\s+\d\d\s+(\d\d):.+`)
	countsByHour := make(map[string]int)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		matches := reHour.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		countsByHour[matches[1]]++
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	if maxProblems == 0 {
		for _, v := range countsByHour {
			if v > maxProblems {
				maxProblems = v
			}
		}
	}

	for i := 0; i < 24; i++ {
		var k string
		if i >= 10 {
			k = strconv.Itoa(i)
		} else {
			k = "0" + strconv.Itoa(i)
		}
		if _, ok := countsByHour[k]; ok {
			p := int((float64(countsByHour[k]) / float64(maxProblems)) * 100.0)
			fmt.Printf("%2s %6d ", k, countsByHour[k])
			for i := 0; i < p; i++ {
				fmt.Printf("*")
			}
			fmt.Println()
		} else {
			fmt.Printf("%2d %6d\n", i, 0)
		}
	}
}
