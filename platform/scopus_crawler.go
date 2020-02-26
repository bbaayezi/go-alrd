package platform

import (
	"context"
	"encoding/json"
	"errors"
	c "go-alrd/crawler"
	"go-alrd/secret"
	t "go-alrd/types"
	"io"
	"net/http"
	"strconv"
	"time"
)

var (
	metaKey = c.ContextKey("meta")

	baseQuery = c.HTTPQuery{
		"query":      secret.SearchString,
		"field":      "affiliation,citedby-count,dc:identifier",
		"httpAccpet": "application/json",
	}

	scopusDecodeHandler c.ResponseHandler = func(r io.Reader) (interface{}, error) {
		result := t.ScopusResponse{}
		err := json.NewDecoder(r).Decode(&result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	ErrorDecode = errors.New("Recieving nil decoding result, please check the log")
)

// CrawlScopusAPI will crawl scopus api synchronously with 25 results
func CrawlScopusAPI(ctx context.Context, startIndex int) (t.ScopusResponse, error) {
	result := t.ScopusResponse{}
	// setup client
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: NewAuthTransport(nil),
	}
	defer client.CloseIdleConnections()
	// set up context with json decode handler
	scopusCtx := context.WithValue(ctx, metaKey, c.ContextValues{scopusDecodeHandler})
	// send one simple request to get total number of papers
	url := secret.BaseURL + secret.APIScopus
	baseReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result, err
	}
	q := map[string]string{
		"start": strconv.Itoa(startIndex),
	}
	baseQuery.Concat(q)
	firstURL := c.AddQueryToReq(baseReq, baseQuery)
	// get the result
	responses := c.Crawl(scopusCtx, client, []string{firstURL})
	// check nil
	if responses[0] != nil {
		if encodedResult, ok := responses[0].(t.ScopusResponse); ok {
			result = encodedResult
		}
	} else {
		return result, ErrorDecode
	}
	return result, nil
	// test: write to file
	// firstScopusRes := responses[0].(t.ScopusResponse)
	// totalPapers := firstScopusRes.Results.TotalPapers
	// c.WriteToFile(firstScopusRes, "firstResponse")
}
