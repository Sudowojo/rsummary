package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type Hostname struct {
	Name  string
	Count int
}

var mutex = &sync.Mutex{}

func processFile(fileName string, hostnameCount map[string]int, totalCount *int32, processedFiles *int32, totalFiles int, wg *sync.WaitGroup, wgCounter *int32) {
	defer wg.Done()

	fmt.Printf("Processing: %s (%.2f%% completed)\n", fileName, float64(*processedFiles)/float64(totalFiles)*100)

	var file io.ReadCloser
	var err error
	if strings.HasSuffix(fileName, ".gz") {
		fileRaw, err := os.Open(fileName)
		if err != nil {
			log.Fatal(err)
		}
		file, err = gzip.NewReader(fileRaw)
		if err != nil {
			log.Fatal(err)
		}
		defer fileRaw.Close()
	} else {
		file, err = os.Open(fileName)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		re := regexp.MustCompile(`(?P<Month>\w+\s\d+\s\d+:\d+:\d+)\s(?P<Hostname>[\w.-]+)`)
		match := re.FindStringSubmatch(scanner.Text())
		if len(match) > 0 {
			mutex.Lock()
			hostnameCount[match[2]]++
			atomic.AddInt32(totalCount, 1)
			mutex.Unlock()
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	atomic.AddInt32(processedFiles, 1)

	if atomic.AddInt32(wgCounter, -1) == 0 {
		wg.Done()
	}
}

func main() {
	dir := flag.String("dir", "", "Directory containing the log files to process")
	threads := flag.Int("T", 5, "Number of files to process concurrently")
	help := flag.Bool("H", false, "Prints out how to use this utility")
	flag.BoolVar(help, "help", false, "Prints out how to use this utility")

	flag.Parse()

	if *help {
		fmt.Println("Usage: ./program_name --dir=<log directory path> --T=<number of threads>")
		fmt.Println("Or: ./program_name -dir=<log directory path> -T=<number of threads>")
		os.Exit(0)
	}

	if *dir == "" {
		log.Fatal("Please provide a directory path")
	}

	files, err := ioutil.ReadDir(*dir)
	if err != nil {
		log.Fatal(err)
	}

	hostnameCount := make(map[string]int)
	var totalCount int32 = 0
	var processedFiles int32 = 0
	totalFiles := len(files)

	var wg sync.WaitGroup
	var wgCounter int32

	filesChan := make(chan string, totalFiles)
	for _, f := range files {
		filesChan <- *dir + "/" + f.Name()
	}
	close(filesChan)

	wg.Add(1)
	wgCounter = int32(totalFiles)

	for i := 0; i < *threads; i++ {
		go func() {
			for fileName := range filesChan {
				wg.Add(1)
				processFile(fileName, hostnameCount, &totalCount, &processedFiles, totalFiles, &wg, &wgCounter)
			}
		}()
	}

	wg.Wait()

	var hostnames []Hostname
	for hostname, count := range hostnameCount {
		hostnames = append(hostnames, Hostname{hostname, count})
	}

	sort.Slice(hostnames, func(i, j int) bool {
		return hostnames[i].Count > hostnames[j].Count
	})


	longestName := 0
	for _, h := range hostnames {
    		if len(h.Name) > longestName {
       			longestName = len(h.Name)
    		}
	}

	fmt.Println("\nSUMMARY:")
	fmt.Printf("%-*s  %10s  %12s  %s\n", longestName, "Hostname", "Count", "Percentage", "IP Address")
	fmt.Println(strings.Repeat("-", longestName+38))

	for _, h := range hostnames {
    		percentage := float64(h.Count) / float64(totalCount) * 100

    		ips, err := net.LookupIP(h.Name)
    		var ipAddress string
    		if err == nil && len(ips) > 0 {
        		ipAddress = ips[0].String()
    		} else {
        		ipAddress = "N/A"
    		}

    		fmt.Printf("%-*s  %10d  %6.2f%%     %s\n", longestName, h.Name, h.Count, percentage, ipAddress)
	}
}
