package main

import (
	_ "net/http/pprof"
	"os"
	"time"
	"uct/common/conf"
	"uct/common/model"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"uct/julia/notifier"
	"uct/notification"
	"uct/redis"

	log "github.com/Sirupsen/logrus"
	"github.com/lib/pq"
	"github.com/pquerna/ffjson/ffjson"
)

var (
	app        = kingpin.New("julia", "An application that queue messages from the database")
	configFile = app.Flag("config", "configuration file for the application").Short('c').File()
	config     = conf.Config{}
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	log.SetLevel(log.InfoLevel)
	// Parse configuration file
	config = conf.OpenConfig(*configFile)
	config.AppName = app.Name

	redisWrapper := redis.NewHelper(config, app.Name)

	// Start profiling
	go model.StartPprof(config.DebugSever(app.Name))

	// Open connection to postgresql
	listener := pq.NewListener(config.DatabaseConfig(app.Name), 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.WithError(err).Fatalln("failure in listener")
		}
	})

	// Listen on channel
	if err := listener.Listen("status_events"); err != nil {
		panic(err)
	}

	postgresNotify := notifier.NewNotifier(listener)

	log.Infoln("Start monitoring PostgreSQL...")

	var process = &Process{
		in:  make(chan model.UCTNotification),
		out: make(chan model.UCTNotification),
	}

	go process.Run(func(uctNotification model.UCTNotification) {
		log.WithFields(log.Fields{"topic": uctNotification.TopicName}).Infoln("queueing")
		if b, err := uctNotification.Marshal(); err != nil {
			log.WithError(err).Fatalln("failed to marshall notification")
		} else if _, err := redisWrapper.Client.Set(notification.MainQueueData+uctNotification.TopicName, b, 0).Result(); err != nil {
			log.WithError(err).Warningln("failed to set notification data")
		} else if redisWrapper.RPush(notification.MainQueue, uctNotification.TopicName); err != nil {
			log.WithError(err).Warningln("failed to push notification unto queue")
		}
	})

	for {
		waitForNotification(postgresNotify, process.Recv)
	}
}

func waitForNotification(l notifier.Notifier, onNotify func(notification *model.UCTNotification)) {
	for {
		select {
		case message, ok := <-l.Notify():
			if !ok {
				return
			}
			if message == "" {
				continue
			}

			go func() {
				var uctNotification model.UCTNotification
				if err := ffjson.UnmarshalFast([]byte(message), &uctNotification); err != nil {
					log.WithError(err).Errorln("failed to unmarsahll notification")
					return
				}

				onNotify(&uctNotification)
			}()

			// Received no notification from the database for 60 seconds.
		case <-time.After(1 * time.Minute):
			go l.Ping()
		}
	}
}
