package harmony

import (
	"math/rand"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/redis"
	"golang.org/x/net/context"
)

var fields = log.Fields{}

type Action func(ctx context.Context)

type Config struct {
	Jitter   int
	Interval time.Duration
	Action   Action
	Redis    *redis.Helper
	Ctx      context.Context
}

func DaemonScraper(c *Config) {
	go func() {
		rsync := newSync(c.Redis, func(config *syncConfig) {
			config.interval = c.Interval
		})

		var cancelFunc context.CancelFunc
		var lastOffset time.Duration = -1
		newInstanceConfig := rsync.sync(context.Background())

		for {
			select {
			case instance := <-newInstanceConfig:
				fields = log.Fields{"interval": c.Interval.String(), "offset": instance.offset().Seconds(), "instances": instance.count(),
					"position": instance.position(), "instance_id": instance.id}

				if instance.offset() != lastOffset {
					log.WithFields(fields).Infoln("new offset recieved")

					// No need to cancel the previous go routine, there isn't one
					if cancelFunc != nil {
						cancelFunc()
					}

					var ctx context.Context
					ctx, cancelFunc = context.WithCancel(c.Ctx)

					go (&actionConfig{
						action:   c.Action,
						offset:   addJitter(c.Jitter, c.Interval, rand.NewSource(time.Now().UnixNano())),
						interval: c.Interval,
						ctx:      ctx,
					}).prepareForSync(ctx)
				}

				lastOffset = instance.offset()
			}
		}
	}()
}

func addJitter(jitter int, interval time.Duration, source rand.Source) time.Duration {
	if jitter == 0 {
		return interval
	}

	maxJitter := interval.Nanoseconds() + int64(float64(interval.Nanoseconds())*1/float64(jitter))
	newJitter := rand.New(source).Int63n(maxJitter)
	return time.Duration(newJitter) + interval
}

type actionConfig struct {
	action   Action
	offset   time.Duration
	interval time.Duration
	ctx      context.Context
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
