package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	"io"
	"io/ioutil"
	"strconv"
)

func MarshalMessage(format string, m University) (*bytes.Reader, error) {
	var out []byte
	var err error
	if format == JSON {
		out, err = json.Marshal(m)
		if err != nil {
			return nil, errors.Wrap(err, "failed to encode message")
		}
	} else if format == PROTOBUF {
		out, err = m.Marshal()
		if err != nil {
			return nil, errors.Wrap(err, "failed to encode message")
		}
	}
	return bytes.NewReader(out), nil
}

func UnmarshallMessage(format string, r io.Reader, m *University) error {
	if format == JSON {
		dec := ffjson.NewDecoder()
		if err := dec.DecodeReader(r, &*m); err != nil {
			return err
		}
	} else if format == PROTOBUF {
		data, err := ioutil.ReadAll(r)
		if err = m.Unmarshal(data); err != nil {
			return err
		}
	}
	if m.Equal(University{}) {
		return fmt.Errorf("%s Reason %s", "Failed to unmarshal message:", "empty data")
	}
	return nil
}

func CheckUniqueSubject(subjects []*Subject) {
	m := make(map[string]int)
	for subjectIndex := range subjects {
		subject := subjects[subjectIndex]
		key := subject.Season + subject.Year + subject.Name + subject.Number
		m[key]++
		if m[key] > 1 {
			log.WithFields(log.Fields{"key": key, "count": m[key]}).Debugln("Duplicate subject")
			subject.Name = subject.Name + "_" + strconv.Itoa(m[key])
		}
	}
}

func CheckUniqueCourse(subject *Subject, courses []*Course) {
	m := map[string]int{}
	for courseIndex := range courses {
		course := courses[courseIndex]
		key := course.Name + course.Number
		m[key]++
		if m[key] > 1 {
			log.WithFields(log.Fields{"subject": subject.Name,
				"season": subject.Season,
				"year":   subject.Year,
				"key":    key,
				"count":  m[key]}).Debugln("Duplicate course")
			course.Name = course.Name + "_" + strconv.Itoa(m[key])
		}
	}
}

func ValidateAll(uni *University) (err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in ValidateAll", r)
			err = fmt.Errorf("%v+", r)
		}
	}()

	uni.Validate()
	CheckUniqueSubject(uni.Subjects)
	for subjectIndex := range uni.Subjects {
		subject := uni.Subjects[subjectIndex]
		subject.Validate(uni)

		courses := subject.Courses
		CheckUniqueCourse(subject, courses)
		for courseIndex := range courses {
			course := courses[courseIndex]
			course.Validate(subject)

			sections := course.Sections
			for sectionIndex := range sections {
				section := sections[sectionIndex]
				section.Validate(course)

				//[]Instructors
				instructors := section.Instructors
				for instructorIndex := range instructors {
					instructor := instructors[instructorIndex]
					instructor.Index = int32(instructorIndex)
					instructor.Validate()
				}

				//[]Meeting
				meetings := section.Meetings
				for meetingIndex := range meetings {
					meeting := meetings[meetingIndex]
					meeting.Index = int32(meetingIndex)
					meeting.Validate()

					// Meeting []Metadata
					metadatas := meeting.Metadata
					for metadataIndex := range metadatas {
						metadata := metadatas[metadataIndex]
						metadata.Validate()
					}
				}

				//[]Books
				books := section.Books
				for bookIndex := range books {
					book := books[bookIndex]
					book.Validate()
				}

				// Section []Metadata
				metadatas := section.Metadata
				for metadataIndex := range metadatas {
					metadata := metadatas[metadataIndex]
					metadata.Validate()
				}
			}

			// Course []Metadata
			metadatas := course.Metadata
			for metadataIndex := range metadatas {
				metadata := metadatas[metadataIndex]
				metadata.Validate()
			}
		}
	}

	for registrations := range uni.Registrations {
		_ = uni.Registrations[registrations]

	}

	// university []Metadata
	metadatas := uni.Metadata
	for metadataIndex := range metadatas {
		metadata := metadatas[metadataIndex]
		metadata.Validate()

	}

	return nil
}
