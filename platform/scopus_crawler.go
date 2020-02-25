package platform

import (
	"context"
	c "go-alrd/crawler"
	"net/http"
	"time"
)

func CrawlScopusAPI() {
	// TODO: setup database
	// setup client
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: NewAuthTransport(nil),
	}
	// TODO: Get urls from the database
	urls := []string{}
	c.Crawl(context.Background(), client, urls)
}
