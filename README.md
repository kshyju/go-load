# go-load
 
A simple load testing tool using go. 

There are tons of go load testing tools out there. This was something I built as I was learning go.


### Usage

#### Prerequisites
Make sure you have installed[ go runtime in your machine](https://golang.org/dl/)

#### Install go-load binary

Open a terminal/command prompt and run the below code which will install the go-load binary.

    go get github.com/kshyju/go-load

#### Command line options
* -c : Specifies the number of concurrent connections(goroutines) to use. You can consider this as the RPS you want for the test. Ex: `-c=50` will send 50 requests/second.
* -d : Specifies the duration of the load test in seconds. Ex: `-d=30` will run the tests for 30 seconds.
* -h : Specifies the the request headers to send in a comma separated form. Each header item can use the `key:value` format. Ex: `my-apikey:foo,cookie:uid`
* -body : Specifies the the path to the file name which has the request payload data present. go-load will read the content of this file and use that as the request body. When this option is passed, POST method will be used to send the request.
* -u : Specifies the the request URL explicitly.

#### Send 10 requests/sec to bing.com for 30 seconds
    go-load -c=10 -d=30 https://www.bing.com

If you have special characters like `&` in your URL, you need to pass the value in quotes.

    go-load -c=10 -d=30 "https://www.bing.com?how=are&you=today"

#### Sending request headers
Request headers can be passed as a comma separated string with the `-h` flag. The string should have header key and value in the `key:value` format. Ex: `User-Agent:Go-http-client/1.1` 

    go-load -c=10 -d=30 -h my-apikey:foo,cookie:uid:bar https://www.bing.com

The above example is sending 2 request headers, "my-apikey" and "cookie".

#### Sending HTTP POST request with payload from a local file

   go-load -c=10 -d=30 -body="C:\\temp\\my-payload.json" "http://your.app/which/accepts/http-post?foo=bar"