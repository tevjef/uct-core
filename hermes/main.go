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
)

var (
	workRoutines = 100
	messageLog   = make(chan uct.UCTNotification)
)

var (
	app           = kingpin.New("gcm", "A server that listens to a database for events and publishes notifications to Google Cloud Messaging")
	debug         = app.Flag("debug", "enable debug mode").Short('d').Bool()
	server        = app.Flag("pprof", "host:port to start profiling on").Short('p').Default(uct.GCM_DEBUG_SERVER).TCP()
	database      *sqlx.DB
	preparedStmts = make(map[string]*sqlx.NamedStmt)
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
	log.SetFormatter(&log.TextFormatter{})
	// Start profiling
	go uct.StartPprof(*server)
	// Start influx logging
	go influxLog()

	dbConnectionString := uct.GetUniversityDB()
	var err error
	// Open database connection
	database, err = uct.InitDB(dbConnectionString)
	uct.CheckError(err)
	PrepareAllStmts()

	// Open connection to postgresql
	listener := pq.NewListener(dbConnectionString, 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
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
					log.Infoln("Recieve notification")

					err := ffjson.Unmarshal([]byte(pgMessage.Extra), &notification)
					uct.CheckError(err)
					// Log notification
					messageLog <- notification
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
	log.WithFields(log.Fields{"notification_id": notification.NotificationId, "status": notification.Status, "topic": notification.TopicName}).Info("PostgresNotify")

	go func() {
		defer uct.TimeTrack(time.Now(), "SendNotification")

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
	if httpResponse, err = gcm.SendHttp(uct.GCM_API_KEY, httpMessage); err != nil {
		return
	}

	log.WithFields(log.Fields{"topic": httpMessage.To, "message_id": httpResponse.MessageId}).Infoln("FCMResponse")
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

func influxLog() {
	influxClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     uct.INFLUX_HOST,
		Username: uct.INFLUX_USER,
		Password: uct.INFLUX_PASS,
	})
	uct.CheckError(err)

	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "universityct",
		Precision: "s",
	})

	for {
		select {
		case message := <-messageLog:
			go func() {
				defer func() {
					if r := recover(); r != nil {
						log.Println("Recovered from influx error", r)
					}
				}()
				subject := message.University.Subjects[0]
				course := subject.Courses[0]
				section := course.Sections[0]

				tags := map[string]string{
					"university": message.University.TopicName,
					"subject":    subject.TopicName,
					"course":     course.TopicName,
					"semester":   subject.Season + subject.Year,
					"topic_name": section.TopicName,
				}

				fields := map[string]interface{}{
					"now":    int(section.Now),
					"status": section.Status,
				}

				point, err := client.NewPoint(
					"section_status",
					tags,
					fields,
					time.Now(),
				)
				uct.CheckError(err)
				bp.AddPoint(point)
			}()
		case <-time.NewTicker(time.Minute).C:

			err := influxClient.Write(bp)
			uct.CheckError(err)
			bp, err = client.NewBatchPoints(client.BatchPointsConfig{
				Database:  "universityct",
				Precision: "s",
			})
			uct.CheckError(err)
		}
	}
}
