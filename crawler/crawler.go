package crawer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type contextKey string

// values struct defines context value
type values struct {
	resHandlerFunc responseHandler
}

type responseHandler func(io.Reader) (interface{}, error)

var (
	// ErrorServer defines server error
	ErrorServer = errors.New("Server error")
)

const (
	rateLimit = 5
)

// Crawl function crawls target urls asynchronously with rate limits
// and returns a slice
func Crawl(ctx context.Context, urls []string) []interface{} {
	// init http client
	// TODO: review roundtrip and setup default headers and query params
	client := &http.Client{
		// setup a 10 seconds timeout
		Timeout: 10 * time.Second,
	}
	// async send
	// init contexts with cancel
	// context value type alredy set
	crawlerCtx, crawlerCancel := context.WithCancel(ctx)
	// init rate limiter
	limiter := rate.NewLimiter(rate.Limit(rateLimit), rateLimit)
	// run response listener at background
	responseChan := make(chan *http.Response)
	resultChan := make(chan interface{})
	go listenResponse(crawlerCtx, responseChan, resultChan)
	for _, url := range urls {
		// send async requests
		// macke requests
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		// try to get valid response
		go trySend(crawlerCtx, limiter, client, req, responseChan)
	}
	resultSlice := []interface{}{}
	// listen to results
	for {
		select {
		case r := <-resultChan:
			resultSlice = append(resultSlice, r)
			// check for desired length
			if len(resultSlice) == len(urls) {
				// cancel crawler context to stop response listener
				crawlerCancel()
				return resultSlice
			}
		// also consider canceled context
		case <-ctx.Done():
			crawlerCancel()
			return resultSlice
		}
	}
}

func trySend(ctx context.Context, limiter *rate.Limiter, client *http.Client, request *http.Request, responseChan chan *http.Response) (err error) {
	// send the request using client
	// install rate limiter
	limiter.Wait(ctx)
	res, err := client.Do(request)
	// check error
	if err != nil {
		return
	}
	// check if the return status code is 200 (OK)
	statusCode := res.StatusCode
	if statusCode == 200 {
		// send the valid response to response channel
		responseChan <- res
	} else if statusCode == 429 {
		// sending too much requests, retry
		trySend(ctx, limiter, client, request, responseChan)
	} else {
		// server error
		// TODO: add this failed record to database
		err = ErrorServer
	}
	return
}

func listenResponse(ctx context.Context, responseChan chan *http.Response, resultChan chan interface{}) {
	// TODO: add summary variable
	for {
		select {
		case r := <-responseChan:
			// get the handler func from context
			// TODO: change contextKey
			handler := ctx.Value(contextKey("key")).(values).resHandlerFunc
			// execute handler func on response
			result, err := handler(r.Body)
			// check for error
			if err != nil {
				// send nil to result channel
				resultChan <- nil
				// TODO: add logging
				// break current select
				r.Body.Close()
				break
			}
			// send result
			resultChan <- result
			r.Body.Close()
		case <-time.After(time.Second):
			fmt.Println("---- Listening to valid HTTP response...")
		case <-ctx.Done():
			// TODO log summary
			return
		}

	}
}

