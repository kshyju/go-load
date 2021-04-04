# go-load
 
A simple load testing tool using go. 

There are tons of go load testing tools out there. This was something I built as I was learning go.

### Usage exmaple


### Send 100 requests bing.com
    go-load -rc=100 https://www.bing.com

If you have special characters like `&` in your URL, you need to pass the value in quotes.

    go-load -rc 100  "https://www.bing.com?how=are&you=today"

### Send 100 requests bing.com with request headers
Request headers can be passed as a comma separated string with the `-h` flag. The string should have header key and value in the `key:value` format. Ex: `User-Agent:Go-http-client/1.1` 

    go-load -h my-apikey:foo,cookie:uid:bar https://www.bing.com

The above example is sending 2 request headers, "my-apikey" and "cookie".

### Command line options
* -rc : Specifies the request count. Ex: `-rc=1000` will issue 1000 requests
* -h : Specifies the the request headers to send in a comma separated form. Each header item can use the `key:value` format. Ex: `my-apikey:foo,cookie:uid`
* -u : Specifies the the request URL explicitly.