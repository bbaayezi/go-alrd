package main

import (
	"fmt"
	"go-alrd/secret"
	sqlutil "go-alrd/sql"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("Starting ALRD app...")
	// config database
	db, err := gorm.Open("postgres", secret.DBString)
	// close the db connection when exit
	defer db.Close()
	if err != nil {
		// TODO: add a better error handler
		log.Fatal(err)
	}
	var searchResult sqlutil.SearchResult
	var count int
	// test db
	db.First(&searchResult)
	db.Model(&sqlutil.SearchResult{}).Where("type = ?", "SCOPUS").Count(&count)
	fmt.Println("Scopus search result count: ", count)

	var scopusRecord sqlutil.ScopusRecord
	db.First(&scopusRecord)
	fmt.Println(time.Now().Local())
	fmt.Println(scopusRecord)

	// check current scopus search status
	// check the start index and end index
	// if both are zero, than there is no records
	if scopusRecord.StartIndex.Int64 == scopusRecord.EndIndex.Int64 {
		//
	}
}
