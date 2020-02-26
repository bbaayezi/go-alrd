package types

import "encoding/json"

type ScopusSearchData struct {
}

func (sd *ScopusSearchData) New() {
	sd = &ScopusSearchData{}
}

// scopus field
type ScopusResponse struct {
	Results ScopusResult `json:"search-results"`
}

type ScopusResult struct {
	// meta
	TotalPapers json.Number `json:"opensearch:totalResults,string"`
	StartIndex  json.Number `json:"opensearch:startIndex,string"`
	// entry
	Entry []ScopusEntryEntity `json:"entry"`
}

type ScopusEntryEntity struct {
	AbstractURL  string              `json:"prism:url"`
	CitedByCount json.Number         `json:"citedby-count,string"`
	Affiliations []AffiliationEntity `json:"affiliation"`
	Identifier   string              `json:"dc:identifier"`
}

type AffiliationEntity struct {
	Affilname    string `json:"affilname"`
	AffilCountry string `json:"affiliation-country"`
}
