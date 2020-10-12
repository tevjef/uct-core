package main

import (
	"net/http"
	"net/http/httptest"

	"github.com/tevjef/uct-backend/ein"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	inputFile = kingpin.Arg("file", "file to input").Required().File()
)

func main() {
	kingpin.Parse()

	req, _ := http.NewRequest("POST", "http://ein.cli?backfill=true", *inputFile)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	ein.Ein(rr, req)
}
