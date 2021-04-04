package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {

	urlPtr := flag.String("u", "", "The URL to test")
	requestCountPtr := flag.Int("rc", 10, "Total number of requests to send")

	flag.Parse()

	var url = *urlPtr
	if (url == "") {
		var tailArgs = flag.Args()
		if (len(tailArgs) > 0) {
			url = tailArgs[0]
		} else {
			fmt.Println("Please provide the URL. Ex:goload https://www.bing.com")
			os.Exit(3)
		}
	}

	var wg sync.WaitGroup

	var maxRequestCount = *requestCountPtr

	fmt.Printf("Will send %d requests to %s \n", maxRequestCount, url)

	start := time.Now()
	// do a warmup call
	wg.Add(1)
	makeRestCallAsync(url, &wg)

	for counter := 1; counter < maxRequestCount; counter++ {
		wg.Add(1)
		go makeRestCallAsync(url, &wg)
	}

	wg.Wait()
	end := time.Now()
	elapsed := end.Sub(start)

	fmt.Printf("Total Elapsed time %s", elapsed)
}

func makeRestCallAsync(url string, wg *sync.WaitGroup) {
	start := time.Now()
	var resp, httpCallError = http.Get(url)
	if httpCallError == nil {
		end := time.Now()
		elapsed := end.Sub(start)
		fmt.Printf("%s %s Elapsed: %s \n", url, resp.Status, elapsed)

		wg.Done()
	} else {
		log.Fatal(httpCallError)
	}
}