package uctfirestore

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	log "github.com/Sirupsen/logrus"
)

type FirestoreNotificationSent struct {
	OpenCount           int       `firestore:"openCount"`
	ClosedCount         int       `firestore:"closedCount"`
	LastStatus          string    `firestore:"lastStatus"`
	LastStatusUpdatedAt time.Time `firestore:"lastStatusUpdatedAt"`
}

func (client Client) InsertNotification(sectionNotification *SectionNotification) error {
	field := log.Fields{"collection": CollectionNotificationSent}

	collection := client.fsClient.Collection(CollectionNotificationSent)
	docRef := collection.Doc(sectionNotification.Section.TopicName)
	return client.fsClient.RunTransaction(client.context, func(context context.Context, tx *firestore.Transaction) error {
		docSnap, err := tx.Get(docRef)
		if err != nil {
			client.logger.WithError(err).WithFields(field).Fatalln("firestore: failed to get docRef")
			return err
		}

		fns := FirestoreNotificationSent{}
		err = docSnap.DataTo(&fns)
		if err != nil {
			client.logger.WithError(err).WithFields(field).Fatalln("firestore: failed to map FirestoreNotificationSent")
			return err
		}

		if sectionNotification.Section.Status == "Open" {
			fns.OpenCount = fns.ClosedCount + 1
		} else {
			fns.ClosedCount = fns.ClosedCount + 1
		}

		fns.LastStatus = sectionNotification.Section.Status
		fns.LastStatusUpdatedAt = time.Now()

		err = tx.Set(docRef, fns, firestore.MergeAll)
		if err != nil {
			client.logger.WithError(err).WithFields(field).Fatalln("firestore: failed to set notification")
			return err
		}

		client.logger.WithError(err).WithFields(field).Infoln("firestore: inserted notification")

		return nil
	})
}
