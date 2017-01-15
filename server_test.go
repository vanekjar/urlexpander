package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"encoding/json"
	"errors"
)

type MockExpander struct {
	result string
	err    error
}

func (ex MockExpander) ExpandUrl(shortenedUrl string) (string, error) {
	return ex.result, ex.err
}

func TestExpandHandler(t *testing.T) {

	shortenedUrl := "http://test.dev/shortened"
	expandedUrl := "http://test.dev/expanded"

	app := App{
		Expander: MockExpander{
			result: expandedUrl,
		}}

	req, err := http.NewRequest("GET", "/api/expand?url=" + shortenedUrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()

	app.expand(response, req)

	if response.Code != 200 {
		t.Fatalf("Response code %d invalid", response.Code)
	}

	var resp responseMessage
	json.Unmarshal(response.Body.Bytes(), &resp)

	expected := responseMessage{Original: shortenedUrl, Expanded: expandedUrl}
	if resp != expected {
		t.Fatalf("Invalid response returned %s", response.Body.String())
	}

	// test failing UrlExpander
	app.Expander = MockExpander{
		err: errors.New("Generic error"),
	}

	response = httptest.NewRecorder()

	app.expand(response, req)
	if response.Code != 500 {
		t.Fatalf("Response code %d invalid", response.Code)
	}

	json.Unmarshal(response.Body.Bytes(), &resp)
	if resp.Error == "" {
		t.Fatal("Expected non-empty error message")
	}
}
