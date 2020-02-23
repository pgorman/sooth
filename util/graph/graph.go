package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
)

func main() {
	reHour := regexp.MustCompile(`(?i)^(?:[a-z0-9\-\.]+)\s+(Packet Loss|Latency|Jitter)+,*\s+[a-z]{3}\s+\d\d\s+(\d\d):.+`)
	countsByHour := make(map[string]int)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() { // Scan() return false when it hits EOF or an error.
		line := scanner.Text()
		matches := reHour.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		//fmt.Println(matches[1], matches[2])
		countsByHour[matches[2]]++
	}
	if err := scanner.Err(); err != nil { // Scan() returns nil if the error was io.EOF.
		log.Fatal(err)
	}

	keys := make([]string, 0, len(countsByHour))
	maxProblems := 0
	for k, v := range countsByHour {
		keys = append(keys, k)
		if v > maxProblems {
			maxProblems = v
		}
	}
	sort.Strings(keys)

	for _, v := range keys {
		p := int((float64(countsByHour[v]) / float64(maxProblems)) * 100.0)
		fmt.Printf("%s ", v)
		for i := 0; i < p; i++ {
			fmt.Printf("*")
		}
		fmt.Println()
	}
}
