package crawer

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	ServerError = errors.New("Server error")
)

func trySend(ctx context.Context, limiter *rate.Limiter, client http.Client, request *http.Request, responseChan chan *http.Response) (err error) {
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
		err = ServerError
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

func listenDecodedResponse(ctx context.Context, resultChan chan interface{}, target []interface{}, wantLen int) {
	for {
		select {
		case r := <-resultChan:
			// append result to target
			target = append(target, r)
			// check for result length
			if len(target) == wantLen {
				// all desired result appended
				return
			}
			// TODO: logging
		case <-ctx.Done():
			return
		}
	}
}
