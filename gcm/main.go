package main

import (
	"bytes"
	"fmt"
	"github.com/lib/pq"
	"time"
	uct "uct/common"
	"net/http"
	"github.com/pquerna/ffjson/ffjson"
	"encoding/json"
)

var (
	httpClient    = http.DefaultClient
	sem      = make(chan int, 100)
	connectInfo   = uct.GetUniversityDB(false)
	notifications int
)

const (
	GCM_SEND = "https://gcm-http.googleapis.com/gcm/send"
)

func main() {
	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(ev)
	}

	listener := pq.NewListener(connectInfo, 10*time.Second, time.Minute, reportProblem)
	err := listener.Listen("status_events")
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
			notifications++
			sem <- 1
			go func() {
				fmt.Println("Received data from channel [", n.Channel, "] :")
				//fmt.Println(n.Extra)
				var postgresMessage uct.PostgresNotify
				ffjson.UnmarshalFast([]byte(n.Extra), &postgresMessage)

				gcmMessage := uct.GCMMessage{
					To:     "/topics/" + postgresMessage.Payload,
					Data:   uct.Data{Message: "This was a successful message"},
					DryRun: false,
				}

				sendNotification(gcmMessage)
				<-sem
			}()
			fmt.Println("Notifications:", notifications)
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

	fmt.Println("Sending Request: ", string(bData))
	resp, err := httpClient.Do(req)
	uct.CheckError(err)

	fmt.Println("Response from", GCM_SEND)
	var gcmResponse uct.GCMResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&gcmResponse)
	uct.CheckError(err)

	if gcmResponse.Error != "" {
		fmt.Println(gcmResponse.Error)
	}
	fmt.Println("HTTP Status:", resp.StatusCode, "\nResponse:", gcmResponse)

	defer resp.Body.Close()
}
