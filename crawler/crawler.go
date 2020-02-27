package crawler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/time/rate"
	// only used for testing
)

type ContextKey string

// ContextValues defines context value
type ContextValues struct {
	ResHandlerFunc ResponseHandler
}

type HTTPQuery map[string]string

type CrawlerError struct {
	ErrorType error
	Code      int
}

func (ce CrawlerError) Error() string {
	return ce.ErrorType.Error()
}

func (hq HTTPQuery) Concat(new map[string]string) {
	if hq == nil {
		hq = map[string]string{}
	}
	for k, v := range new {
		hq[k] = v
	}
}

type ErrorServer struct {
	msg string
}

func (e *ErrorServer) Error() string {
	return e.msg
}

type ErrorReponseHandler struct {
	msg string
}

func (e *ErrorReponseHandler) Error() string {
	return e.msg
}




var (
	// ErrorServer defines server error
	errorServer          = &ErrorServer{"Server Error"}
	errorResponseHandler = &ErrorReponseHandler{"Response handler error: cannot decode response"}
	metaKey              = ContextKey("meta")
)

const (
	rateLimit = 5
)

// Crawl function crawls target urls asynchronously with rate limits
// and returns a slice;
// context must specify "meta" context key which includes response handler functions
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
		fmt.Println("---- [Crawler] Error occured while executing trySend: ", err, "\nProbabaly timeout, retrying...")
		trySend(ctx, limiter, client, request, responseChan)
	}
	// check if the return status code is 200 (OK)
	statusCode := res.StatusCode
	// if recieve 429 response, than retry
	// otherwise send the response to the response channel, including server error request
	if statusCode == http.StatusTooManyRequests {
		fmt.Println("---- [Crawler] Sending too much request, retrying")
		trySend(ctx, limiter, client, request, responseChan)
	} else {
		responseChan <- res
	}
	return
}

func listenResponse(ctx context.Context, responseChan chan *http.Response, resultChan chan interface{}) {
	// TODO: add summary variable
	for {
		select {
		case r := <-responseChan:
			// check for valid response
			if r.StatusCode == http.StatusOK {
				// get the handler func from context
				handler := ctx.Value(metaKey).(ContextValues).ResHandlerFunc
				// execute handler func on response
				result, err := handler(r.Body)
				// check for error
				if err != nil {
					// send nil to result channel
					fmt.Println("---- [Crawler] Error occured while encoding: ", err)
					resultChan <- errorResponseHandler
					// break current select
					r.Body.Close()
					break
				}
				// send result
				resultChan <- result
			} else {
				// Server error
				fmt.Println("---- [Crawler] Server error, sending error to result channel")
				resultChan <- errorServer
			}
			r.Body.Close()
		case <-time.After(time.Second):
			fmt.Println("---- [Crawler] Listening to valid HTTP response...")
		case <-ctx.Done():
			// TODO log summary
			return
		}

	}
}

func AddQueryToReq(req *http.Request, query HTTPQuery) (newURL string) {
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
