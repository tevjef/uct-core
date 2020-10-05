package uct_backend

import (
	"context"
	"net/http"

	uctfirestore "github.com/tevjef/uct-backend/common/firestore"
	"github.com/tevjef/uct-backend/ein"
	"github.com/tevjef/uct-backend/hermes"
	"github.com/tevjef/uct-backend/scrapers/rutgers"
	"github.com/tevjef/uct-backend/spike"
)

func RutgersScraper(w http.ResponseWriter, r *http.Request) {
	rutgers.RutgersScraper(w, r)
}

func Ein(w http.ResponseWriter, r *http.Request) {
	ein.Ein(w, r)
}

func Spike(w http.ResponseWriter, r *http.Request) {
	spike.Spike(w, r)
}

func Hermes(context context.Context, event uctfirestore.FirestoreEvent) error {
	return hermes.Hermes(context, event)
}
