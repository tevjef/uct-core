package uctfirestore

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/model"
)

func (client Client) InsertSections(ctx context.Context, sectionMetas []SectionMeta) error {
	field := log.Fields{"collection": CollectionSectionTopicName, "sections": len(sectionMetas)}
	ctx, span := makeFirestoreTrace(ctx, "InsertSections", field, client.logger.Data)
	defer span.End()

	collection := client.fsClient.Collection(CollectionSectionTopicName)

	BatchSections(500, sectionMetas, func(s []SectionMeta) {
		batch := client.fsClient.Batch()

		for sectionIndex := range s {
			sectionMeta := s[sectionIndex]
			secData, err := sectionMeta.Section.Marshal()
			if err != nil {
				client.logger.WithError(err).WithFields(field).Fatalln("firestore: failed to marshal Section")
			}
			docRef := collection.Doc(sectionMeta.Section.TopicName)
			batch.Set(docRef, SectionFirestoreData{
				Data:                secData,
				CourseName:          sectionMeta.Course.Name,
				UniversityTopicName: sectionMeta.University.TopicName,
				Year:                sectionMeta.Subject.Year,
				Season:              sectionMeta.Subject.Season,
			})
		}

		results, err := batch.Commit(ctx)
		if err != nil {
			client.logger.Fatalln(err)
		}
		client.logger.WithField("results", len(results)).WithFields(field).Debugln("firestore: section batch set complete")
	})

	return nil
}

func (client Client) GetSection(ctx context.Context, topicName string) (*model.Section, error) {
	field := log.Fields{"collection": CollectionSectionTopicName, "topicName": topicName}
	ctx, span := makeFirestoreTrace(ctx, "GetSection", field, client.logger.Data)
	defer span.End()

	collection := client.fsClient.Collection(CollectionSectionTopicName)
	docRef := collection.Doc(topicName)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Errorln("firestore: failed to get docRef")
		return nil, err
	}

	sectionFirebaseData := SectionFirestoreData{}

	err = docSnap.DataTo(&sectionFirebaseData)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Errorln("firestore: failed to map SectionFirestoreData")
		return nil, err
	}

	section, err := SectionFromBytes(sectionFirebaseData.Data)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Errorln("firestore: failed to unmarshal model.Section")
		return nil, err
	}

	return section, nil
}

func (client Client) GetSectionNotification(ctx context.Context, topicName string) (*SectionNotification, error) {
	field := log.Fields{"collection": CollectionSectionTopicName, "topicName": topicName}
	ctx, span := makeFirestoreTrace(ctx, "GetSectionNotification", field, client.logger.Data)
	defer span.End()

	collection := client.fsClient.Collection(CollectionSectionTopicName)
	docRef := collection.Doc(topicName)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		client.logger.WithError(err).WithFields(field).Errorln("firestore: failed to get docRef")
		return nil, err
	}

	sectionFirebaseData := SectionFirestoreData{}

	err = docSnap.DataTo(&sectionFirebaseData)
	if err != nil {
		client.logger.WithError(err).WithFields(field).Fatalf("firestore: failed to map SectionFirestoreData")
		return nil, err
	}

	section, err := SectionFromBytes(sectionFirebaseData.Data)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalf("firestore: failed to unmarshal model.Section")
		return nil, err
	}

	university, err := client.GetUniversity(ctx, sectionFirebaseData.UniversityTopicName)
	if err != nil {
		client.logger.WithError(err).WithFields(field).Fatalf("firestore: failed to get section university")
		return nil, err
	}

	sectionNotification := &SectionNotification{
		Section:    section,
		University: university,
		CourseName: sectionFirebaseData.CourseName,
	}
	return sectionNotification, nil
}

func SectionFromBytes(bytes []byte) (*model.Section, error) {
	section := &model.Section{}
	err := section.Unmarshal(bytes)
	return section, err
}

func BatchSections(count int, items []SectionMeta, callback func([]SectionMeta)) {
	var result []SectionMeta

	for i := 0; i < len(items); i++ {
		if len(result) == count {
			callback(result)
			result = []SectionMeta{}
		}

		result = append(result, items[i])
	}

	callback(items[len(items)-len(items)%count:])
}
