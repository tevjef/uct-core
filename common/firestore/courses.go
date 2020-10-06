package uctfirestore

import (
	"context"
	"sort"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/model"
)

type CourseHotness struct {
	View []*FireStoreSubscriptionView `firestore:"view"`
}

type FireStoreSubscriptionView struct {
	TopicName   string `firestore:"topic_name"`
	Subscribers int64  `firestore:"subscribers"`
	IsHot       bool   `firestore:"is_hot"`
}

func (client Client) GetCourses(ctx context.Context, topicName string) ([]*model.Course, error) {
	field := log.Fields{"topicName": topicName}
	ctx, span := makeFirestoreTrace(ctx, "GetCourses", field, client.logger.Data)
	defer span.End()

	subject, err := client.GetSubject(ctx, topicName)
	if err != nil {
		return nil, err
	}

	sort.Sort(model.CourseByNumber{Courses: subject.Courses})

	return subject.Courses, nil
}

func (client Client) InsertCourses(ctx context.Context, courses []*model.Course) error {
	field := log.Fields{"collection": CollectionCourseTopicName, "courses": len(courses)}
	ctx, span := makeFirestoreTrace(ctx, "InsertCourses", field, client.logger.Data)
	defer span.End()

	collection := client.fsClient.Collection(CollectionCourseTopicName)

	BatchCourses(500, courses, func(c []*model.Course) {

		batch := client.fsClient.Batch()

		for courseIndex := range c {
			course := c[courseIndex]
			courseBytes, err := course.Marshal()
			if err != nil {
				client.logger.WithError(err).WithFields(field).Fatalln("firestore: failed to marshal Course")
			}
			docRef := collection.Doc(course.TopicName)
			batch.Set(docRef, FirestoreData{
				Data: courseBytes,
			})
		}

		results, err := batch.Commit(ctx)
		if err != nil {
			client.logger.Fatalln(err)
		}
		client.logger.WithField("results", len(results)).WithFields(field).Debugln("firestore: courses batch set complete")
	})

	return nil
}

func (client Client) GetCourse(ctx context.Context, topicName string) (*model.Course, error) {
	field := log.Fields{"collection": CollectionCourseTopicName}
	ctx, span := makeFirestoreTrace(ctx, "GetCourse", field, client.logger.Data)
	defer span.End()

	collection := client.fsClient.Collection(CollectionCourseTopicName)
	docRef := collection.Doc(topicName)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalln("firestore: failed to get docRef")
	}

	firebaseData := FirestoreData{}

	err = docSnap.DataTo(&firebaseData)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalf("firestore: failed to map FirestoreData")
		return nil, err
	}

	course := &model.Course{}

	err = course.Unmarshal(firebaseData.Data)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalf("firestore: failed to unmarshal model.Course")
		return nil, err
	}

	return course, nil
}

func BatchCourses(count int, items []*model.Course, callback func([]*model.Course)) {
	var result []*model.Course

	for i := 0; i < len(items); i++ {
		if len(result) == count {
			callback(result)
			result = []*model.Course{}
		}

		result = append(result, items[i])
	}

	callback(items[len(items)-len(items)%count:])
}

func (client Client) SetCourseHotness(ctx context.Context, courseTopicName string, courseHotness *CourseHotness) error {
	field := log.Fields{"collection": CollectionCourseHotness}
	ctx, span := makeFirestoreTrace(ctx, "SetCourseHotness", field, client.logger.Data)
	defer span.End()

	collection := client.fsClient.Collection(CollectionCourseHotness)

	docRef := collection.Doc(courseTopicName)
	result, err := docRef.Set(ctx, courseHotness)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("result", result).Errorln("firestore: failed to set CourseHotness")
		return errors.Wrap(err, "firestore: failed to set CourseHotness")
	}

	return nil
}

func (client Client) GetCourseHotness(ctx context.Context, courseTopicName string) (*CourseHotness, *time.Time, error) {
	field := log.Fields{"collection": CollectionCourseHotness}
	ctx, span := makeFirestoreTrace(ctx, "GetCourseHotness", field, client.logger.Data)
	defer span.End()

	collection := client.fsClient.Collection(CollectionCourseHotness)

	docRef := collection.Doc(courseTopicName)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Errorln("firestore: failed to get docRef")
		return nil, nil, errors.Wrap(err, "firestore: failed to get docRef")
	}

	sc := CourseHotness{}

	err = docSnap.DataTo(&sc)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Errorln("firestore: failed to map CourseHotness")
		return nil, nil, errors.Wrap(err, "firestore: failed to map CourseHotness")
	}

	return &sc, &docSnap.UpdateTime, nil
}
