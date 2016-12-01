package model

import (
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
)

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.WithFields(log.Fields{"elapsed": elapsed, "name": name}).Debug("")
}

func StartPprof(host *net.TCPAddr) {
	log.Info("starting debug server on...", (*host).String())
	log.Info(http.ListenAndServe((*host).String(), nil))
}

func InitDB(connection string) (database *sqlx.DB, err error) {
	database, err = sqlx.Open("postgres", connection)
	if err != nil {
		return nil, err
	}
	return database, err
}
