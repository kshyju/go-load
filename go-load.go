package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type ResponseItem struct {
	status  string
	latency int64
}

type RunSummary struct {
	totalRequests                 int
	latencyNinetyNinePercentile   int64
	latencyNinetyFifthPercentile  int64
	latencySeventyFifthPercentile int64
	latencyFiftyPercentile        int64
	latencyForSlowestRequest      int64
	latencyForFastestRequest      int64
	responseStatusCountMap        map[string]int
	latencyAverage                int64
}

// DefaultDuration The default duration of the runs, in seconds.
const DefaultDuration = 1

// DefaultRPS The default number of requests to issue per second.
const DefaultRPS = 1

func main() {

	urlPtr := flag.String("u", "", "The URL to send traffic to")
	durationPtr := flag.Int("d", DefaultDuration, "Duration in seconds. Default is 1.")
	rpsPtr := flag.Int("c", DefaultRPS, "Number of connections to use per second. This is almost same as RPS. Default is 12")
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
			fmt.Println("❗ Please provide the URL. Ex:go-load \"https://www.bing.com\"")
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

	httpTransport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: httpTransport}
	var mutex = &sync.Mutex{}

	start := time.Now()

	// Build the header dictionary if user has provided it in comma separated format.
	headerMap := buildHeaderDictionary(headerStringCommaSeparated)

	emojis := [10]string{"🌿", "🍁", "🌞", "🌷", "🌼", "🐱", "❄️", "🌱", "🍂", "🌴"}
	httpCallResponseItems := make([]ResponseItem, 0)

	fmt.Printf("📢 Will send %d requests per seconds for %d seconds to %s \n", rps, duration, url)
	var wg sync.WaitGroup
	for secondsCounter := 1; secondsCounter <= duration; secondsCounter++ {

		for counter := 0; counter < rps; counter++ {
			wg.Add(1)
			go makeRestCallAsync(client, url, bodyContentToSend, headerMap, &wg, verboseLoggingEnabled, mutex, &httpCallResponseItems)
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

	var summary = getRunSummary(httpCallResponseItems)

	fmt.Println("======================")
	fmt.Println("✨ RUN SUMMARY ✨")
	fmt.Printf("Total requests: %d\n", summary.totalRequests)
	fmt.Printf("Total Elapsed time %s \n", elapsed)
	fmt.Println("✨ Response codes received(count)")
	for k, v := range summary.responseStatusCountMap {
		fmt.Printf("       %s: %d\n", k, v)
	}
	fmt.Println("Latencies observed in milli seconds")
	fmt.Printf("   Average: %d\n", summary.latencyNinetyNinePercentile)
	fmt.Printf("   99th percentile: %d\n", summary.latencyNinetyNinePercentile)
	fmt.Printf("   95th percentile: %d\n", summary.latencyNinetyFifthPercentile)
	fmt.Printf("   75th percentile: %d\n", summary.latencySeventyFifthPercentile)
	fmt.Printf("   50th percentile: %d\n", summary.latencyFiftyPercentile)
	fmt.Printf("🐌 Slowest request: %d\n", summary.latencyForSlowestRequest)
	fmt.Printf("🚀 Fastest request: %d\n", summary.latencyForFastestRequest)
	fmt.Println("======================")
}

//buildHeaderDictionary Builds a map for request headers to be used.
func buildHeaderDictionary(headerStringCommaSeparated string) map[string]string {
	var headerMap = make(map[string]string)
	allHeaders := strings.Split(headerStringCommaSeparated, ",")
	for _, header := range allHeaders {
		headerStringNameAndValueArray := strings.Split(header, ":")
		if len(headerStringNameAndValueArray) == 2 {
			headerMap[headerStringNameAndValueArray[0]] = headerStringNameAndValueArray[1]
		}
	}
	return headerMap
}

//getRunSummary Gets the run summary.
func getRunSummary(allResponses []ResponseItem) RunSummary {
	var runSummary RunSummary

	allResponsesSliceLength := len(allResponses)
	runSummary.totalRequests = allResponsesSliceLength

	// sort so we can pick percentile of latency
	sort.Slice(allResponses, func(i, j int) bool { return allResponses[i].latency < allResponses[j].latency })

	runSummary.latencyForFastestRequest = allResponses[0].latency
	runSummary.latencyForSlowestRequest = allResponses[allResponsesSliceLength-1].latency

	// get percentile latencies
	runSummary.latencyNinetyNinePercentile = getPercentileLatency(allResponses, 99)
	runSummary.latencyNinetyFifthPercentile = getPercentileLatency(allResponses, 95)
	runSummary.latencySeventyFifthPercentile = getPercentileLatency(allResponses, 75)
	runSummary.latencyFiftyPercentile = getPercentileLatency(allResponses, 50)

	// get response status code count
	//Creates a map for response status code and count.
	var responseStatusCountMap = make(map[string]int)
	for i := 0; i < allResponsesSliceLength; i++ {
		response := allResponses[i]
		statusCount, keyPresentForThisStatusCode := responseStatusCountMap[response.status]
		if keyPresentForThisStatusCode {
			responseStatusCountMap[response.status] = statusCount + 1
		} else {
			responseStatusCountMap[response.status] = 1
		}
	}
	// Find average latency
	var latencySum int64 = 0
	for _, item := range allResponses {
		latencySum = latencySum + item.latency
	}
	var averageLatency int64 = latencySum / int64(len(allResponses))

	runSummary.responseStatusCountMap = responseStatusCountMap
	runSummary.latencyAverage = averageLatency

	return runSummary
}

// gets the percentile latency from the sorted (by latency slice)
func getPercentileLatency(sortedLatencies []ResponseItem, percentileAskedFor int) int64 {
	sortedLatencyArrayLength := len(sortedLatencies)
	percentileItemIndex := (sortedLatencyArrayLength * percentileAskedFor) / 100
	// if there are more than one items in result, take the previous item as the percentile value item index
	// If there are 10 items and 75th percentile is requested, 10*75 / 100 = 7.5. We will pick 6th item
	if sortedLatencyArrayLength > 1 {
		percentileItemIndex = percentileItemIndex - 1
	}
	return sortedLatencies[percentileItemIndex].latency
}

// Makes an HTTP call to the URL passed in.
// If "bodyContentToSend" is not nil, we default the request method to POST.
func makeRestCallAsync(client *http.Client, url string, bodyContentToSend []byte, headerMap map[string]string, wg *sync.WaitGroup, verboseLogging bool, mutex *sync.Mutex,
	responseItems *[]ResponseItem) {

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
	mutex.Lock()
	start := time.Now()
	var resp, httpCallError = client.Do(req)
	end := time.Now()
	mutex.Unlock()
	elapsed := end.Sub(start)

	if httpCallError == nil {

		if verboseLogging {
			fmt.Printf("%s Elapsed: %s \n", resp.Status, elapsed)
		}
		responseStatusLatencyItem := ResponseItem{resp.Status, elapsed.Milliseconds()}

		// Record the response status code to our dictionary so we can print the summary later.
		//mutex.Lock()
		*responseItems = append(*responseItems, responseStatusLatencyItem)
		//mutex.Unlock()

		wg.Done()
	} else {
		fmt.Printf("ERROR: %s\n", httpCallError)
		wg.Done()
	}
}
