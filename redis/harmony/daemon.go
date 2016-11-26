package harmony

import (
	"time"
	"uct/redis"

	log "github.com/Sirupsen/logrus"
)

var fields = log.Fields{}

func DaemonScraper(wrapper *redis.Helper, interval time.Duration, start func(cancel chan struct{})) {
	go func() {
		rsync := newSync(wrapper, func(config *config) {
			config.interval = interval
		})

		newInstanceConfig := rsync.sync(make(chan struct{}))

		var cancelSync chan struct{}
		var lastOffset time.Duration = -1

		for {
			select {
			case instance := <-newInstanceConfig:
				fields = log.Fields{"offset": instance.offset().Seconds(), "instances": instance.count(),
					"position": instance.position(), "instance_id":instance.id}

				if instance.offset() != lastOffset {
					log.WithFields(fields).Infoln("new offset recieved")

					// No need to cancel the previous go routine, there isn't one
					if cancelSync != nil {
						cancelSync <- struct{}{}
					}
					cancelSync = make(chan struct{})
					go prepareForSync(instance, start, interval, cancelSync)
				}

				lastOffset = instance.offset()
			}
		}
	}()
}

func prepareForSync(instance instance, start func(cancel chan struct{}), interval time.Duration, cancel chan struct{}) {
	secondsTilNextMinute := time.Duration(60-time.Now().Second()) * time.Second
	// Sleeps until the next minute + the calculated offset
	dur := secondsTilNextMinute + instance.off
	log.Infoln("sleeping to syncronize for", dur.String())

	cancelTicker := make(chan struct{}, 1)

	syncTimer := time.AfterFunc(dur, func() {
		log.Infoln("ticker for", interval.String())
		ticker := time.NewTicker(interval)
		scrapeOnTick(start, ticker, cancelTicker)
	})

	<-cancel

	log.Infoln("cancelling previous ticker")
	cancelTicker <- struct{}{}

	// Stop timer if it has not stopped already
	syncTimer.Stop()
}

func scrapeOnTick(start func(cancel chan struct{}), ticker *time.Ticker, cancel chan struct{}) {
	notifyCancel := make(chan struct{}, 1)
	for {
		select {
		case <-ticker.C:
			log.WithFields(fields).Infoln("sync info")
			go start(notifyCancel)
		case <-cancel:
			log.Infoln("new offset received, cancelling old ticker")
			// Clean up then break
			ticker.Stop()
			notifyCancel <- struct{}{}
			close(cancel)
			return
		}
	}
}
