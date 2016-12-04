package harmony

import (
	"time"
	"uct/common/redis"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

var fields = log.Fields{}

type Action func(ctx context.Context)

func DaemonScraper(wrapper *redis.Helper, interval time.Duration, start Action) {
	go func() {
		rsync := newSync(wrapper, func(config *config) {
			config.interval = interval
		})

		var cancelFunc context.CancelFunc
		var lastOffset time.Duration = -1
		newInstanceConfig := rsync.sync(context.Background())

		for {
			select {
			case instance := <-newInstanceConfig:
				fields = log.Fields{"offset": instance.offset().Seconds(), "instances": instance.count(),
					"position": instance.position(), "instance_id": instance.id}

				if instance.offset() != lastOffset {
					log.WithFields(fields).Infoln("new offset recieved")

					// No need to cancel the previous go routine, there isn't one
					if cancelFunc != nil {
						cancelFunc()
					}

					var ctx context.Context
					ctx, cancelFunc = context.WithCancel(context.Background())

					go (&actionConfig{
						action:start,
						offset:instance.offset(),
						interval:interval,
						ctx: ctx,
					}).prepareForSync(ctx)
				}

				lastOffset = instance.offset()
			}
		}
	}()
}

type actionConfig struct {
	action Action
	offset time.Duration
	interval time.Duration
	ctx context.Context
}

func (ac *actionConfig) prepareForSync(ctx context.Context) {
	secondsTilNextMinute := time.Duration(60-time.Now().Second()) * time.Second
	// Sleeps until the next minute + the calculated offset
	dur := secondsTilNextMinute + ac.offset
	log.Infoln("sleeping to syncronize for", dur.String())

	syncTimer := time.AfterFunc(dur, ac.run)

	<-ctx.Done()

	log.Infoln("cancelling previous ticker")

	// Stop timer if it has not stopped already
	syncTimer.Stop()
}

func (ac *actionConfig) run() {
	log.Infoln("ticker for", ac.interval.String())
	ctx, _ := context.WithCancel(ac.ctx)

	for c := time.Tick(ac.interval); ; {
		log.WithFields(fields).Infoln("sync info")
		ctx, _ := context.WithTimeout(ctx, time.Minute)
		go ac.action(ctx)

		select {
		case <-c:
			continue
		case <-ctx.Done():
			log.Infoln("new offset received, cancelling old ticker")
			return
		}
	}
}