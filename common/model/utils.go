package model

import (
	"fmt"
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"

	"github.com/pkg/errors"
)

func CheckError(err error) {
	if err != nil {
		log.Fatalf("%+v\n", errors.Wrap(err, ""))
	}
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.WithFields(log.Fields{"elapsed": elapsed, "name": name}).Debug("")
}

func TimeTrackWithLog(start time.Time, logger *log.Logger, name string) {
	elapsed := time.Since(start)
	logger.WithFields(log.Fields{"elapsed": elapsed.Seconds() * 1e3, "name": name}).Info()
}

func StartPprof(host *net.TCPAddr) {
	log.Info("starting debug server on...", (*host).String())
	log.Info(http.ListenAndServe((*host).String(), nil))
}

func InitDB(connection string) (database *sqlx.DB, err error) {
	database, err = sqlx.Open("postgres", connection)
	if err != nil {
		err = fmt.Errorf("failed to open postgres databse connection. %s", err)
	}
	return
}
