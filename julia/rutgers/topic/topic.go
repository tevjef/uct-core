package topic

import (
	"time"
	"uct/common/model"

	"golang.org/x/net/context"
)

type Routine struct {
	expiration time.Duration
	first      model.UCTNotification
	last       model.UCTNotification
	in         chan model.UCTNotification
	out        chan model.UCTNotification
	done       chan string
	ctx        context.Context
}

func NewTopicRoutine(expiration time.Duration) *Routine {
	routine := &Routine{
		in:         make(chan model.UCTNotification),
		out:        make(chan model.UCTNotification),
		done:       make(chan string),
		ctx:        context.Background(),
		expiration: expiration,
	}

	go routine.loop()

	return routine
}

func (tr *Routine) Send(uctNotification model.UCTNotification) {
	tr.in <- uctNotification
}

func (tr *Routine) Done() <-chan string {
	return tr.done
}

func (tr *Routine) Out() <-chan model.UCTNotification {
	return tr.out
}

func (tr *Routine) loop() {
	defer func() {
		tr.done <- tr.last.TopicName
		close(tr.done)
	}()

	for {
		after := time.After(tr.expiration)
		select {
		case <-after:
			//log.Printf("timer complete %+v\n", tr)
			if tr.first.Status != tr.last.Status {
				tr.out <- tr.last
			}
			return
		case <-tr.ctx.Done():
			//log.Printf("ctx done %+v\n", tr)
			return
		case notification := <-tr.in:
			if (tr.first.NotificationId == 0) && (tr.last.NotificationId == 0) {
				tr.first = notification
				tr.last = notification

				// send first notification
				tr.out <- tr.first
			}

			//log.Printf("recv %+v\n", notification)
			if tr.last.Status != notification.Status {
				tr.last = notification
			}
		}
	}
}
