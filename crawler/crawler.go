package crawler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type ContextKey string

// ContextValues defines context value
type ContextValues struct {
	resHandlerFunc responseHandler
}

type HTTPQuery map[string]string

var (
	// ErrorServer defines server error
	ErrorServer = errors.New("Server error")
	metaKey     = ContextKey("meta")
)

const (
	rateLimit = 5
)

// Crawl function crawls target urls asynchronously with rate limits
// and returns a slice
func Crawl(ctx context.Context, client *http.Client, urls []string) []interface{} {
	// init http client
	// TODO: review roundtrip and setup default headers and query params

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
			// TODO: add a better error handler
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
	if statusCode == http.StatusOK {
		// send the valid response to response channel
		responseChan <- res
	} else if statusCode == http.StatusTooManyRequests {
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
			// NOTE: changed to metaKey
			handler := ctx.Value(metaKey).(ContextValues).resHandlerFunc
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

func addQueryToReq(req *http.Request, query HTTPQuery) (newURL string) {
	// get query
	q := req.URL.Query()
	// add query
	// iterate through query map
	for k, v := range query {
		q.Set(k, v)
	}
	// encode
	req.URL.RawQuery = q.Encode()
	// return thr url
	newURL = req.URL.String()
	return
}
