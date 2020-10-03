package hermes

import (
	"fmt"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/go-fcm"
)

func (hermes *hermes) sendFcmNotification(pair notificationPair) error {
	var title string
	var body string
	var color string

	course := pair.n.University.Subjects[0].Courses[0]
	section := course.Sections[0]

	if pair.n.Status == "Open" {
		title = "A section has opened!"
		body = "Section " + section.Number + " of " + course.Name + " has opened!"
		color = "#4CAF50"
	} else {
		title = "A section has closed!"
		body = "Section " + section.Number + " of " + course.Name + " has closed!"
		color = "#F44336"
	}

	data := map[string]string{
		"notificationId":  fmt.Sprintf("%d", pair.n.NotificationId),
		"status":          pair.n.Status,
		"topicName":       pair.n.TopicName,
		"topicId":         section.TopicId,
		"title":           title,
		"body":            body,
		"color":           color,
		"registrationUrl": pair.n.University.RegistrationPage,
	}

	apnsPayload := &fcm.ApnsPayload{
		Aps: &fcm.ApsDictionary{
			Alert: &fcm.ApnsAlert{
				Title: title,
				Body:  body,
			},
			Badge:            1,
			Category:         "SECTION_NOTIFICATION_CATEGORY",
			ContentAvailable: int(fcm.ApnsContentUnavailable),
		},
	}

	payload, err := apnsPayload.ToMap()
	if err != nil {
		return err
	}

	// Only applications on the in the foreground get this data. Notifications only show in the foreground
	payload["message"] = data

	sendReq := &fcm.SendRequest{
		ValidateOnly: hermes.config.dryRun,
		Message: &fcm.Message{
			Topic: pair.n.TopicName,
			Android: &fcm.AndroidConfig{
				// Won't work on preupdate devices
				// Data: data,
				Notification: &fcm.AndroidNotification{
					Title:       title,
					Body:        body,
					Color:       color,
					ClickAction: "NOTIFICATION_CLICK_ACTION",
				},
			},
			Apns: &fcm.ApnsConfig{
				Payload: payload,
			},
		},
	}

	resp, err := hermes.fcmClient.Send(sendReq)
	if err != nil {
		if v, ok := err.(fcm.HttpError); ok {
			log.Error(v.RequestDump)
			log.Error(v.ResponseDump)
		}

		return err
	}

	msgID, err := strconv.ParseInt(resp.MessageID(), 10, 64)
	if err != nil {
		return nil
	}

	log.WithFields(log.Fields{
		"topic":           pair.n.TopicName,
		"university_name": pair.n.University.TopicName,
		"message_id":      msgID}).Infoln("fcm_response")

	hermes.acknowledgeNotification(pair.n.NotificationId, msgID)

	return nil
}
