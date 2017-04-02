package notifier

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/uct-core/common/model"
)

type FakeNotifier struct {
	Notifications []string
	Ch            chan string
	Sent          int32
}

func (mn *FakeNotifier) Send() {
	go func() {
		for _, val := range mn.Notifications {
			fakeNoti := model.UCTNotification{TopicName: val}
			b, _ := json.Marshal(fakeNoti)
			mn.Ch <- string(b)
		}
	}()
}

func (mn *FakeNotifier) Notify() <-chan string {
	return mn.Ch
}

func (mn *FakeNotifier) Ping() {
	log.Debugln("Pinging notifier")
}
