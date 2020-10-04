package uctfirestore

import (
	log "github.com/sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/model"
)

func (client Client) InsertSection(sectionMetas []SectionMeta) error {
	field := log.Fields{"collection": CollectionSectionTopicName}

	collection := client.fsClient.Collection(CollectionSectionTopicName)

	Batch(500, sectionMetas, func(s []SectionMeta) {

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

		results, err := batch.Commit(client.context)
		if err != nil {
			client.logger.Fatalln(err)
		}
		client.logger.WithField("results", len(results)).WithFields(field).Debugln("firestore: batch set complete")
	})

	return nil
}

func (client Client) GetSection(topicName string) (*model.Section, error) {
	field := log.Fields{"collection": CollectionSectionTopicName}

	collection := client.fsClient.Collection(CollectionSectionTopicName)
	docRef := collection.Doc(topicName)
	docSnap, err := docRef.Get(client.context)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalln("firestore: failed to get docRef")
	}

	sectionFirebaseData := SectionFirestoreData{}

	err = docSnap.DataTo(&sectionFirebaseData)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalf("firestore: failed to map SectionFirestoreData")
		return nil, err
	}

	section, err := SectionFromBytes(sectionFirebaseData.Data)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalf("firestore: failed to unmarshal model.Section")
		return nil, err
	}

	return section, nil
}

func (client Client) GetSectionNotification(topicName string) (*SectionNotification, error) {
	field := log.Fields{"collection": CollectionSectionTopicName}

	collection := client.fsClient.Collection(CollectionSectionTopicName)
	docRef := collection.Doc(topicName)
	docSnap, err := docRef.Get(client.context)
	if err != nil {
		client.logger.WithError(err).WithFields(field).Fatalln("firestore: failed to get docRef")
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

	university, err := client.GetUniversity(sectionFirebaseData.UniversityTopicName)
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

func Batch(count int, items []SectionMeta, callback func([]SectionMeta)) {
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
