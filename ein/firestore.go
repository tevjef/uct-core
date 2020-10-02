package ein

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/model"
)

func (ein *ein) insertUniversity(newUniversity model.University, university model.University) {

	defer model.TimeTrack(time.Now(), "insertUniversity")

	data, err := university.Marshal()

	university.Subjects = nil
	universityView := FirestoreData{data}
	collections := ein.firestoreClient.Collection("university.topicName")
	docRef := collections.Doc(university.TopicName)
	_, err = docRef.Set(ein.ctx, universityView)
	if err != nil {
		log.Fatalln(err)
	}

	ein.insertSubjects(&university)

	ein.insertSemester(&university)
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
		log.Fatalln(err)
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

func (ein *ein) insertSubjects(university *model.University) {
	ein.injectSubjectBySemester(university, university.ResolvedSemesters.Current)
	ein.injectSubjectBySemester(university, university.ResolvedSemesters.Next)
	ein.injectSubjectBySemester(university, university.ResolvedSemesters.Last)

	for subjectIndex := range university.Subjects {
		subject := university.Subjects[subjectIndex]

		ein.insertSubject(subject)
	}
}

func (ein *ein) injectSubjectBySemester(university *model.University, semester *model.Semester) {
	filteredSubjects := getSubjectsForSemester(university.Subjects, semester)

	data, err := (&model.Data{Subjects: filteredSubjects}).Marshal()
	if err != nil {
		log.Fatalln(err)
	}

	firestoreData := FirestoreData{Data: data}
	collections := ein.firestoreClient.Collection("university.topicName")
	docRef := collections.Doc(university.TopicName + "." + makeSemesterKey(university.ResolvedSemesters.Current))
	_, err = docRef.Set(ein.ctx, firestoreData)
	if err != nil {
		log.Fatalln(err)
	}
}

func makeSemesterKey(semester *model.Semester) string {
	return semester.Season + string(semester.Year)
}

func getSubjectsForSemester(subjects []*model.Subject, semester *model.Semester) []*model.Subject {
	var current []*model.Subject

	for subjectIndex := range subjects {
		subject := subjects[subjectIndex]

		if subject.Year == string(semester.Year) && subject.Season == semester.Season {
			current = append(current, subject)
		}
	}

	return current
}

func (ein *ein) insertSubject(sub *model.Subject) {
	if !ein.config.fullUpsert {
		// Update all
		return
	}

	subCopy := *sub

	data, _ := subCopy.Marshal()

	subView := FirestoreData{
		Data: data,
	}

	collections := ein.firestoreClient.Collection("subject.topicName")
	docRef := collections.Doc(subCopy.TopicName)
	_, err := docRef.Set(ein.ctx, subView)
	if err != nil {
		log.Fatalln(err)
	}
}
