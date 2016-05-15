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
	influxClient client.Client
	workRoutines = 500

)

const GCM_SEND = "https://gcm-http.googleapis.com/gcm/send"

var (
	app     = kingpin.New("gcm", "A server that listens to a database for events and publishes notifications to Google Cloud Messaging")
	debug   = app.Flag("debug", "Enable debug mode").Short('d').Bool()
	verbose = app.Flag("verbose", "Verbose log of object representations.").Short('v').Bool()
)

func main() {
	go func() {
		log.Println("**Starting debug server on...", uct.GCM_DEBUG_SERVER)
		log.Println(http.ListenAndServe(uct.GCM_DEBUG_SERVER, nil))
	}()
	kingpin.MustParse(app.Parse(os.Args[1:]))

	var err error

	influxClient, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     uct.INFLUX_HOST,
		Username: uct.INFLUX_USER,
		Password: uct.INFLUX_PASS,
	})
	uct.CheckError(err)

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Println(err.Error())
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
	sem := make(chan int, workRoutines)
	for {
		select {
		case notification := <-l.Notify:
			select {
			// Acquire a workRoutine
			case sem <- 1:
				// Log notification
				audit(notification)
				// Process and send notification, free workRoutine when done.
				recvNotification(notification, sem)
			// If all workRoutines are filled for 5 minutes, kill the program. Very unlikely that this will happen.
			// Notifications will be lost.
			case <-time.After(5 * time.Minute):
				log.Fatalf("Go routine buffer was filled too long")
			}
		// Received no notification from the database for 90 seconds.
		case <-time.After(90 * time.Second):
			fmt.Println("Received no events for 90 seconds, checking connection")
			go func() {
				l.Ping()
			}()
			return
		}
	}
}

func recvNotification(n *pq.Notification, sem chan int) {
	go func() {
		var pgNotification uct.PostgresNotify
		err := json.Unmarshal([]byte(n.Extra), &pgNotification)
		uct.CheckError(err)
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

	fmt.Println("InfluxDB log: ", tags, fields)
}

func audit(message uct.PostgresNotify) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in auditStatus", r)
		}
	}()
	go func() {
		auditStatus(message)
	}()
}

func LogVerbose(v interface{}) {
	if *verbose {
		uct.LogVerbose(v)
	}
}
