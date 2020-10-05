package uctfirestore

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FirestoreNotificationSent struct {
	OpenCount           int       `firestore:"openCount"`
	ClosedCount         int       `firestore:"closedCount"`
	LastStatus          string    `firestore:"lastStatus"`
	LastStatusUpdatedAt time.Time `firestore:"lastStatusUpdatedAt"`
}

func (client Client) InsertSectionNotification(sectionNotification *SectionNotification) error {
	field := log.Fields{"collection": CollectionNotificationSent, "sectionNotification": sectionNotification}

	collection := client.fsClient.Collection(CollectionNotificationSent)
	docRef := collection.Doc(sectionNotification.Section.TopicName)
	return client.fsClient.RunTransaction(client.context, func(context context.Context, tx *firestore.Transaction) error {
		docSnap, err := tx.Get(docRef)
		docNotFound := status.Code(err) == codes.NotFound
		if err != nil && status.Code(err) != codes.NotFound {
			client.logger.WithError(err).WithFields(field).Fatalln("firestore: failed to get docRef")
			return err
		}

		fns := FirestoreNotificationSent{}
		if !docNotFound {
			err = docSnap.DataTo(&fns)
			if err != nil {
				client.logger.WithError(err).WithFields(field).Fatalln("firestore: failed to map FirestoreNotificationSent")
				return err
			}
		} else {
			client.logger.WithFields(field).Debugln("firestore: existing notification found!")
		}

		if sectionNotification.Section.Status == "Open" {
			fns.OpenCount = fns.ClosedCount + 1
		} else {
			fns.ClosedCount = fns.ClosedCount + 1
		}

		fns.LastStatus = sectionNotification.Section.Status
		fns.LastStatusUpdatedAt = time.Now()

		err = tx.Set(docRef, fns)
		if err != nil {
			client.logger.WithError(err).WithFields(field).Fatalln("firestore: failed to set notification")
			return err
		}

		client.logger.WithError(err).WithFields(field).Infoln("firestore: inserted notification")

		return nil
	})
}

type DeviceNotification struct {
	NotificationId   string    `firebase:"notificationId"`
	SectionTopicName string    `firestore:"topicName"`
	ReceivedAt       time.Time `firestore:"receivedAt"`
	FcmToken         string    `firestore:"fcmToken"`
	IsSubscribed     bool      `firestore:"isSubscribed"`
	Os               string    `firestore:"os"`
	OsVersion        string    `firestore:"osVersion"`
	AppVersion       string    `firestore:"appVersion"`
}

func (client Client) InsertDeviceNotification(deviceNotification *DeviceNotification) error {
	field := log.Fields{"collection": CollectionNotificationReceived, "deviceNotification": deviceNotification}

	collection := client.fsClient.Collection(CollectionNotificationReceived)
	_, result, err := collection.Add(client.context, deviceNotification)
	if err != nil {
		client.logger.WithError(err).
			WithField("result", result).
			WithFields(field).Errorln("firestore: failed to add DeviceNotification")
		return errors.Wrap(err, "firestore: failed to add DeviceNotification")
	}

	return nil
}
