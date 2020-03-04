package platform

import (
	"context"
	"go-alrd/util"
	"testing"
)

func TestCrawlAbstracts(t *testing.T) {
	type args struct {
		ctx  context.Context
		urls []string
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
				context.Background(),
				[]string{
					"https://api.elsevier.com/content/abstract/scopus_id/44349176059",
					"https://api.elsevier.com/content/abstract/scopus_id/55749113518",
					"https://api.elsevier.com/content/abstract/scopus_id/85076892999",
				},
			},
			3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CrawlAbstracts(tt.args.ctx, tt.args.urls)
			// write to file
			util.WriteToJSONFile(got, "testAbstractData")
			if len(got) != tt.wantLen {
				t.Errorf("CrawlAbstracts() = %v, wantLen %v", got, tt.wantLen)
			}
		})
	}
}
