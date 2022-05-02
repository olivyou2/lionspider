package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func startsWith(src string, target string) bool {
	if len(src) < len(target) {
		return false
	}
	if src[0:len(target)] == target {
		return true
	} else {
		return false
	}
}

func endsWith(src string, target string) bool {
	if len(src) < len(target) {
		return false
	}
	if src[len(src)-len(target):] == target {
		return true
	} else {
		return false
	}
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func checkHtml(src string) bool {
	headResp, headErr := http.Head(src)

	if headErr != nil {
		return false
	}

	if len(headResp.Header["Content-Type"]) == 0 {

		return false
	}

	isHtml := startsWith(headResp.Header["Content-Type"][0], "text/html")
	return isHtml
}
