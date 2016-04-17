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
	"os"
	"time"
	uct "uct/common"
)

var (
	httpClient   = http.DefaultClient
	sem          = make(chan int, 100)
	connectInfo  = uct.GetUniversityDB()
	influxClient client.Client
)

var (
	app     = kingpin.New("gcm", "A server that listens to a database for events and publishes notifications to Google Cloud Messaging")
	debug   = app.Flag("debug", "Enable debug mode").Short('d').Bool()
	verbose = app.Flag("verbose", "Verbose log of object representations.").Short('v').Bool()
)

const (
	GCM_SEND = "https://gcm-http.googleapis.com/gcm/send"
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
	kingpin.Version("1.0.0")

	var err error

	influxClient, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     uct.INFLUX_HOST,
		Username: uct.INFLUX_USER,
		Password: uct.INFLUX_PASS,
	})
	uct.CheckError(err)

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	listener := pq.NewListener(connectInfo, 10*time.Second, time.Minute, reportProblem)
	err = listener.Listen("status_events")
	if err != nil {
		panic(err)
	}

	fmt.Println("Start monitoring PostgreSQL...")
	for {
		waitForNotification(listener)
	}
}

func waitForNotification(l *pq.Listener) {
	for {
		select {
		case notification := <-l.Notify:
			n := notification
			sem <- 1
			go func() {
				fmt.Println("Received data from channel [", n.Channel, "] :")

				var postgresMessage uct.PostgresNotify
				err := json.Unmarshal([]byte(n.Extra), &postgresMessage)
				uct.CheckError(err)

				go func(message uct.PostgresNotify) {
					auditStatus(message)
					defer func() {
						if r := recover(); r != nil {
							log.Println("Recovered in auditStatus", r)
						}
					}()
				}(postgresMessage)

				uniBytes, err := ffjson.Marshal(postgresMessage.University)
				uct.CheckError(err)

				gcmMessage := uct.GCMMessage{
					To:     "/topics/" + postgresMessage.Payload,
					Data:   uct.Data{Message: string(uniBytes)},
					DryRun: *debug,
				}

				sendNotification(gcmMessage)
				<-sem
			}()
			return
		case <-time.After(90 * time.Second):
			fmt.Println("Received no events for 90 seconds, checking connection")
			go func() {
				l.Ping()
			}()
			return
		}
	}
}

func sendNotification(message uct.GCMMessage) {
	bData, err := ffjson.Marshal(message)
	uct.CheckError(err)

	req, err := http.NewRequest("POST", GCM_SEND, bytes.NewReader(bData))
	uct.CheckError(err)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "key="+uct.GCM_API_KEY)

	fmt.Println("GCM Sending notification: to", message.To)
	LogVerbose(string(bData))

	resp, err := httpClient.Do(req)
	uct.CheckError(err)

	var gcmResponse uct.GCMResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&gcmResponse)
	uct.CheckError(err)

	if gcmResponse.Error != "" {
		fmt.Println(gcmResponse.Error)
	}
	fmt.Println("GCM Response:", gcmResponse, "\nGCM HTTP Status:", resp.StatusCode)

	defer func() {
		resp.Body.Close()
		if r := recover(); r != nil {
			log.Println("Recovered in sendNotification", r)
		}
	}()
}

func auditStatus(message uct.PostgresNotify) {
	subject := message.University.Subjects[0]
	course := subject.Courses[0]
	section := course.Sections[0]

	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "universityct",
		Precision: "s",
	})

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

	err = influxClient.Write(bp)
	uct.CheckError(err)

	fmt.Println("InfluxDB logging: ", tags, fields)
}

func LogVerbose(v interface{}) {
	if *verbose {
		uct.LogVerbose(v)
	}
}
