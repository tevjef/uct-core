package uctfirestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Subscription struct {
	SectionTopicName string `firestore:"topicName"`
	FcmToken         string `firestore:"fcmToken"`
	IsSubscribed     bool   `firestore:"isSubscribed"`
	Os               string `firestore:"os"`
	OsVersion        string `firestore:"osVersion"`
	AppVersion       string `firestore:"appVersion"`
}

type SubscriptionCount struct {
	subscriberCount int `firestore:"subscriberCount"`
}

func (client Client) InsertSubscriptionAndUpdateCount(ctx context.Context, subscription *Subscription) error {
	field := log.Fields{"collection": CollectionSubscriptions, "subscription": subscription}
	ctx, span := makeFirestoreTrace(ctx, "InsertSubscriptionAndUpdateCount", field, client.logger.Data)
	defer span.End()

	collection := client.fsClient.Collection(CollectionSubscriptions)

	_, result, err := collection.Add(ctx, subscription)
	if err != nil {
		client.logger.WithError(err).WithField("result", result).WithFields(field).Errorln("firestore: failed to add subscription")
		return err
	}

	return client.UpdateSubscription(ctx, subscription)
}

func (client Client) UpdateSubscription(ctx context.Context, subscription *Subscription) error {
	field := log.Fields{"collection": CollectionSubscriptionsCount, "subscription": subscription}
	ctx, span := makeFirestoreTrace(ctx, "UpdateSubscription", field, client.logger.Data)
	defer span.End()

	collection := client.fsClient.Collection(CollectionSubscriptionsCount)

	docRef := collection.Doc(subscription.SectionTopicName)
	err := client.fsClient.RunTransaction(ctx, func(context context.Context, tx *firestore.Transaction) error {
		docSnap, err := tx.Get(docRef)
		docNotFound := status.Code(err) == codes.NotFound
		if err != nil && status.Code(err) != codes.NotFound {
			return errors.Wrap(err, "firestore: failed to get docRef")
		}

		sc := SubscriptionCount{}
		if !docNotFound {
			err = docSnap.DataTo(&sc)
			if err != nil {
				return errors.Wrap(err, "firestore: failed to map SubscriptionCount")
			}
		} else {
			client.logger.WithFields(field).Debugln("firestore: existing subscription count!")
		}

		if subscription.IsSubscribed {
			sc.subscriberCount = sc.subscriberCount + 1
		} else if sc.subscriberCount > 0 {
			sc.subscriberCount = sc.subscriberCount + -1
		}

		err = tx.Set(docRef, sc)
		if err != nil {
			return errors.Wrap(err, "firestore: failed to update subscription count")
		}

		return nil
	})

	if err != nil {
		client.logger.WithError(err).WithFields(field).Errorln("firestore: failed to update subscription count")
		return err
	}

	return nil
}

func (client Client) GetSubscriptionCount(ctx context.Context, topicName string) (int, error) {
	field := log.Fields{"collection": CollectionSubscriptionsCount, "topicName": topicName}
	ctx, span := makeFirestoreTrace(ctx, "GetSubscriptionCount", field, client.logger.Data)
	defer span.End()

	collection := client.fsClient.Collection(CollectionSubscriptionsCount)
	docRef := collection.Doc(topicName)
	docSnap, err := docRef.Get(ctx)
	if err != nil && status.Code(err) == codes.NotFound {
		return 0, nil
	}

	if err != nil {
		client.logger.WithError(err).WithFields(field).Errorln("firestore: failed to get subscription count")
		return 0, err
	}

	sc := SubscriptionCount{}
	err = docSnap.DataTo(&sc)
	if err != nil {
		client.logger.WithError(err).WithFields(field).Errorln("firestore: failed to map SubscriptionCount")
		return 0, err
	}

	return sc.subscriberCount, nil
}
