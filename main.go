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

	// send one request to get meta data
	peekRes, err := platform.CrawlScopusAPI(scopusCtx, 0)
	if err != nil {
		// TODO: add a better error handler
		//
		fmt.Println("---- Error peeking, returning")
		// log.Fatal(err)
		return
	}

	// get cached count
	var cachedCount int64
	db.Table("t_abstract_retrieve").Count(&cachedCount)
	fmt.Println("Cached Count: ", cachedCount)
	// compare cachedCount with retrieved count
	targetCount, _ := peekRes.Results.TotalPapers.Int64()
	if targetCount > cachedCount {
		fmt.Printf("%d new articles detected.\n", targetCount-cachedCount)
		// get all cached Scopus_ID
		var cachedIDs []string
		// Pluck() extracts scopus_id column and stores them into cachedIDs
		db.Model(&sqlutil.AbstractRetrieve{}).Pluck("scopus_id", &cachedIDs)
		// initilize abstract search wait list to store urls
		finishedSearchCount := 0
		// start scopus search from page 0
		searchIndex := 0
		for {
			// loop
			res, err := platform.CrawlScopusAPI(scopusCtx, searchIndex)
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
						Int64: int64(searchIndex),
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
				continue
			} else {
				// update search result database
				newResult := sqlutil.SearchResult{
					ID: sql.NullInt64{
						Int64: int64(searchIndex),
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
				for _, entry := range res.Results.Entry {
					id := entry.Identifier
					// check if this id is in the cached ID slice
					if !util.StringInSlice(id, cachedIDs) {
						fmt.Println("New article found! ID : ", id)
						// not in cached slice, add it to wait list
						record := sqlutil.AbstractRetrieve{
							ID: sql.NullInt64{
								Int64: cachedCount + int64(finishedSearchCount),
								Valid: true,
							},
							ScopusID: id,
							URL:      entry.AbstractURL,
							StatusCode: sql.NullInt64{
								Int64: 0,
								Valid: true,
							},
						}
						// update abstract retrieve table
						db.Create(&record)
						finishedSearchCount++
					}
				}
			}
			// check if wait list is complete
			if int64(finishedSearchCount) == targetCount-cachedCount {
				fmt.Println("abstractWaitList is complete.")
				// break the loop
				break
			} else {
				// add up search index, start next loop
				searchIndex += itemPerPage
				fmt.Println("Starting next search round with search index: ", searchIndex)
				// prevent deadlock
				if int64(searchIndex) > targetCount {
					fmt.Println("Breaking to prevent deadlock...")
					break
				}
			}
		}
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
