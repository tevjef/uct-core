package uctfirestore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tevjef/uct-backend/common/model"
)

const key = "firestoreclient"

// Setter defines a context that enables setting values.
type Setter interface {
	Set(string, interface{})
}

func FromContext(ctx context.Context) *firestore.Client {
	return ctx.Value(key).(*firestore.Client)
}

func ToContext(s Setter, client *firestore.Client) {
	s.Set(key, client)
}

func Firestore(client *firestore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ToContext(c, client)
		c.Next()
	}
}

const CollectionUniversityTopicName = "university.topicName"
const CollectionUniversitySemesters = "university.semesters"
const CollectionUniversitySubjects = "university.subjects"
const CollectionSubjectTopicName = "subject.topicName"
const CollectionSectionTopicName = "section.topicName"
const CollectionNotificationSent = "notification.sent"

type FirestoreSemesters struct {
	CurrentSeason string `firestore:"currentSeason"`
	NextSeason    string `firestore:"nextSeason"`
	LastSeason    string `firestore:"lastSeason"`
}

type FirestoreData struct {
	Data []byte `firestore:"data"`
}

type SectionFirestoreData struct {
	Data                []byte `firestore:"data"`
	UniversityTopicName string `firestore:"university"`
	CourseName          string `firestore:"courseName"`
	Year                string `firestore:"year"`
	Season              string `firestore:"season"`
}

type SectionMeta struct {
	Section    *model.Section
	Subject    *model.Subject
	Course     *model.Course
	University model.University
}

type SectionNotification struct {
	Section    *model.Section
	University *model.University
	CourseName string
}

// FirestoreEvent is the payload of a Firestore event.
type FirestoreEvent struct {
	OldValue   FirestoreValue `json:"oldValue"`
	Value      FirestoreValue `json:"value"`
	UpdateMask struct {
		FieldPaths []string `json:"fieldPaths"`
	} `json:"updateMask"`
}

// FirestoreValue holds Firestore fields.
type FirestoreValue struct {
	CreateTime time.Time `json:"createTime"`
	// Fields is the data for this value. The type depends on the format of your
	// database. Log the interface{} value and inspect the result to see a JSON
	// representation of your database fields.
	Fields     interface{} `json:"fields"`
	Name       string      `json:"name"`
	UpdateTime time.Time   `json:"updateTime"`
}

type SectionFirestoreValue struct {
	Data                Value `json:"data"`
	UniversityTopicName Value `json:"university"`
	CourseName          Value `json:"courseName"`
	Year                Value `json:"year"`
	Season              Value `json:"season"`
}

func FromFirestoreValue(value FirestoreValue) (*SectionFirestoreData, error) {
	data := SectionFirestoreValue{}
	b, err := json.Marshal(value.Fields)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to marshal FireStoreValue: %+v", value))

	}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal FireStoreValue: %+v", value))
	}

	return data.ToSectionFirestoreData(), nil
}

func (v SectionFirestoreValue) ToSectionFirestoreData() *SectionFirestoreData {
	return &SectionFirestoreData{
		Data:                v.Data.BytesValue,
		UniversityTopicName: v.Data.StringValue,
		CourseName:          v.Data.StringValue,
		Year:                v.Year.StringValue,
		Season:              v.Season.StringValue,
	}
}

type Value struct {
	StringValue    string `json:"stringValue"`
	BytesValue     []byte `json:"bytesValue"`
	ReferenceValue string `json:"referenceValue"`
	IntegerValue   int    `json:"integerValue"`
}

type Client struct {
	fsClient *firestore.Client
	logger   *log.Entry

	context context.Context
}

func NewClient(context context.Context, fsClient *firestore.Client, logger *log.Entry) *Client {
	return &Client{
		fsClient,
		logger,
		context,
	}
}
