package main

import (
	"context"
	"database/sql"
	"fmt"
	"go-alrd/crawler"
	"go-alrd/platform"
	"go-alrd/secret"
	sqlutil "go-alrd/sql"
	"go-alrd/util"
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
	// updateAbstract()
	checkForUpdate()
}

func updateScopusData() {
	scopusCtx, scopusCancel := context.WithCancel(context.Background())
	defer scopusCancel()
	fmt.Println("Checking for latest scopus data...")
	// config database
	db, err := gorm.Open("postgres", secret.DBString)
	if err != nil {
		// log.Fatal(err)
		fmt.Println("---- Error connecting database: ", err, ", returning")
		return
	}
	// close the db connection when exit
	defer db.Close()
	var scopusRecord sqlutil.ScopusRecord
	db.First(&scopusRecord)

	// send one request to get meta data
	peekRes, err := platform.CrawlScopusAPI(scopusCtx, 0)
	if err != nil {
		// TODO: add a better error handler
		//
		fmt.Println("---- Error peeking, returning")
		// log.Fatal(err)
		return
	}
	totalPapers, _ := peekRes.Results.TotalPapers.Int64()
	// Test
	// totalPapers = 100
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
			res, err := platform.CrawlScopusAPI(scopusCtx, recordEndIndex)
			if err != nil {
				fmt.Println("---- Error occured for current scopus search. Recording result and skip")
				// skip current iteration
				// update search result database
				statusCode := 0
				switch err.(type) {
				case *crawler.ErrorServer:
					statusCode = http.StatusInternalServerError
				case *crawler.ErrorReponseHandler:
					statusCode = -1
				default:
					statusCode = 0
				}
				newResult := sqlutil.SearchResult{
					ID: sql.NullInt64{
						Int64: int64(recordEndIndex),
						Valid: true,
					},
					Type: "SCOPUS",
					StatusCode: sql.NullInt64{
						Int64: int64(statusCode),
						Valid: true,
					},
					URL: scopusURL,
				}
				// add in database
				db.Create(&newResult)
				break
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
					ScopusID: scID,
					URL:      url,
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
		fmt.Println("---- All Scopus data is up to date, wait for next round")
	}
}

// updateAbstract will find 100 records with status code 0 from the db
// and update abtract data. if no records found, return
func updateAbstract() bool {
	abstractCtx, abstractCancel := context.WithCancel(context.Background())
	defer abstractCancel()
	fmt.Println("Checking for latest abstract data...")
	// config database
	db, err := gorm.Open("postgres", secret.DBString)
	if err != nil {
		// log.Fatal(err)
		fmt.Println("---- Error connecting database: ", err, ", returning")
		// skip this round
		return true
	}
	// close the db connection when exit
	defer db.Close()
	// test db
	freeAbstractRetrieve := []sqlutil.AbstractRetrieve{}
	db.Limit(100).Where("status_code = ?", sql.NullInt64{
		Int64: int64(0),
		Valid: true,
	}).Find(&freeAbstractRetrieve)
	if len(freeAbstractRetrieve) == 0 {
		fmt.Println("---- No retrievable abstract data. Returning")
		return true
	}
	// fmt.Println(freeAbstractRetrieve)
	// extract urls
	urls := []string{}
	for _, ar := range freeAbstractRetrieve {
		urls = append(urls, ar.URL)
	}
	abstractResponses := platform.CrawlAbstracts(abstractCtx, urls)
	// update db
	fmt.Println("---- Updating new abstract data")
	for _, res := range abstractResponses {
		authorNameArr := []string{}
		subjectAreas := []string{}
		// setup country
		countryArr := []string{}
		// setup author keyword
		authorKeywordArr := []string{}
		// check nil
		if res.Results.Affiliation != nil {
			if afil, ok := res.Results.Affiliation.([]interface{}); !ok {
				// not an array
				country := res.Results.Affiliation.(map[string]interface{})["affiliation-country"]
				countryArr = append(countryArr, country.(string))
			} else {
				// is an array, can be duplicate
				for _, af := range afil {
					country := af.(map[string]interface{})["affiliation-country"]
					// check nil
					if country != nil {
						countryArr = append(countryArr, country.(string))
					}
				}
			}
			// remove duplicate
			countryArr = util.RemoveDuplicates(countryArr)
		}
		// construct cited by count
		citedbyCount, err := res.Results.Coredata.CitedbyCount.Int64()
		if err != nil {
			fmt.Println(err)
			citedbyCount = int64(0)
		}
		// construct subject area array
		for _, subject := range res.Results.SubjectAreas.SubjectArea {
			subjectAreas = append(subjectAreas, subject.Name)
		}
		// construct author keyword array
		// check if it is an array
		if kArr, ok := res.Results.AuthKeywords.AuthorKeyword.([]interface{}); ok {
			for _, k := range kArr {
				name := k.(map[string]interface{})["$"].(string)
				authorKeywordArr = append(authorKeywordArr, name)
			}
		} else {
			// it is an object or nil
			if key, ok := res.Results.AuthKeywords.AuthorKeyword.(map[string]interface{}); ok {
				name := key["$"].(string)
				authorKeywordArr = append(authorKeywordArr, name)
			}
		}

		// construct author name array
		for _, author := range res.Results.Coredata.Creator.Authors {
			name := author.PreferredName.GivenName + " " + author.PreferredName.SurName
			authorNameArr = append(authorNameArr, name)
		}
		// wash publisher data
		res.Results.Coredata.WashPublisherInfo()
		// create new abstract data record

		newAbstractData := sqlutil.AbstractData{
			ScopusID: res.Results.Coredata.Identifier,
			Title:    res.Results.Coredata.Title,
			Author:   authorNameArr, // stores an array of string
			Date:     res.Results.Coredata.Date,
			CitedbyCount: sql.NullInt64{
				Int64: citedbyCount,
				Valid: true,
			},
			ContentType:     res.Results.Coredata.ContentType,
			SubjectArea:     subjectAreas,
			AuthorKeyword:   authorKeywordArr,
			Publisher:       res.Results.Coredata.Publisher,
			PublicationName: res.Results.Coredata.PublicationName,
			Language:        res.Results.Language.Value,
			Country:         countryArr,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		dbErr := db.Create(&newAbstractData).Error
		statusCode := 0
		if dbErr == nil {
			// update abstract retrieve
			statusCode = http.StatusOK
		} else {
			// update status code with -2 to inform db insertion failed
			fmt.Println("---- DB insertion failed. Please check record in abstract retrieve table with status code -2")
			statusCode = -2
		}

		code := sql.NullInt64{
			Int64: int64(statusCode),
			Valid: true,
		}
		// update abstract retrieve
		db.Model(&sqlutil.AbstractRetrieve{}).Where("scopus_id = ?", res.Results.Coredata.Identifier).UpdateColumn("status_code", code)
	}
	return false
}

func checkForUpdate() {
	for {
		// first check scopus data
		updateScopusData()
		// than update abstract data until there is no retrievable record
		for {
			if ok := updateAbstract(); !ok {
				// wait for a sec to start next round of updating
				fmt.Println("Wait for 10 seconds to start next abstract retrieve round :)")
				time.Sleep(10 * time.Second)
			} else {
				break
			}
		}
		// check for updates every 12 hours
		fmt.Println("Wait for 12 hours...")
		time.Sleep(12 * time.Hour)
	}
}
