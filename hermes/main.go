package main

import (
	_ "net/http/pprof"
	"os"
	"time"
	"uct/common/conf"
	"uct/common/model"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"strconv"
	"uct/common/try"
	"uct/notification"
	"uct/redis"

	_ "github.com/lib/pq"
	gcm "github.com/tevjef/go-gcm"

	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
)

var (
	app           = kingpin.New("hermes", "A server that listens to a database for events and publishes notifications to Google Cloud Messaging")
	dryRun        = app.Flag("dry-run", "enable dry-run").Short('d').Default("true").Bool()
	configFile    = app.Flag("config", "configuration file for the application").Short('c').File()
	config        = conf.Config{}
	database      *sqlx.DB
	preparedStmts = make(map[string]*sqlx.NamedStmt)
)

var helper *redis.Helper

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Parse configuration file
	config = conf.OpenConfig(*configFile)
	config.AppName = app.Name

	if enableFcm, _ := strconv.ParseBool(os.Getenv("ENABLE_FCM")); enableFcm {
		log.Infoln("Enabling FCM in production mode")
		*dryRun = false
	}

	// Start profiling
	go model.StartPprof(config.DebugSever(app.Name))

	helper = redis.NewHelper(config, app.Name)

	var err error
	// Open database connection
	if database, err = model.OpenPostgres(config.DatabaseConfig(app.Name)); err != nil {
		log.WithError(err).Fatalln("failed to open database connection")
	} else {
		prepareAllStmts()
	}

	resultChan := waitForPop()

	for {
		select {
		case pair := <-resultChan:
			go recvNotification(pair)
		}
	}
}

func recvNotification(pair notificationPair) {
	log.WithFields(log.Fields{"university_name": pair.n.University.TopicName,
		"notification_id": pair.n.NotificationId, "status": pair.n.Status,
		"topic": pair.n.TopicName}).Info("postgres_notification")

	defer func(start time.Time) {
		log.WithFields(log.Fields{"elapsed": time.Since(start).Seconds() * 1e3, "name": "send_notification"}).Infoln()
	}(time.Now())

	// Retry in case of SSL/TLS timeout errors. FCM itself should be rock solid
	err := try.Do(func(attempt int) (retry bool, err error) {
		if err = sendGcmNotification(pair); err != nil {
			return true, err
		}
		return false, nil
	})

	if err != nil {
		log.WithError(err).Errorln()
	}
}

func waitForPop() chan notificationPair {
	c := make(chan notificationPair)
	go func() {
		for {
			if pair, err := popNotification(); err == nil {
				c <- *pair
			} else {
				log.WithError(err).Warningln()
			}
		}

	}()
	return c
}

func popNotification() (*notificationPair, error) {
	if topic, err := helper.Client.BRPopLPush(notification.MainQueue, notification.DoneQueue, 0).Result(); err == nil {
		dataNamespace := notification.MainQueueData + topic
		if b, err := helper.Client.Get(dataNamespace).Bytes(); err != nil {
			return nil, errors.Wrap(err, "error getting notification data")
		} else {
			uctNotification := &model.UCTNotification{}
			if err := uctNotification.Unmarshal(b); err != nil {
				return nil, err
			} else if jsonBytes, err := ffjson.Marshal(uctNotification); err != nil {
				return nil, err
			} else if _, err := helper.Client.Del(topic).Result(); err != nil {
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

func sendGcmNotification(pair notificationPair) (err error) {
	httpMessage := gcm.HttpMessage{
		To:               "/topics/" + pair.n.TopicName,
		Data:             map[string]interface{}{"message": pair.raw},
		ContentAvailable: true,
		Priority:         "high",
		DryRun:           *dryRun,
	}

	var httpResponse *gcm.HttpResponse
	if httpResponse, err = gcm.SendHttp(config.Hermes.ApiKey, httpMessage); err != nil {
		return
	}

	log.WithFields(log.Fields{"topic": httpMessage.To,
		"message_id": httpResponse.MessageId, "error": httpResponse.Error}).Infoln("fcm_response")
	// Print FCM errors, but don't panic
	if httpResponse.Error != "" {
		return fmt.Errorf(httpResponse.Error)
	}

	acknowledgeNotification(pair.n.NotificationId, httpResponse.MessageId)
	return
}

type notificationPair struct {
	n   *model.UCTNotification
	raw string
}
