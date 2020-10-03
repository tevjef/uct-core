package ein

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/model"
)

func (ein *ein) insertUniversity(newUniversity model.University, university model.University) {
	defer model.TimeTrack(time.Now(), "insertUniversity")

	universityCopy := university
	universityCopy.Subjects = nil

	data, err := universityCopy.Marshal()
	if err != nil {
		ein.logger.WithError(err).Fatalln("failed to marshal University")
	}

	universityView := NewFirestoreData(data)
	collections := ein.firestoreClient.Collection("university.topicName")
	docRef := collections.Doc(universityCopy.TopicName)
	_, err = docRef.Set(ein.ctx, universityView)
	if err != nil {
		ein.logger.WithError(err).Fatalln("firestore: failed to set university.topicName")
	}

	ein.logger.Infoln("set university.topicName")

	ein.insertSubjects(&newUniversity)

	ein.insertSemester(&newUniversity)
}

type FirestoreSemesters struct {
	CurrentSeason string `firestore:"currentSeason"`
	NextSeason    string `firestore:"nextSeason"`
	LastSeason    string `firestore:"lastSeason"`
}

func (ein *ein) insertSemester(university *model.University) {
	firestoreSeason := FirestoreSemesters{
		makeSemesterKey(university.ResolvedSemesters.Current),
		makeSemesterKey(university.ResolvedSemesters.Next),
		makeSemesterKey(university.ResolvedSemesters.Last),
	}

	collections := ein.firestoreClient.Collection("university.semesters")
	docRef := collections.Doc(university.TopicName)
	_, err := docRef.Set(ein.ctx, firestoreSeason)
	if err != nil {
		ein.logger.WithError(err).Fatalln("firestore: failed to set university.semesters")
	}
}

type SubjectView struct {
	Season    string `firestore:"season"`
	Year      string `firestore:"year"`
	TopicName string `firestore:"topic"`
	Data      []byte `firestore:"data"`
}

type FirestoreData struct {
	Data []byte `firestore:"data"`
}

type SectionFirestoreData struct {
	Data       []byte                 `firestore:"data"`
	University *firestore.DocumentRef `firestore:"universityRef"`
	Year       string                 `firestore:"year"`
	Season     string                 `firestore:"season"`
}

func NewFirestoreData(data []byte) FirestoreData {
	return FirestoreData{data}
}

func (ein *ein) insertSubjects(university *model.University) {
	ein.injectSubjectBySemester(university, university.ResolvedSemesters.Current)
	ein.injectSubjectBySemester(university, university.ResolvedSemesters.Next)
	ein.injectSubjectBySemester(university, university.ResolvedSemesters.Last)

	collection := ein.firestoreClient.Collection("subject.topicName")
	batch := ein.firestoreClient.Batch()
	for subjectIndex := range university.Subjects {
		subject := university.Subjects[subjectIndex]
		docRef := collection.Doc(subject.TopicName)

		data, _ := subject.Marshal()
		subView := NewFirestoreData(data)

		batch.Set(docRef, subView)
	}
	results, err := batch.Commit(ein.ctx)
	if err != nil {
		ein.logger.WithError(err).Fatalln("firestore: failed to commit subject transaction")
	}

	ein.logger.WithField("results", len(results)).Infoln("firestore: set subject.topicName")
}

func (ein *ein) injectSubjectBySemester(university *model.University, semester *model.Semester) {
	field := log.Fields{"semester": makeSemesterKey(semester)}
	filteredSubjects := getSubjectsForSemester(university.Subjects, semester)

	data, err := (&model.Data{Subjects: filteredSubjects}).Marshal()
	if err != nil {
		ein.logger.WithError(err).WithFields(field).Fatalln("failed to marshal Data{[]Subject}")
	}

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err = zw.Write(data)
	if err != nil {
		ein.logger.WithError(err).WithFields(field).Fatalln("failed to gzip subjects")
	}
	err = zw.Close()
	if err != nil {
		ein.logger.WithError(err).WithFields(field).Fatalln(err)
	}

	NewFirestoreData(data)

	firestoreData := FirestoreData{Data: buf.Bytes()}
	collections := ein.firestoreClient.Collection("university.subjects")
	docRef := collections.Doc(university.TopicName + "." + makeSemesterKey(semester))
	_, err = docRef.Set(ein.ctx, firestoreData)
	if err != nil {
		ein.logger.WithError(err).WithFields(field).Fatalln("firestore: failed to set university.semester")
	}

	ein.logger.WithFields(field).Infoln("firestore: set university.semester")
}

func makeSemesterKey(semester *model.Semester) string {
	return semester.Season + fmt.Sprint(semester.Year)
}

func getSubjectsForSemester(subjects []*model.Subject, semester *model.Semester) []*model.Subject {
	var current []*model.Subject

	for subjectIndex := range subjects {
		subject := subjects[subjectIndex]

		if subject.Year == fmt.Sprint(semester.Year) && subject.Season == semester.Season {
			current = append(current, subject)
		}
	}

	return current
}

func (ein *ein) insertSubject(sub *model.Subject) {
	field := log.Fields{"subject": sub.TopicName}

	if !ein.config.fullUpsert {
		// Update all
		return
	}

	subCopy := *sub

	data, err := subCopy.Marshal()
	if err != nil {
		ein.logger.WithError(err).WithFields(field).Fatalln("failed to marshal Subject")
	}

	subView := NewFirestoreData(data)

	collections := ein.firestoreClient.Collection("subject.topicName")
	docRef := collections.Doc(subCopy.TopicName)
	results, err := docRef.Set(ein.ctx, subView)
	if err != nil {
		ein.logger.WithFields(field).Fatalln(err)
	}

	ein.logger.WithFields(field).WithField("result", fmt.Sprintf("%+v", results)).WithField("topicName", sub.TopicName).Debugln("firestore: set subject.topicName")
}

func (ein *ein) updateSerialSection(sectionMeta []SectionMeta) {
	collection := ein.firestoreClient.Collection("section.topicName")
	universityCollection := ein.firestoreClient.Collection("university.topicName")

	Batch(500, sectionMeta, func(s []SectionMeta) {
		batch := ein.firestoreClient.Batch()

		for sectionIndex := range s {
			sectionMeta := s[sectionIndex]
			secData, err := sectionMeta.section.Marshal()
			if err != nil {
				ein.logger.WithError(err).Fatalln("failed to marshal Section")
			}
			docRef := collection.Doc(sectionMeta.section.TopicName)
			batch.Set(docRef, SectionFirestoreData{
				secData,
				universityCollection.Doc(sectionMeta.university.TopicName),
				sectionMeta.subject.Year,
				sectionMeta.subject.Season,
			})
		}

		results, err := batch.Commit(ein.ctx)
		if err != nil {
			ein.logger.Fatalln(err)
		}
		ein.logger.WithField("results", len(results)).Infoln("firestore: set sections")
	})
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
