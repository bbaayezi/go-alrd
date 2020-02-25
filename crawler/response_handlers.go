package crawler

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

// This file defines response handlers

type ResponseHandler func(io.Reader) (interface{}, error)

func DefaultHandler(body io.Reader) (interface{}, error) {
	return ioutil.ReadAll(body)
}

func JSONDecodeHandler(body io.Reader) (interface{}, error) {
	// decode json to target
	var target interface{}
	err := json.NewDecoder(body).Decode(&target)
	if err != nil {
		return nil, err
	}
	return target, nil
}
