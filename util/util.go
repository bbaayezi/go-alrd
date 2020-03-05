package util

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
)

// JSONUnescapedMarshal did not escape certain unicode
func JSONUnescapedMarshal(v interface{}, prefix string, indent string) ([]byte, error) {
	wBuffer := &bytes.Buffer{}
	enc := json.NewEncoder(wBuffer)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	err := enc.Encode(v)
	return wBuffer.Bytes(), err
}

func WriteToJSONFile(data interface{}, name string) {
	output, _ := JSONUnescapedMarshal(data, "", "  ")
	ioutil.WriteFile(name+".json", output, 0644)
}

func RemoveDuplicates(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key, _ := range encountered {
		result = append(result, key)
	}
	return result
}

func StringInSlice(str string, slice []string) bool {
	for _, b := range slice {
		if b == str {
			return true
		}
	}
	return false
}
