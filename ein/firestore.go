package ein

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"time"

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
		ein.logger.WithError(err).Fatalln("firestore: failed to write university.topicName")
	}

	ein.logger.Infoln("set university.topicName")

	ein.insertSubjects(&newUniversity)

	ein.insertSemester(&newUniversity)
}

type FirestoreSeason struct {
	CurrentSeason string `firestore:"currentSeason"`
	NextSeason    string `firestore:"nextSeason"`
	LastSeason    string `firestore:"lastSeason"`
}

func (ein *ein) insertSemester(university *model.University) {
	firestoreSeason := FirestoreSeason{
		makeSemesterKey(university.ResolvedSemesters.Current),
		makeSemesterKey(university.ResolvedSemesters.Next),
		makeSemesterKey(university.ResolvedSemesters.Last),
	}

	collections := ein.firestoreClient.Collection("university.topicName")
	docRef := collections.Doc(university.TopicName)
	_, err := docRef.Set(ein.ctx, firestoreSeason)
	if err != nil {
		ein.logger.Fatalln(err)
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
	collections := ein.firestoreClient.Collection("university.semester")
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

func (ein *ein) updateSerialSection(sections []*model.Section) {
	collection := ein.firestoreClient.Collection("section.topicName")

	Batch(500, sections, func(s []*model.Section) {
		batch := ein.firestoreClient.Batch()

		for sectionIndex := range s {
			section := s[sectionIndex]
			secData, err := section.Marshal()
			if err != nil {
				ein.logger.WithError(err).Fatalln("failed to marshal Section")
			}
			docRef := collection.Doc(section.TopicName)
			batch.Set(docRef, FirestoreData{secData})
		}

		results, err := batch.Commit(ein.ctx)
		if err != nil {
			ein.logger.Fatalln(err)
		}
		ein.logger.WithField("results", len(results)).Infoln("firestore: set sections")
	})
}

func Batch(count int, items []*model.Section, callback func([]*model.Section)) {
	var result []*model.Section

	for i := 0; i < len(items); i++ {
		if len(result) == count {
			callback(result)
			result = []*model.Section{}
		}

		result = append(result, items[i])
	}

	callback(items[len(items)-len(items)%count:])
}
