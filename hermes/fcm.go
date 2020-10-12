package hermes

import (
	"firebase.google.com/go/messaging"
	log "github.com/sirupsen/logrus"
	uctfirestore "github.com/tevjef/uct-backend/common/firestore"
)

func (hermes *hermes) sendNotification(sectionNotification *uctfirestore.SectionNotification) error {
	var title string
	var body string
	var color string

	if sectionNotification.Section.Status == "Open" {
		title = "A section has opened!"
		body = "Section " + sectionNotification.Section.Number + " of " + sectionNotification.CourseName + " has opened!"
		color = "#4CAF50"
	} else {
		title = "A section has closed!"
		body = "Section " + sectionNotification.Section.Number + " of " + sectionNotification.CourseName + " has closed!"
		color = "#F44336"
	}

	data := map[string]interface{}{
		"status":          sectionNotification.Section.Status,
		"topicName":       sectionNotification.Section.TopicName,
		"topicId":         sectionNotification.Section.TopicId,
		"title":           title,
		"body":            body,
		"color":           color,
		"registrationUrl": sectionNotification.University.RegistrationPage,
	}

	badge := 1

	apnsConfig := &messaging.APNSConfig{
		Payload: &messaging.APNSPayload{
			Aps: &messaging.Aps{
				Alert: &messaging.ApsAlert{
					Title: title,
					Body:  body,
				},
				Badge:            &badge,
				Category:         "SECTION_NOTIFICATION_CATEGORY",
				ContentAvailable: false,
			},
			// Only applications on the in the foreground get this data. Notifications only show in the foreground
			CustomData: data,
		},
	}

	androidConfig := &messaging.AndroidConfig{
		Notification: &messaging.AndroidNotification{
			Title:       title,
			Body:        body,
			Color:       color,
			ClickAction: "NOTIFICATION_CLICK_ACTION",
		},
	}

	message := &messaging.Message{
		Topic:   sectionNotification.Section.TopicName,
		APNS:    apnsConfig,
		Android: androidConfig,
	}

	field := map[string]interface{}{
		"is_dry_run": hermes.config.dryRun,
		"section":    sectionNotification.Section.TopicName,
	}

	if hermes.config.dryRun {
		result, err := hermes.fcmClient.SendDryRun(hermes.ctx, message)
		log.WithError(err).WithFields(field).Infof("dry run notification sent: %v %v", sectionNotification.Section.TopicName, result)
		if err != nil {
			return err
		}
	} else {
		result, err := hermes.fcmClient.Send(hermes.ctx, message)
		log.WithError(err).WithFields(field).Infof("prod notification sent: %v %v", sectionNotification.Section.TopicName, result)
		if err != nil {
			return err
		}
	}

	return nil
}
