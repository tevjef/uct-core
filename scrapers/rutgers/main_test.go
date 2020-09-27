package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRutgersScraper(t *testing.T) {
	req := httptest.NewRequest("GET", "/", strings.NewReader("{}"))
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	RutgersScraper(rr, req)

	if got := rr.Body.String(); got != "Complete" {
		t.Errorf("Test failed")
	}
}
