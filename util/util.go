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
