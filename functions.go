package uct_backend

import (
	"context"
	"fmt"
	"net/http"

	uctfirestore "github.com/tevjef/uct-backend/common/firestore"
	"github.com/tevjef/uct-backend/ein"
	"github.com/tevjef/uct-backend/hermes"
	"github.com/tevjef/uct-backend/scrapers/rutgers"
)

func RutgersScraper(w http.ResponseWriter, r *http.Request) {
	rutgers.RutgersScraper(w, r)
}

func Ein(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in Ein", r)
		}
	}()

	ein.Ein(w, r)
}

func Hermes(context context.Context, event uctfirestore.FirestoreEvent) error {
	return hermes.Hermes(context, event)
}
