package types

import (
	"encoding/json"
	"regexp"
	"strings"
)

// abstract field
type AbstractResponse struct {
	Results AbstractResult `json:"abstracts-retrieval-response"`
}

type AbstractResult struct {
	Coredata     AbstractCore      `json:"coredata"`
	Language     LanguageEntity    `json:"language"`
	SubjectAreas SubjectAreaEntity `json:"subject-areas"`
}

type SubjectAreaEntity struct {
	SubjectArea []Subject `json:"subject-area"` // can be an object or an array
}

type Subject struct {
	Name         string `json:"$"`
	Abbreviation string `json:"@abbrev"`
}

type LanguageEntity struct {
	Value string `json:"@xml:lang"`
}

type AbstractCore struct {
	Title        string      `json:"dc:title"`
	CitedbyCount json.Number `json:"citedby-count,string"`
	Date         string      `json:"prism:coverDate"`
	ContentType  string      `json:"prism:aggregationType"`
	Creator      Creator     `json:"dc:creator"`
	Identifier   string      `json:"dc:identifier"`
	Publisher    string      `json:"dc:publisher"`
}

func (core *AbstractCore) WashPublisherInfo() {
	re := regexp.MustCompile("[a-zA-Z0-9.!#$%&’*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\\.[a-zA-Z0-9-]+)*")
	excludeMail := re.ReplaceAll([]byte(core.Publisher), []byte(""))
	newStr := string(excludeMail)
	if strings.Contains(newStr, "Inc") {
		newStr = strings.Split(newStr, "Inc")[0] + "Inc."
	}
	if strings.Contains(newStr, "Ltd") {
		newStr = strings.Split(string(excludeMail), "Ltd")[0] + "Ltd."
	}
	// trim suffix
	core.Publisher = strings.TrimSuffix(newStr, " ")
}

type Creator struct {
	Authors []AuthorEntity `json:"author"`
}

type AuthorEntity struct {
	GivenName     string              `json:"ce:given-name"`
	SurName       string              `json:"ce:surname"`
	Initials      string              `json:"ce:initials"`
	PreferredName AuthorPreferredName `json:"preferred-name"`
	IsFirstAuthor json.Number         `json:"@seq,string"`
	Affiliation   interface{}         `json:"affiliation"` // can be an object or an array
}

type AuthorPreferredName struct {
	GivenName   string `json:"ce:given-name"`
	SurName     string `json:"ce:surname"`
	Initials    string `json:"ce:initials"`
	IndexedName string `json:"ce:indexed-name"`
}
