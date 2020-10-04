package uctfirestore

import (
	log "github.com/sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/model"
)

func (client Client) InsertUniversity(university model.University) error {
	universityCopy := university
	universityCopy.Subjects = nil

	data, err := universityCopy.Marshal()
	if err != nil {
		client.logger.WithError(err).Fatalln("failed to marshal University")
	}

	universityView := FirestoreData{data}
	collections := client.fsClient.Collection(CollectionUniversityTopicName)
	docRef := collections.Doc(universityCopy.TopicName)
	_, err = docRef.Set(client.context, universityView)
	if err != nil {
		client.logger.WithError(err).Fatalln("firestore: failed to set university.topicName")
		return err
	}

	client.logger.Debugf("%v set university.topicName", university.TopicName)

	return nil
}

func (client Client) GetUniversity(topicName string) (university *model.University, err error) {
	field := log.Fields{"collection": CollectionUniversityTopicName}

	collections := client.fsClient.Collection(CollectionUniversityTopicName)
	docRef := collections.Doc(topicName)
	docSnap, err := docRef.Get(client.context)
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

func (client Client) GetUniversities() (universities []*model.University, err error) {
	field := log.Fields{"collection": CollectionUniversityTopicName}

	collections := client.fsClient.Collection(CollectionUniversityTopicName)
	docRef := collections.Documents(client.context)
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
