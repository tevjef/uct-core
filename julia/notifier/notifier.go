package notifier

import (
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type pgNotifier struct {
	l  *pq.Listener
	ch chan string
}

func NewNotifier(listener *pq.Listener) Notifier {
	pg := &pgNotifier{l: listener, ch: make(chan string)}
	go func() {
		for n := range pg.l.NotificationChannel() {
			pg.ch <- n.Extra
		}
	}()

	return pg
}

func (pg *pgNotifier) Notify() <-chan string {
	return pg.ch
}

func (pg *pgNotifier) Ping() {
	if err := pg.l.Ping(); err != nil {
		log.WithError(err).Fatalln("failed to ping server")
	}
}

type Notifier interface {
	Notify() <-chan string
	Ping()
}
