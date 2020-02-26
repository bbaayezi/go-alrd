package main

import (
	"context"
	"database/sql"
	"fmt"
	"go-alrd/platform"
	"go-alrd/secret"
	sqlutil "go-alrd/sql"
	"log"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

const (
	itemPerPage = 25
)

var (
	scopusURL = secret.BaseURL + secret.APIScopus
)

func main() {
	run()
}

func run() {
	mainCtx, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()
	fmt.Println("Starting ALRD app...")
	// config database
	db, err := gorm.Open("postgres", secret.DBString)
	// close the db connection when exit
	defer db.Close()
	if err != nil {
		// TODO: add a better error handler
		log.Fatal(err)
	}
	var scopusRecord sqlutil.ScopusRecord
	db.First(&scopusRecord)

	// send one request to get meta data
	peekRes, err := platform.CrawlScopusAPI(mainCtx, 0)
	if err != nil {
		// TODO: add a better error handler
		log.Fatal(err)
	}
	totalPapers, _ := peekRes.Results.TotalPapers.Int64()
	// Test
	totalPapers = 100
	// recordStartIndex := int(totalPapers-scopusRecord.StartIndex.Int64)
	recordEndIndex := int(scopusRecord.EndIndex.Int64)
	// compare total papers with end index
	if totalPapers > scopusRecord.EndIndex.Int64 {
		fmt.Println("---- New papers detected. Updating database")
		var targetScopusSearchTimes int
		mod := int((totalPapers - scopusRecord.EndIndex.Int64) % itemPerPage)
		if mod == 0 {
			targetScopusSearchTimes = int((totalPapers - scopusRecord.EndIndex.Int64) / itemPerPage)
		} else if mod > 0 {
			targetScopusSearchTimes = int((totalPapers-scopusRecord.EndIndex.Int64)/itemPerPage) + 1
		}
		// update scopus search
		for i := 0; i < targetScopusSearchTimes; i++ {
			// start scopus search with end index
			res, err := platform.CrawlScopusAPI(mainCtx, recordEndIndex)
			if err != nil {
				log.Fatal(err)
			}
			// update search result database
			newResult := sqlutil.SearchResult{
				ID: sql.NullInt64{
					Int64: int64(recordEndIndex),
					Valid: true,
				},
				Type: "SCOPUS",
				StatusCode: sql.NullInt64{
					Int64: int64(http.StatusOK),
					Valid: true,
				},
				URL: scopusURL,
			}
			// add in database
			db.Create(&newResult)

			// iterate through scopus response entries
			for i, entry := range res.Results.Entry {
				// add new abstract search retrieve record
				// id stands for the current position
				id := recordEndIndex + i
				scID := entry.Identifier
				url := entry.AbstractURL
				newAbstractRetrieve := sqlutil.AbstractRetrieve{
					ID: sql.NullInt64{
						Int64: int64(id),
						Valid: true,
					},
					Scopus_ID: scID,
					URL:       url,
					StatusCode: sql.NullInt64{
						Int64: 0,
						Valid: true,
					},
				}
				// add to database
				db.Create(&newAbstractRetrieve)
			}

			// update end index
			recordEndIndex += len(res.Results.Entry)
			// update scopus record
			updateScopusRecord := sqlutil.ScopusRecord{
				EndIndex: sql.NullInt64{
					Int64: int64(recordEndIndex),
					Valid: true,
				},
				UpdatedAt: time.Now().Local(),
			}

			// update in database
			db.Model(&sqlutil.ScopusRecord{}).Updates(updateScopusRecord)
		}
		// TODO: check search result table and update abstract retrieve
	} else {
		fmt.Println("---- All data is up to date, wait for next round")
	}
}

func checkUpdate(totalPapers int64, endIndex int64) {

}
