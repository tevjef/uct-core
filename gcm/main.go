package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/lib/pq"
	"github.com/pquerna/ffjson/ffjson"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
	uct "uct/common"
)

var (
	httpClient   = http.DefaultClient
	connectInfo  = uct.GetUniversityDB()
	workRoutines = 500
	messageLog   = make(chan uct.PostgresNotify)
)

const GCM_SEND = "https://gcm-http.googleapis.com/gcm/send"

var (
	app     = kingpin.New("gcm", "A server that listens to a database for events and publishes notifications to Google Cloud Messaging")
	debug   = app.Flag("debug", "enable debug mode").Short('d').Bool()
	server  = app.Flag("pprof", "host:port to start profiling on").Short('p').Default(uct.GCM_DEBUG_SERVER).TCP()
	verbose = app.Flag("verbose", "verbose log of object representations.").Short('v').Bool()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Start profiling
	go uct.StartPprof(*server)
	// Start influx logging
	go influxLog()

	// Open connection to postgresql
	listener := pq.NewListener(connectInfo, 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Println(err.Error())
		}
	})

	// Listen on channel
	if err := listener.Listen("status_events"); err != nil {
		panic(err)
	}

	fmt.Println("Start monitoring PostgreSQL...")
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
					//messageLog <- pgNotification
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
			fmt.Println("Received no events 1 minute, checking connection")
			go func() {
				l.Ping()
			}()
			return
		}
	}
}

func recvNotification(pgNotification uct.PostgresNotify, sem chan int) {
	go func() {
		defer uct.TimeTrack(time.Now(), "Send notification")
		// Retry in case of SSL/TLS timeout errors. GCM itself should be rock solid. 10 retries is quite generous.
		for i, retries := 0, 10; i < retries; i++ {
			time.Sleep(time.Duration(i*2) * time.Second)
			if i > 0 {
				log.Println("Retrying...", i)
			}
			if err := sendGcmNotification(pgNotification); err != nil && i == retries-1 {
				log.Println("ERROR:", err, string(uct.Stack(3)))
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

	// Construct GCM message
	gcmMessage := uct.GCMMessage{
		To:     "/topics/" + pgNotification.Payload,
		Data:   uct.Data{Message: string(uniBytes)},
		DryRun: *debug,
	}
	return sendNotification(gcmMessage)
}

func sendNotification(message uct.GCMMessage) (err error) {
	var messageBytes []byte
	if messageBytes, err = ffjson.Marshal(message); err != nil {
		return
	}
	LogVerbose(string(messageBytes))

	// New request to GCM
	var req *http.Request
	if req, err = http.NewRequest("POST", GCM_SEND, bytes.NewReader(messageBytes)); err != nil {
		return
	}

	// Add request authorization headers so that GCM knows who this is.
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "key="+uct.GCM_API_KEY)

	fmt.Println("GCM Sending notification: to", message.To)

	// Send request to GCM
	var response *http.Response
	if response, err = httpClient.Do(req); err != nil {
		return
	}
	defer response.Body.Close()

	// Read response from GCM
	var gcmResponse uct.GCMResponse
	if err = json.NewDecoder(response.Body).Decode(&gcmResponse); err != nil {
		return
	}

	// Print GCM errors, but don't panic
	if gcmResponse.Error != "" {
		log.Println(gcmResponse.Error)
	}

	fmt.Println("GCM Response: ", gcmResponse, " Status: ", response.StatusCode)
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
						log.Println("Recovered from influx error", r, string(uct.Stack(3)))
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
				fmt.Println("InfluxDB log: ", tags, fields)
			}()
		case <-time.NewTicker(time.Minute).C:
			/*	err := influxClient.Write(bp)
				uct.CheckError(err)
				bp, err = client.NewBatchPoints(client.BatchPointsConfig{
					Database:  "universityct",
					Precision: "s",
				})
				uct.CheckError(err)*/
		}
	}
}

func LogVerbose(v interface{}) {
	if *verbose {
		uct.LogVerbose(v)
	}
}
