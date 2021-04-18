package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type ResponseItem struct {
	status  string
	latency int64
}

func main() {

	urlPtr := flag.String("u", "", "The URL to send traffic to")
	durationPtr := flag.Int("d", 1, "Duration in seconds. Default is 1.")
	rpsPtr := flag.Int("c", 3, "Number of connections to use per second. This is almost same as RPS. Default is 12")
	headersPtr := flag.String("h", "", "The request headers in comma separated form")
	bodyFileNamePtr := flag.String("body", "", "The file name which contains request body. Used for POST calls.")
	verboseLoggingPtr := flag.Bool("v", false, "Is verbose logging enabled")

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
	var headerStringCommaSeparated = *headersPtr
	var rps = *rpsPtr
	var duration = *durationPtr
	var requestBodyFileName = *bodyFileNamePtr
	var verboseLoggingEnabled = *verboseLoggingPtr

	// If we got a body payload file from user, user that for request Body.
	var bodyContentToSend []byte

	if requestBodyFileName != "" {
		content, err := ioutil.ReadFile(requestBodyFileName)
		if err != nil {
			log.Fatal(err)
		}
		bodyContentToSend = content
	}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	var mutex = &sync.Mutex{}
	responseStatusCountMap := make(map[string]int)

	start := time.Now()

	// Build the header dictionary if user has provided it in comma separated format.
	var headerMap = make(map[string]string)
	allHeaders := strings.Split(headerStringCommaSeparated, ",")
	for _, header := range allHeaders {
		headerStringNameAndValueArray := strings.Split(header, ":")
		if len(headerStringNameAndValueArray) == 2 {
			headerMap[headerStringNameAndValueArray[0]] = headerStringNameAndValueArray[1]
		}
	}

	emojis := [10]string{"🌿", "🍁", "🌞", "🌷", "🌼", "🐱", "❄️", "🌱", "🍂", "🌴"}
	s := make([]ResponseItem, 0)

	fmt.Printf("📢 Will send %d requests per seconds for %d seconds to %s \n", rps, duration, url)
	var wg sync.WaitGroup
	for secondsCounter := 1; secondsCounter <= duration; secondsCounter++ {

		for counter := 0; counter < rps; counter++ {
			wg.Add(1)
			go makeRestCallAsync(client, url, bodyContentToSend, headerMap, &wg, responseStatusCountMap, verboseLoggingEnabled, mutex, &s)
		}

		var finished = secondsCounter * rps
		var emojiCounter = secondsCounter % 10

		fmt.Printf("  %s Finished sending %d \n", emojis[emojiCounter], finished)
		// Sleep for a second
		time.Sleep(1 * time.Second)
	}
	wg.Wait()

	end := time.Now()
	elapsed := end.Sub(start)

	//sort.Slice(s, func(i, j int) bool { return s[i].latency < s[j].latency })

	for i := 0; i < len(s); i++ {
		ist := s[i]
		fmt.Printf("%s %d\n", ist.status, ist.latency)

	}

	// fmt.Println("✨ Response codes received(count)")
	// for k, v := range responseStatusCountMap {
	// 	fmt.Printf("       %s: %d\n", k, v)
	// }
	fmt.Printf("%d\n", len(s))
	fmt.Printf("🎉 Total Elapsed time %s \n", elapsed)

	//fmt.Println("By latency:", s)
}

// Makes an HTTP call to the URL passed in.
// If "bodyContentToSend" is not nil, we default the request method to POST.
func makeRestCallAsync(client *http.Client, url string, bodyContentToSend []byte, headerMap map[string]string, wg *sync.WaitGroup, responseStatusCountMap map[string]int, verboseLogging bool, mutex *sync.Mutex,
	s *[]ResponseItem) {
	start := time.Now()

	reqBody := bytes.NewBuffer(bodyContentToSend)
	var method = "GET"
	if len(bodyContentToSend) > 0 {
		method = "POST"
	}

	req, _ := http.NewRequest(method, url, reqBody)

	if len(headerMap) > 0 {
		for headerName, headerValue := range headerMap {
			req.Header.Set(headerName, headerValue)
		}
	}

	if len(bodyContentToSend) > 0 {
		req.Header.Set("content-type", "application/json")
	}

	var resp, httpCallError = client.Do(req)

	if httpCallError == nil {
		end := time.Now()
		elapsed := end.Sub(start)
		if verboseLogging {
			fmt.Printf("%s Elapsed: %s \n", resp.Status, elapsed)
		}
		f := ResponseItem{"200", elapsed.Milliseconds()}

		// Record the response status code to our dictionary so we can print the summary later.
		mutex.Lock()
		*s = append(*s, f)
		mutex.Unlock()

		wg.Done()
	} else {
		fmt.Printf("ERROR: %s\n", httpCallError)
		wg.Done()
	}
}
