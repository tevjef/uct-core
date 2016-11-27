package topic

import (
	"testing"
	"time"
	"uct/common/model"

	"log"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func Test_collapse(t *testing.T) {
	data := []model.UCTNotification{
		{NotificationId: 1, TopicName: "section.1", Status: "Open"},
		{NotificationId: 2, TopicName: "section.1", Status: "Closed"},
		{NotificationId: 3, TopicName: "section.1", Status: "Open"},
		{NotificationId: 4, TopicName: "section.1", Status: "Closed"},
	}

	in := make(chan model.UCTNotification)
	out := make(chan model.UCTNotification)
	done := make(chan string)
	ctx := context.Background()

	go func() {
		for _, val := range data {
			in <- val
			time.Sleep(time.Second * 1)
		}
	}()

	routine := &Routine{
		expiration: 2 * time.Second,
		in:         in,
		out:        out,
		done:       done,
		ctx:        ctx,
	}
	go routine.loop()

	expected := []model.UCTNotification{
		{NotificationId: 1, TopicName: "section.1", Status: "Open"},
		{NotificationId: 4, TopicName: "section.1", Status: "Closed"},
	}

	result := []model.UCTNotification{}

	testDone := make(chan struct{})
	go func() {
		for {
			select {
			case uctNotification := <-out:
				log.Printf("test noti %+v\n", uctNotification)
				result = append(result, uctNotification)
			case <-done:
				assert.Equal(t, expected, result)
				testDone <- struct{}{}
			}
		}

	}()
	<-testDone
}

func Test_collapse2(t *testing.T) {
	data := []model.UCTNotification{
		{NotificationId: 1, TopicName: "section.1", Status: "Open"},
		{NotificationId: 2, TopicName: "section.1", Status: "Closed"},
		{NotificationId: 3, TopicName: "section.1", Status: "Open"},
	}

	in := make(chan model.UCTNotification)
	out := make(chan model.UCTNotification)
	done := make(chan string)
	ctx := context.Background()

	go func() {
		for _, val := range data {
			in <- val
			time.Sleep(time.Second * 1)
		}
	}()

	routine := &Routine{
		expiration: 2 * time.Second,
		in:         in,
		out:        out,
		done:       done,
		ctx:        ctx,
	}
	go routine.loop()

	expected := []model.UCTNotification{
		{NotificationId: 1, TopicName: "section.1", Status: "Open"},
	}

	result := []model.UCTNotification{}

	testDone := make(chan struct{})
	go func() {
		for {
			select {
			case uctNotification := <-out:
				log.Printf("test noti %+v\n", uctNotification)
				result = append(result, uctNotification)
			case <-done:
				assert.Equal(t, expected, result)
				testDone <- struct{}{}
			}
		}

	}()
	<-testDone
}
