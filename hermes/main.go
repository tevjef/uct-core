package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"time"
	"uct/common/conf"
	"uct/common/model"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/tevjef/go-gcm"
	"gopkg.in/alecthomas/kingpin.v2"
	"strconv"
)

var (
	workRoutines = 50
)

var (
	app           = kingpin.New("hermes", "A server that listens to a database for events and publishes notifications to Google Cloud Messaging")
	dryRun        = app.Flag("dry-run", "enable dry-run").Short('d').Bool()
	configFile    = app.Flag("config", "configuration file for the application").Short('c').File()
	config        = conf.Config{}
	database      *sqlx.DB
	preparedStmts = make(map[string]*sqlx.NamedStmt)
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Parse configuration file
	config = conf.OpenConfig(*configFile)
	config.AppName = app.Name

	if enableFcm, _ := strconv.ParseBool(os.Getenv("ENABLE_FCM")); enableFcm {
		*dryRun = false
	}

	// Start profiling
	go model.StartPprof(config.GetDebugSever(app.Name))

	var err error
	// Open database connection
	database, err = model.InitDB(config.GetDbConfig(app.Name))
	model.CheckError(err)
	prepareAllStmts()

	// Open connection to postgresql
	listener := pq.NewListener(config.GetDbConfig(app.Name), 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Println(err.Error())
		}
	})

	// Listen on channel
	if err := listener.Listen("status_events"); err != nil {
		panic(err)
	}

	log.Infoln("Start monitoring PostgreSQL...")
	for {
		waitForNotification(listener)
	}
}

func waitForNotification(l *pq.Listener) {
	sem := make(chan int, workRoutines)
	for {
		select {
		case pgMessage := <-l.Notify:
			select {
			// Acquire a workRoutine
			case sem <- 1:
				go func() {
					var notification model.UCTNotification
					err := ffjson.Unmarshal([]byte(pgMessage.Extra), &notification)
					model.CheckError(err)

					// Process and send notification, free workRoutine when done.
					recvNotification(pgMessage.Extra, notification)
					sem <- 1
				}()
			// If all workRoutines are filled for 5 minutes, kill the program. Very unlikely that this will happen.
			// Notifications will be lost.
			case <-time.After(5 * time.Minute):
				log.Fatalf("Go routine buffer was filled too long")
			}
		// Received no notification from the database for 60 seconds.
		case <-time.After(1 * time.Minute):
			log.Infoln("Received no events 1 minute, checking connection")
			go func() {
				l.Ping()
			}()
			return
		}
	}
}

func recvNotification(rawNotification string, notification model.UCTNotification) {
	log.WithFields(log.Fields{"university_name": notification.University.TopicName,
		"notification_id": notification.NotificationId, "status": notification.Status,
		"topic": notification.TopicName}).Info("postgres_notification")
	defer model.TimeTrackWithLog(time.Now(), log.StandardLogger(), "send_notification")

	// Retry in case of SSL/TLS timeout errors. FCM itself should be rock solid
	for i, retries := 0, 3; i < retries; i++ {
		time.Sleep(time.Duration(i*2) * time.Second)
		if err := sendGcmNotification(rawNotification, notification); err != nil {
			log.Errorln("Retrying", i, err)
		} else {
			break
		}
	}
}

func sendGcmNotification(rawNotification string, pgNotification model.UCTNotification) (err error) {
	httpMessage := gcm.HttpMessage{
		To:               "/topics/" + pgNotification.TopicName,
		Data:             map[string]interface{}{"message": rawNotification},
		ContentAvailable: true,
		Priority:         "high",
		DryRun:           *dryRun,
	}

	var httpResponse *gcm.HttpResponse
	if httpResponse, err = gcm.SendHttp(config.Hermes.ApiKey, httpMessage); err != nil {
		return
	}

	log.WithFields(log.Fields{"success": httpResponse.Success, "topic": httpMessage.To,
		"message_id": httpResponse.MessageId, "error": httpResponse.Error,
		"failure": httpResponse.Failure}).Infoln("fcm_response")
	// Print FCM errors, but don't panic
	if httpResponse.Error != "" {
		return fmt.Errorf(httpResponse.Error)
	}

	acknowledgeNotification(pgNotification.NotificationId, httpResponse.MessageId)

	return
}
