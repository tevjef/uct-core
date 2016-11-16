package main

import (
	"testing"
	"uct/common/model"
	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"encoding/json"
)

func Test_waitForNotification(t *testing.T) {
	expected := []string{"Tevin", "Wade", "Jeffrey"}

	ch := make(chan string)
	mock := &MockNotifier{expected, ch}
	go mock.send()

	s := 0
	waitForNotification(mock, func(notification *model.UCTNotification) {
		assert.Contains(t, expected, notification.TopicName)
		log.Debugf("onNotify %s", notification.TopicName)
		s++
		if s == len(expected) - 1 {
			close(ch)
		}
	})
}

type MockNotifier struct {
	notifications []string
	ch chan string
}

func (pg *MockNotifier) send() {
	go func() {
		for _, val := range pg.notifications {
			fakeNoti := model.UCTNotification{TopicName:val}
			b, _ := json.Marshal(fakeNoti)
			pg.ch <- string(b)
		}
	}()
}

func (pg *MockNotifier) Notify() <-chan string {
	return pg.ch
}

func (pg *MockNotifier) Ping() {
	log.Debugln("Pinging notifier")
}