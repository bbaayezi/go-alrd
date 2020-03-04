package sqlutil

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// SearchResult represents search result structure
type SearchResult struct {
	ID         sql.NullInt64 `gorm:"AUTO_INCREMENT;PRIMARY_KEY;column:id"`
	Type       string        `gorm:"column:type"`
	StatusCode sql.NullInt64 `gorm:"column:status_code"`
	URL        string        `gorm:"column:url"`
	CreatedAt  time.Time     `gorm:"column:created_at"`
}

// TableName specifies table name for search result
func (SearchResult) TableName() string {
	return "t_search_result"
}

type ScopusRecord struct {
	StartIndex sql.NullInt64 `gorm:"column:start_index"`
	EndIndex   sql.NullInt64 `gorm:"column:end_index"`
	CreatedAt  time.Time     `gorm:"column:created_at"`
	UpdatedAt  time.Time     `gorm:"column:updated_at"`
}

func (ScopusRecord) TableName() string {
	return "t_scopus_search_record"
}

type AbstractRetrieve struct {
	ID         sql.NullInt64 `gorm:"PRIMARY_KEY;column:id"`
	ScopusID   string        `gorm:"column:scopus_id"`
	URL        string        `gorm:"column:url"`
	StatusCode sql.NullInt64 `gorm:"column:status_code;default:0"`
}

func (AbstractRetrieve) TableName() string {
	return "t_abstract_retrieve"
}

type AbstractData struct {
	ScopusID        string         `gorm:"scopus_id;PRIMARY_KEY"`
	Title           string         `gorm:"title"`
	Author          pq.StringArray `gorm:"author,type:varchar(200)[]"`
	Date            string         `gorm:"date"`
	CitedbyCount    sql.NullInt64  `gorm:"citedby_count"`
	ContentType     string         `gorm:"content_type"`
	SubjectArea     pq.StringArray `gorm:"subject_area,type:varchar(400)[]"`
	Publisher       string         `gorm:"publisher"`
	Language        string         `gorm:"language"`
	Country         pq.StringArray `gorm:"country,type:varchar(200)[]"`
	PublicationName string         `gorm:"publication_name"`
	AuthorKeyword   pq.StringArray `gorm:"author_keyword,type:varchar(200)[]"`
	CreatedAt       time.Time      `gorm:"created_at"`
	UpdatedAt       time.Time      `gorm:"updated_at"`
}

func (AbstractData) TableName() string {
	return "t_abstract_data"
}
