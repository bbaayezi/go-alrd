package platform

import (
	"context"
	"encoding/json"
	"fmt"
	c "go-alrd/crawler"
	t "go-alrd/types"
	"io"
	"net/http"
	"time"
)

var (
	abstractQuery = c.HTTPQuery{
		// "field":      "creator,authors,coverDate,aggregationType,publisher,subject-area,language,dc:identifier,citedby-count",
		"view":       "FULL",
		"httpAccept": "application/json",
	}

	abstractDecodeHandler c.ResponseHandler = func(r io.Reader) (interface{}, error) {
		result := t.AbstractResponse{}
		err := json.NewDecoder(r).Decode(&result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
)

func CrawlAbstracts(ctx context.Context, urls []string) []t.AbstractResponse {
	var abstractResponses []t.AbstractResponse
	if len(urls) == 0 {
		return abstractResponses
	}
	// setup client
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: NewAuthTransport(nil),
	}
	// setup context
	abstractCtx := context.WithValue(ctx, metaKey, c.ContextValues{abstractDecodeHandler})
	// construct urls
	parsedURLs := []string{}
	for _, url := range urls {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println(err)
			break
		}
		parsedURLs = append(parsedURLs, c.AddQueryToReq(req, abstractQuery))
	}
	responses := c.Crawl(abstractCtx, client, parsedURLs)
	for _, res := range responses {
		if parsedRes, ok := res.(t.AbstractResponse); ok {
			abstractResponses = append(abstractResponses, parsedRes)
		} else {
			switch res.(type) {
			case c.ErrorReponseHandler:
				fmt.Println("---- Error decoding response, skipping")
			case c.ErrorServer:
				fmt.Println("---- Server Error")
			default:
				fmt.Println("---- Unknown Error, just return nil error")
			}
			break
		}
	}
	// test
	// util.WriteToJSONFile(abstractResponses, "abstractData")
	return abstractResponses
	// return abstractResponses, nil
}
