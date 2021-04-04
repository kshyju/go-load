package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {

	urlPtr := flag.String("u", "", "The URL to test")
	requestCountPtr := flag.Int("rc", 1, "Total number of requests to send")
	headersPtr := flag.String("h", "", "The request headers in comma separated form")

	flag.Parse()

	var url = *urlPtr
	if url == "" {
		var tailArgs = flag.Args()
		if len(tailArgs) > 0 {
			url = tailArgs[0]
		} else {
			fmt.Println("❗ Please provide the URL. Ex:goload \"https://www.bing.com\"")
			os.Exit(3)
		}
	}
	var headerString = *headersPtr
	var maxRequestCount = *requestCountPtr

	var wg sync.WaitGroup

	fmt.Printf("📢 Will send %d requests to %s \n", maxRequestCount, url)

	client := &http.Client{}
	responseStatusCountMap := make(map[string]int)

	start := time.Now()

	// do a warmup call
	wg.Add(1)
	makeRestCallAsync(client, url, headerString, &wg, responseStatusCountMap)

	for counter := 1; counter < maxRequestCount; counter++ {
		wg.Add(1)
		go makeRestCallAsync(client, url, headerString, &wg, responseStatusCountMap)
	}

	wg.Wait()
	end := time.Now()
	elapsed := end.Sub(start)

	fmt.Println("✨ Response codes received(count)")
	for k, v := range responseStatusCountMap {
		fmt.Printf("       %s: %d\n", k, v)
	}
	fmt.Printf("🎉 Total Elapsed time %s \n", elapsed)
}

func makeRestCallAsync(client *http.Client, url string, headersCommaSeparated string, wg *sync.WaitGroup, responseStatusCountMap map[string]int) {
	start := time.Now()

	req, _ := http.NewRequest("GET", url, nil)
	allHeaders := strings.Split(headersCommaSeparated, ",")

	for _, header := range allHeaders {
		headerStringNameAndValueArray := strings.Split(header, ":")
		if len(headerStringNameAndValueArray) == 2 {
			req.Header.Set(headerStringNameAndValueArray[0], headerStringNameAndValueArray[1])
		}
	}

	var resp, httpCallError = client.Do(req)
	if httpCallError == nil {
		end := time.Now()
		elapsed := end.Sub(start)
		fmt.Printf("%s %s Elapsed: %s \n", url, resp.Status, elapsed)

		val, keyPresent := responseStatusCountMap[resp.Status]

		if keyPresent {
			responseStatusCountMap[resp.Status] = val + 1
		} else {
			responseStatusCountMap[resp.Status] = 1
		}

		wg.Done()
	} else {
		log.Fatal(httpCallError)
	}
}
