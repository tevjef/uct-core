package harmony

import (
	"time"
	"uct/redis"

	"github.com/satori/go.uuid"

	log "github.com/Sirupsen/logrus"
)

func DaemonScraper(wrapper *redishelper.RedisWrapper, interval time.Duration, start func(cancel chan bool)) {
	// Start redis client
	rsync := New(wrapper, interval, uuid.NewV4().String())

	newOffsetChan := rsync.Sync(make(chan bool))

	var cancelSync chan bool
	for {
		select {
		case offset := <-newOffsetChan:
			log.WithFields(log.Fields{"offset": offset.Seconds(), "instances": rsync.Instances, "position": rsync.position}).Debugln("New offset recieved")

			// No need to cancel the previous go routine, there isn't one
			if cancelSync != nil {
				cancelSync <- true
			}
			cancelSync = make(chan bool)
			go prepareForSync(offset, start, interval, cancelSync)
		}
	}

}

func prepareForSync(offset time.Duration, start func(cancel chan bool), interval time.Duration, cancel chan bool) {
	secondsTilNextMinute := time.Duration(60-time.Now().Second()) * time.Second
	// Sleeps until the next minute + the calculated offset
	dur := secondsTilNextMinute + offset
	log.Debugln("Sleeping to syncronize for", dur.String())

	cancelTicker := make(chan bool, 1)

	syncTimer := time.AfterFunc(dur, func() {
		log.Debugln("Ticker for", interval.String())
		ticker := time.NewTicker(interval)
		scrapeOnTick(start, ticker, cancelTicker)
	})

	<-cancel

	log.Debugln("Cancelling previous ticker")
	cancelTicker <- true

	// Stop timer if it has not stopped already
	syncTimer.Stop()
}

func scrapeOnTick(start func(cancel chan bool), ticker *time.Ticker, cancel chan bool) {
	notifyCancel := make(chan bool, 1)
tickLoop:
	for {
		select {
		case <-ticker.C:
			go start(notifyCancel)
		case <-cancel:
			log.Debugln("New offset received, cancelling old ticker")
			// Clean up then break
			ticker.Stop()
			notifyCancel <- true
			close(cancel)
			break tickLoop
		}
	}
}
