package sqlutil

import (
	"database/sql"
	"time"
)

// SearchResult represents search result structure
type SearchResult struct {
	ID         sql.NullInt64 `gorm:"AUTO_INCREMENT;PRIMARY_KEY;column:id"`
	Type       string        `gorm:"column:type"`
	StatusCode sql.NullInt64 `gorm:"column:status_code"`
	URL        string        `gorm:"column:url"`
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
	Scopus_ID  string        `gorm:"column:scopus_id"`
	URL        string        `gorm:"column:url"`
	StatusCode sql.NullInt64 `gorm:"column:status_code;default:0"`
}

func (AbstractRetrieve) TableName() string {
	return "t_abstract_retrieve"
}
