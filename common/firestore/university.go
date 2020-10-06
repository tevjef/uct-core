package uctfirestore

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/model"
	"go.opencensus.io/trace"
)

func (client Client) InsertUniversity(ctx context.Context, university model.University) error {
	field := log.Fields{"collection": CollectionUniversityTopicName, "topicName": university.TopicName}
	ctx, span := makeFirestoreTrace(ctx, "InsertUniversity", field, client.logger.Data)
	defer span.End()

	universityCopy := university
	universityCopy.Subjects = nil

	data, err := universityCopy.Marshal()
	if err != nil {
		client.logger.WithError(err).Fatalln("failed to marshal University")
	}

	universityView := FirestoreData{data}
	collections := client.fsClient.Collection(CollectionUniversityTopicName)
	docRef := collections.Doc(universityCopy.TopicName)
	_, err = docRef.Set(ctx, universityView)
	if err != nil {
		client.logger.WithError(err).Fatalln("firestore: failed to set university.topicName")
		return err
	}

	client.logger.WithFields(field).Debugf("firestore: set %v", CollectionUniversityTopicName)

	return nil
}

func (client Client) GetUniversity(ctx context.Context, topicName string) (university *model.University, err error) {
	field := log.Fields{"collection": CollectionUniversityTopicName, "topicName": topicName}
	ctx, span := makeFirestoreTrace(ctx, "GetUniversity", field, client.logger.Data)
	defer span.End()

	collections := client.fsClient.Collection(CollectionUniversityTopicName)
	docRef := collections.Doc(topicName)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalln("firestore: failed to get docRef")
		return nil, err
	}

	firestoreData := FirestoreData{}
	err = docSnap.DataTo(&firestoreData)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalln("firestore: failed to get doc ref")
		return
	}

	uni := &model.University{}

	err = uni.Unmarshal(firestoreData.Data)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalln("firestore: failed to get doc ref")
		return
	}

	semesters := []*model.Semester{
		uni.ResolvedSemesters.Current,
		uni.ResolvedSemesters.Next,
		uni.ResolvedSemesters.Last,
	}

	uni.AvailableSemesters = semesters

	return uni, err
}

func (client Client) GetUniversities(ctx context.Context) (universities []*model.University, err error) {
	field := log.Fields{"collection": CollectionUniversityTopicName}
	ctx, span := makeFirestoreTrace(ctx, "GetUniversities", field, client.logger.Data)
	defer span.AddAttributes(trace.Int64Attribute("count", int64(len(universities))))
	defer span.End()

	collections := client.fsClient.Collection(CollectionUniversityTopicName)
	docRef := collections.Documents(ctx)
	docSnaps, err := docRef.GetAll()
	if err != nil {
		client.logger.WithError(err).WithFields(field).Fatalln("firestore: failed to get all docRefs")
		return nil, err
	}

	for i := range docSnaps {
		docSnap := docSnaps[i]
		firestoreData := FirestoreData{}
		err := docSnap.DataTo(&firestoreData)
		if err != nil {
			client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalln("firestore: failed to get doc ref")
			return nil, err
		}

		uni := &model.University{}

		err = uni.Unmarshal(firestoreData.Data)
		if err != nil {
			client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalln("firestore: failed to get doc ref")
			return nil, err
		}

		semesters := []*model.Semester{
			uni.ResolvedSemesters.Current,
			uni.ResolvedSemesters.Next,
			uni.ResolvedSemesters.Last,
		}

		uni.AvailableSemesters = semesters

		universities = append(universities, uni)
	}

	return universities, err
}
