package main

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/tevjef/uct-backend/scrapers/rutgers"
)

func main() {
	req, _ := http.NewRequest("GET", "/", strings.NewReader("{}"))
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	rutgers.RutgersScraper(rr, req)
}
