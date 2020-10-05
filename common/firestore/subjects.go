package uctfirestore

import (
	"bytes"
	"compress/gzip"
	"context"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/model"
)

func (client Client) InsertSubjects(ctx context.Context, subjects []*model.Subject) error {
	field := log.Fields{"collection": CollectionSubjectTopicName, "subjects": len(subjects)}
	ctx, span := makeFirestoreTrace(ctx, "InsertSubjects", field, client.logger.Data)
	defer span.End()

	collection := client.fsClient.Collection(CollectionSubjectTopicName)
	batch := client.fsClient.Batch()
	for subjectIndex := range subjects {
		subject := subjects[subjectIndex]
		docRef := collection.Doc(subject.TopicName)

		data, _ := subject.Marshal()
		firestoreData := FirestoreData{Data: data}
		batch.Set(docRef, firestoreData)
	}

	results, err := batch.Commit(ctx)
	if err != nil {
		client.logger.WithError(err).WithFields(field).Fatalln("firestore: failed to commit subject transaction")
	}

	client.logger.WithField("results", len(results)).
		WithFields(field).
		Debugln("firestore: subjects batch set complete")

	return nil
}

func (client Client) GetSubject(ctx context.Context, topicName string) (*model.Subject, error) {
	field := log.Fields{"topicName": topicName, "collection": CollectionSubjectTopicName}
	ctx, span := makeFirestoreTrace(ctx, "GetSubject", field, client.logger.Data)
	defer span.End()

	collection := client.fsClient.Collection(CollectionSubjectTopicName)
	docRef := collection.Doc(topicName)

	docSnap, err := docRef.Get(ctx)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalf("firestore: failed to get docRef")
		return nil, err
	}

	firestoreData := FirestoreData{}
	err = docSnap.DataTo(&firestoreData)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalf("firestore: failed to map FirestoreData")
		return nil, err
	}

	subject := &model.Subject{}

	err = subject.Unmarshal(firestoreData.Data)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalf("firestore: failed to unmarshal model.Subject")
		return nil, err
	}

	return subject, nil
}

func (client Client) InsertSubjectsBySemester(ctx context.Context, university model.University, semester *model.Semester) error {
	field := log.Fields{"semester": MakeSemesterKey(semester), "subjects": len(university.Subjects)}
	ctx, span := makeFirestoreTrace(ctx, "InsertSubjectsBySemester", field, client.logger.Data)
	defer span.End()

	filteredSubjects := getSubjectsForSemester(university.Subjects, semester)

	data, err := (&model.Data{Subjects: filteredSubjects}).Marshal()
	if err != nil {
		client.logger.WithError(err).WithFields(field).Fatalln("failed to marshal Data{[]Subject}")
		return err
	}

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err = zw.Write(data)
	if err != nil {
		client.logger.WithError(err).WithFields(field).Fatalln("firestore:  failed to gzip subjects")
		return err
	}
	err = zw.Close()
	if err != nil {
		client.logger.WithError(err).WithFields(field).Fatalln("firestore: failed to close gzip writer")
		return err
	}

	firestoreData := FirestoreData{Data: buf.Bytes()}
	collections := client.fsClient.Collection(CollectionUniversitySubjects)
	docRef := collections.Doc(university.TopicName + "." + MakeSemesterKey(semester))
	_, err = docRef.Set(ctx, firestoreData)
	if err != nil {
		client.logger.WithError(err).WithFields(field).Fatalf("firestore: failed to set %s", CollectionUniversitySubjects)
	}

	client.logger.WithFields(field).Debugln("firestore: set %s", CollectionUniversitySubjects)

	return nil
}

func (client Client) GetSubjectsBySemester(ctx context.Context, universityTopicName string, semester *model.Semester) ([]*model.Subject, error) {
	field := log.Fields{"semester": MakeSemesterKey(semester), "collection": CollectionUniversitySubjects}
	ctx, span := makeFirestoreTrace(ctx, "GetSubjectsBySemester", field, client.logger.Data)
	defer span.End()

	collections := client.fsClient.Collection(CollectionUniversitySubjects)
	docRef := collections.Doc(universityTopicName + "." + MakeSemesterKey(semester))
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalln("firestore: failed to get docRef")
	}

	data := FirestoreData{}
	err = docSnap.DataTo(&data)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalln("firestore: failed to map to FirestoreData")
		return nil, err
	}

	reader, err := gzip.NewReader(bytes.NewReader(data.Data))
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalln("firestore: failed to unzip FirestoreData")
		return nil, err
	}

	uctDataBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalln("firestore: failed to unzip FirestoreData")
		return nil, err
	}

	err = reader.Close()
	if err != nil {
		client.logger.WithError(err).WithFields(field).WithField("path", docSnap.Ref.Path).Fatalln("firestore: failed to close gzip reader")
		return nil, err
	}

	uctData := &model.Data{}
	err = uctData.Unmarshal(uctDataBytes)
	if err != nil {
		client.logger.WithError(err).WithFields(field).Fatalln("firestore: failed to unmarshal Data{[]Subject}")
		return nil, err
	}

	return uctData.Subjects, nil
}
