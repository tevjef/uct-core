package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/tevjef/go-gcm"
	"gopkg.in/alecthomas/kingpin.v2"
	_ "net/http/pprof"
	"os"
	"time"
	uct "uct/common"
	"uct/common/conf"
	"github.com/vlad-doru/influxus"
	"uct/influxdb"
)

var (
	workRoutines = 50
)

var (
	app           = kingpin.New("hermes", "A server that listens to a database for events and publishes notifications to Google Cloud Messaging")
	debug         = app.Flag("debug", "enable debug mode").Short('d').Bool()
	configFile    = app.Flag("config", "configuration file for the application").Short('c').File()
	config = conf.Config{}
	database      *sqlx.DB
	preparedStmts = make(map[string]*sqlx.NamedStmt)
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Parse configuration file
	config = conf.OpenConfig(*configFile)
	config.AppName = app.Name

	// Start profiling
	go uct.StartPprof(config.GetDebugSever(app.Name))

	// Start influx logging
	initInflux()

	var err error
	// Open database connection
	database, err = uct.InitDB(config.GetDbConfig(app.Name))
	uct.CheckError(err)
	PrepareAllStmts()

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
					var notification uct.UCTNotification
					err := ffjson.Unmarshal([]byte(pgMessage.Extra), &notification)
					uct.CheckError(err)

					// Process and send notification, free workRoutine when done.
					recvNotification(pgMessage.Extra, notification, sem)
				}()
			// If all workRoutines are filled for 5 minutes, kill the program. Very unlikely that this will happen.
			// Notifications will be lost.
			case <-time.After(5 * time.Minute):
				log.Fatalf("Go routine buffer was filled too long")
			}
		// Received no notification from the database for 90 seconds.
		case <-time.After(1 * time.Minute):
			log.Infoln("Received no events 1 minute, checking connection")
			go func() {
				l.Ping()
			}()
			return
		}
	}
}

func recvNotification(rawNotification string, notification uct.UCTNotification, sem chan int) {
	auditLogger.WithFields(log.Fields{"university_name": notification.University.TopicName, "notification_id": notification.NotificationId, "status": notification.Status, "topic": notification.TopicName}).Info("postgres_notification")

	go func() {
		defer uct.TimeTrackWithLog(time.Now(), auditLogger, "send_notification")

		// Retry in case of SSL/TLS timeout errors. GCM itself should be rock solid
		for i, retries := 0, 3; i < retries; i++ {
			time.Sleep(time.Duration(i*2) * time.Second)
			if err := sendGcmNotification(rawNotification, notification); err != nil {
				log.Errorln("Retrying", i, err)
			} else {
				break
			}
		}
		<-sem
	}()
}

func sendGcmNotification(rawNotification string, pgNotification uct.UCTNotification) (err error) {
	httpMessage := gcm.HttpMessage{
		To:     "/topics/" + pgNotification.TopicName,
		Data:   map[string]interface{}{"message": rawNotification},
		ContentAvailable: true,
		Priority: "high",
		DryRun: *debug,
	}

	var httpResponse *gcm.HttpResponse
	if httpResponse, err = gcm.SendHttp(config.Hermes.ApiKey, httpMessage); err != nil {
		return
	}

	auditLogger.WithFields(log.Fields{"success": httpResponse.Success, "topic": httpMessage.To, "message_id": httpResponse.MessageId, "error":httpResponse.Error, "failure":httpResponse.Failure}).Infoln("fcm_response")
	// Print GCM errors, but don't panic
	if httpResponse.Error != "" {
		return fmt.Errorf(httpResponse.Error)
	}

	acknowledgeNotification(pgNotification.NotificationId, httpResponse.MessageId)

	return
}

func acknowledgeNotification(notificationId, messageId int64) (id int64) {
	args := map[string]interface{}{"notification_id": notificationId, "message_id": messageId}

	if rows, err := GetCachedStmt(AckNotificationQuery).Queryx(args); err != nil {
		log.WithFields(log.Fields{"args": args}).Panic(err)
	} else {
		count := 0
		for rows.Next() {
			count++
			if err = rows.Scan(&id); err != nil {
				log.WithFields(log.Fields{"args": args}).Panic(err)
			}
			rows.Close()
		}
		if count > 1 {
			log.WithFields(log.Fields{"args": args}).Panic("Multiple rows updated at once")
		}
		if id == 0 {
			log.WithFields(log.Fields{"args": args}).Panic(errors.New("Id is 0 retuinig fro updating notification in database"))
		}
	}
	return
}

func GetCachedStmt(query string) *sqlx.NamedStmt {
	if stmt := preparedStmts[query]; stmt == nil {
		preparedStmts[query] = Prepare(query)
	}
	return preparedStmts[query]
}

func Prepare(query string) *sqlx.NamedStmt {
	if named, err := database.PrepareNamed(query); err != nil {
		log.Panicln(fmt.Errorf("Error: %s Query: %s", query, err))
		return nil
	} else {
		return named
	}
}

func PrepareAllStmts() {
	queries := []string{
		AckNotificationQuery,
	}

	for _, query := range queries {
		preparedStmts[query] = Prepare(query)
	}
}

var (
	AckNotificationQuery = `UPDATE notification SET (ack_at, message_id) = (now(), :message_id) WHERE id = :notification_id RETURNING notification.id`
)

var (
	influxClient client.Client
	auditLogger *log.Logger
)

func initInflux() {
	var err error
	// Create the InfluxDB client.
	influxClient, err = influxdbhelper.GetClient(config)

	if err != nil {
		log.Fatalf("Error while creating the client: %v", err)
	}

	// Create and add the hook.
	auditHook, err := influxus.NewHook(
		&influxus.Config{
			Client:             influxClient,
			Database:           "universityct", // DATABASE MUST BE CREATED
			DefaultMeasurement: "hermes_ops",
			BatchSize:          1, // default is 100
			BatchInterval:      1, // default is 5 seconds
			Tags:               []string{"university_name", "status"},
			Precision: "ms",
		})

	uct.CheckError(err)

	// Add the hook to the standard logger.
	auditLogger = log.New()
	auditLogger.Formatter = new(log.JSONFormatter)
	auditLogger.Hooks.Add(auditHook)
}
