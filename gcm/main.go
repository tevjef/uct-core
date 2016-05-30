package main

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/influxdata/influxdb/client/v2"
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
	messageLog   = make(chan uct.PostgresNotify)
)

var (
	app    = kingpin.New("gcm", "A server that listens to a database for events and publishes notifications to Google Cloud Messaging")
	debug  = app.Flag("debug", "enable debug mode").Short('d').Bool()
	server = app.Flag("pprof", "host:port to start profiling on").Short('p').Default(uct.GCM_DEBUG_SERVER).TCP()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)
	// Start profiling
	go uct.StartPprof(*server)
	// Start influx logging
	go influxLog()

	// Open connection to postgresql
	listener := pq.NewListener(uct.GetUniversityDB(), 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
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
		case notification := <-l.Notify:
			select {
			// Acquire a workRoutine
			case sem <- 1:
				go func() {
					var pgNotification uct.PostgresNotify
					err := json.Unmarshal([]byte(notification.Extra), &pgNotification)
					uct.CheckError(err)
					// Log notification
					messageLog <- pgNotification
					// Process and send notification, free workRoutine when done.
					recvNotification(pgNotification, sem)
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

func recvNotification(pgNotification uct.PostgresNotify, sem chan int) {
	log.WithFields(log.Fields{"status": pgNotification.Status, "topic": pgNotification.Payload}).Info("PostgresNotify")

	go func() {
		defer uct.TimeTrack(time.Now(), "SendNotification")
		// Retry in case of SSL/TLS timeout errors. GCM itself should be rock solid. 10 retries is quite generous.
		for i, retries := 0, 10; i < retries; i++ {
			time.Sleep(time.Duration(i*2) * time.Second)
			if err := sendGcmNotification(pgNotification); err != nil {
				log.Errorln("Retrying", i, err)
			} else if err == nil {
				break
			}
		}
		<-sem
	}()
}

func sendGcmNotification(pgNotification uct.PostgresNotify) (err error) {
	var uniBytes []byte
	if uniBytes, err = ffjson.Marshal(pgNotification.University); err != nil {
		return
	}

	message := gcm.HttpMessage{
		To:     "/topics/" + pgNotification.Payload,
		Data:   map[string]interface{}{"message": string(uniBytes)},
		DryRun: *debug,
	}
	return sendNotification(message)
}

func sendNotification(message gcm.HttpMessage) (err error) {
	var httpResponse *gcm.HttpResponse

	if httpResponse, err = gcm.SendHttp(uct.GCM_API_KEY, message); err != nil {
		return
	}
	log.WithFields(log.Fields{"topic": message.To, "dry_run": message.DryRun, "gcm_message_id": httpResponse.MessageId}).Infoln("GCMResponse")
	// Print GCM errors, but don't panic
	if httpResponse.Error != "" {
		return fmt.Errorf(httpResponse.Error)
	}
	return
}

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
