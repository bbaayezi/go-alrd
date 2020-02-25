package sqlutil

import "go-alrd/secret"

var (
	QueryAllSearchResult = "SELECT * FROM " + secret.DBInfo.Scheme + "." + secret.DBInfo.Tables["searchResult"]
)
