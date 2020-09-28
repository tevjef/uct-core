package uct_backend

import (
	"net/http"

	"github.com/tevjef/uct-backend/ein"
	"github.com/tevjef/uct-backend/scrapers/rutgers"
)

func RutgersScraper(w http.ResponseWriter, r *http.Request) {
	rutgers.RutgersScraper(w, r)
}

func Ein(w http.ResponseWriter, r *http.Request) {
	ein.Ein(w, r)
}
