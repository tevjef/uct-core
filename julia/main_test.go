package main

import (
	"context"
	"sync/atomic"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/julia/notifier"
)

func Test_waitForNotification(t *testing.T) {
	expected := []string{"a", "b", "c"}

	ch := make(chan string)
	mock := &notifier.FakeNotifier{Notifications: expected, Ch: ch}
	mock.Send()

	waitForNotification(context.Background(), mock, func(notification *model.UCTNotification) {
		assert.Contains(t, expected, notification.TopicName)
		log.Debugf("onNotify %s", notification.TopicName)
		if i := atomic.AddInt32(&mock.Sent, 1); int32(len(expected)) == i {
			close(ch)
		}
	})
}
