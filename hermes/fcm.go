package main

import (
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/go-fcm"
	"github.com/tevjef/uct-core/common/model"
)

func (hermes *hermes) sendFcmNotification(pair notificationPair) error {
	//Sent to "/topics/android:" + pair.n.TopicName | with android payload
	//Sent to "/topics/ios:" + pair.n.TopicName | with ios notification payload
	//Sent to "/topics/" + pair.n.TopicName | for backwards compatibility

	sendReq := &fcm.SendRequest{
		ValidateOnly: hermes.config.dryRun,
		Message: &fcm.Message{
			Topic:        pair.n.TopicName,
			Notification: globalNotification(pair.n),
		},
	}

	resp, err := hermes.fcmClient.Send(sendReq)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{"topic": pair.n.TopicName, "university_name": pair.n.University.TopicName}).Infoln("fcm_response")

	msgID, err := strconv.ParseInt(resp.MessageID(), 10, 64)
	if err != nil {
		return nil
	}

	log.Panicln(msgID)

	//hermes.acknowledgeNotification(pair.n.NotificationId, msgID)

	return nil
}

func globalNotification(n *model.UCTNotification) *fcm.Notification {
	var title string
	var body string

	course := n.University.Subjects[0].Courses[0]
	section := course.Sections[0]

	if n.Status == "Open" {
		title = "A section has opened!"
		body = "Section " + section.Number + " of " + course.Name + " has opened!"
	} else {
		title = "A section has closed!"
		body = "Section " + section.Number + " of " + course.Name + " has closed!"
	}

	return &fcm.Notification{
		Title: title,
		Body:  body,
	}
}
