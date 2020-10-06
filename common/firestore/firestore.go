package uctfirestore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/model"
	"go.opencensus.io/trace"
)

const key = "uctfirestoreclient"

// Setter defines a context that enables setting values.
type Setter interface {
	Set(string, interface{})
}

func FromContext(ctx context.Context) *Client {
	client := ctx.Value(key).(*Client)
	client.context = ctx
	return client
}

func ToContext(s Setter, client *Client) {
	s.Set(key, client)
}

func Firestore(client *Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = c.Request.Clone(context.WithValue(c.Request.Context(), key, client))
		c.Next()
	}
}

const CollectionUniversityTopicName = "university.topicName"
const CollectionUniversitySemesters = "university.semesters"
const CollectionUniversitySubjects = "university.subjects"
const CollectionSubjectTopicName = "subject.topicName"
const CollectionSectionTopicName = "section.topicName"
const CollectionCourseTopicName = "course.topicName"
const CollectionNotificationSent = "notification.sent"
const CollectionNotificationReceived = "notification.received"
const CollectionSubscriptionUser = "subscription.user"
const CollectionSubscriptionCount = "subscription.count"
const CollectionCourseHotness = "course.hotness"

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

func makeFirestoreTrace(ctx context.Context, name string, attributes ...map[string]interface{}) (context.Context, *trace.Span) {
	newContext, span := trace.StartSpan(ctx, "/firestore."+name)
	for i := range attributes {
		for k, v := range attributes[i] {
			span.AddAttributes(trace.StringAttribute(k, fmt.Sprintf("%v", v)))
		}
	}
	return newContext, span
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
		UniversityTopicName: v.UniversityTopicName.StringValue,
		CourseName:          v.CourseName.StringValue,
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
		fsClient: fsClient,
		logger:   logger,
		context:  context,
	}
}
