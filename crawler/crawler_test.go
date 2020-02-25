package crawler

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestCrawl(t *testing.T) {
	urls := []string{
		"http://www.mocky.io/v2/5e54414a2e000049e72db27e",
		"http://www.mocky.io/v2/5e5441652e0000d8ec2db27f",
	}
	ctx := context.WithValue(context.Background(), metaKey, ContextValues{
		resHandlerFunc: JSONDecodeHandler,
	})
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	type args struct {
		ctx    context.Context
		client *http.Client
		urls   []string
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
	}{
		// TODO: Add test cases.
		{
			"test",
			args{
				ctx:    ctx,
				client: client,
				urls:   urls,
			},
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Crawl(tt.args.ctx, tt.args.client, tt.args.urls); len(got) != tt.wantLen {
				t.Errorf("Crawl() = %v, wantLen %v", got, tt.wantLen)
			} else {
				t.Logf("Got responses: %+v", got)
			}
		})
	}
}
