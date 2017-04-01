package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	gcm "github.com/tevjef/go-gcm"
	"github.com/tevjef/uct-core/common/conf"
	"github.com/tevjef/uct-core/common/database"
	"github.com/tevjef/uct-core/common/model"
	"github.com/tevjef/uct-core/common/notification"
	"github.com/tevjef/uct-core/common/redis"
	"github.com/tevjef/uct-core/common/try"
	"golang.org/x/net/context"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type hermes struct {
	app      *kingpin.ApplicationModel
	config   *hermesConfig
	redis    *redis.Helper
	postgres database.Handler
	ctx      context.Context
}

type hermesConfig struct {
	service conf.Config
	dryRun  bool
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
}

func main() {
	hconf := &hermesConfig{}

	app := kingpin.New("hermes", "A server that listens to a database for events and publishes notifications to Firebase Cloud Messaging")

	app.Flag("dry-run", "enable dry-run").
		Short('d').
		Default("true").
		Envar("HERMES_ENABLE_FCM").
		BoolVar(&hconf.dryRun)

	configFile := app.Flag("config", "configuration file for the application").
		Short('c').
		Envar("HERMES_CONFIG").
		File()

	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Parse configuration file
	hconf.service = conf.OpenConfigWithName(*configFile, app.Name)

	if hconf.dryRun {
		log.Infoln("Enabling FCM in production mode")
	}

	// Open database connection
	pgDatabase, err := model.OpenPostgres(hconf.service.DatabaseConfig(app.Name))
	if err != nil {
		log.WithError(err).Fatalln("failed to open database connection")
	}

	// Start profiling
	go model.StartPprof(hconf.service.DebugSever(app.Name))

	(&hermes{
		app:      app.Model(),
		config:   hconf,
		redis:    redis.NewHelper(hconf.service, app.Name),
		postgres: database.NewHandler(app.Name, pgDatabase, queries),
	}).init()
}

func (hermes *hermes) init() {
	resultChan := hermes.waitForPop()

	for {
		select {
		case pair := <-resultChan:
			go hermes.recvNotification(pair)
		}
	}
}

func (hermes *hermes) recvNotification(pair notificationPair) {
	log.WithFields(log.Fields{"university_name": pair.n.University.TopicName,
		"notification_id": pair.n.NotificationId, "status": pair.n.Status,
		"topic": pair.n.TopicName}).Info("postgres_notification")

	defer func(start time.Time) {
		log.WithFields(log.Fields{"elapsed": time.Since(start).Seconds() * 1e3,
			"university_name": pair.n.University.TopicName,
			"name":            "send_notification"}).Infoln()
	}(time.Now())

	// Retry in case of SSL/TLS timeout errors. FCM itself should be rock solid
	err := try.Do(func(attempt int) (retry bool, err error) {
		if err = hermes.sendGcmNotification(pair); err != nil {
			return true, err
		}
		return false, nil
	})

	if err != nil {
		log.WithError(err).Errorln()
	}
}

func (hermes *hermes) waitForPop() chan notificationPair {
	c := make(chan notificationPair)
	go func() {
		for {
			if pair, err := hermes.popNotification(); err == nil {
				c <- *pair
			} else {
				log.WithError(err).Warningln()
			}
		}

	}()
	return c
}

func (hermes *hermes) popNotification() (*notificationPair, error) {
	if topic, err := hermes.redis.Client.BRPopLPush(notification.MainQueue, notification.DoneQueue, 0).Result(); err == nil {
		if b, err := hermes.redis.Client.Get(notification.MainQueueData + topic).Bytes(); err != nil {
			return nil, errors.Wrap(err, "error getting notification data")
		} else {
			uctNotification := &model.UCTNotification{}
			if err := uctNotification.Unmarshal(b); err != nil {
				return nil, err
			} else if jsonBytes, err := ffjson.Marshal(uctNotification); err != nil {
				return nil, err
			} else if _, err := hermes.redis.Client.Del(topic).Result(); err != nil {
				log.WithError(err).Warningln("failed to del topic data")
				return &notificationPair{n: uctNotification, raw: string(jsonBytes)}, nil
			} else {
				return &notificationPair{n: uctNotification, raw: string(jsonBytes)}, nil
			}
		}
	} else {
		return nil, err
	}
}

func (hermes *hermes) sendGcmNotification(pair notificationPair) (err error) {
	httpMessage := gcm.HttpMessage{
		To:               "/topics/" + pair.n.TopicName,
		Data:             map[string]interface{}{"message": pair.raw},
		ContentAvailable: true,
		Priority:         "high",
		DryRun:           hermes.config.dryRun,
	}

	var httpResponse *gcm.HttpResponse
	if httpResponse, err = gcm.SendHttp(hermes.config.service.Hermes.ApiKey, httpMessage); err != nil {
		return
	}

	log.WithFields(log.Fields{"topic": httpMessage.To, "university_name": pair.n.University.TopicName,
		"message_id": httpResponse.MessageId, "error": httpResponse.Error}).Infoln("fcm_response")
	// Print FCM errors, but don't panic
	if httpResponse.Error != "" {
		return fmt.Errorf(httpResponse.Error)
	}

	hermes.acknowledgeNotification(pair.n.NotificationId, httpResponse.MessageId)
	return
}

type notificationPair struct {
	n   *model.UCTNotification
	raw string
}

func (hermes *hermes) acknowledgeNotification(notificationId, messageId int64) int64 {
	args := map[string]interface{}{"notification_id": notificationId, "message_id": messageId}
	return hermes.postgres.Update(AckNotificationQuery, args)
}

var queries = []string{
	AckNotificationQuery,
}

const AckNotificationQuery = `UPDATE notification SET (ack_at, message_id) = (now(), :message_id) WHERE id = :notification_id RETURNING notification.id`
