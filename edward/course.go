package main

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gonum/stat"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/tevjef/uct-backend/common/firestore"
	"github.com/tevjef/uct-backend/common/middleware"
	"github.com/tevjef/uct-backend/common/middleware/httperror"
	mtrace "github.com/tevjef/uct-backend/common/middleware/trace"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/spike/store"
)

type FireStoreSubscriptionDocument struct {
	View []*FireStoreSubscriptionView `firestore:"view"`
}

type FireStoreSubscriptionView struct {
	TopicName   string `firestore:"topic_name"`
	Subscribers int64  `firestore:"subscribers"`
	IsHot       bool   `firestore:"is_hot"`
}

func courseHandler(c *gin.Context) {
	defer model.TimeTrack(time.Now(), "courseHandler")
	ctx, cancel := context.WithCancel(c)
	defer cancel()

	courseTopicName := strings.ToLower(c.Param("topic"))

	if course, _, err := SelectCourse(ctx, courseTopicName); err != nil {
		if err == sql.ErrNoRows {
			httperror.NotFound(c, err)
			return
		}
		httperror.ServerError(c, err)
		return
	} else {
		var subs []*model.SubscriptionView

		firestoreClient := firestore.FromContext(c)
		hotnessRef := firestoreClient.Collection("course.hotness")
		courseRef := hotnessRef.Doc(courseTopicName)

		var lastUpdate time.Time

		// Check if exists
		if documentSnapshot, _ := courseRef.Get(c); documentSnapshot.Exists() {
			lastUpdate = documentSnapshot.UpdateTime
			data := documentSnapshot.Data()["view"]
			if bytes, err := ffjson.Marshal(data); err != nil {
				httperror.ServerError(c, err)
				return
			} else if err := ffjson.Unmarshal(bytes, &subs); err != nil {
				httperror.ServerError(c, err)
				return
			}
		}

		if lastUpdate.Add(time.Hour * 6).Before(time.Now()) {
			for i := range course.Sections {
				count, _ := GetSubscriberCount(ctx, course.Sections[i].TopicName)

				view := &model.SubscriptionView{
					TopicName:   course.Sections[i].TopicName,
					Subscribers: int64(count),
				}

				subs = append(subs, view)
			}

			var x []float64
			var weights []float64
			for key := range subs {
				value := subs[key]
				x = append(x, float64(value.Subscribers))
				weights = append(weights, 1)
			}

			mean, std := stat.MeanStdDev(x, weights)

			for key := range subs {
				value := subs[key]
				deviation := (float64(value.Subscribers) - mean) / std
				value.IsHot = deviation > float64(1)

				x = append(x, float64(value.Subscribers))
				weights = append(weights, 1)
			}

			var fireStoreSubs []*FireStoreSubscriptionView

			for key := range subs {
				value := subs[key]

				view := &FireStoreSubscriptionView{
					TopicName:   value.TopicName,
					Subscribers: value.Subscribers,
					IsHot:       value.IsHot,
				}

				fireStoreSubs = append(fireStoreSubs, view)
			}

			m := FireStoreSubscriptionDocument{
				View: fireStoreSubs,
			}

			if _, err := courseRef.Set(ctx, m); err != nil {
				httperror.ServerError(c, err)
				return
			}
		}

		response := model.Response{
			Data: &model.Data{SubscriptionView: subs},
		}

		c.Set(middleware.ResponseKey, response)
	}
}

func SelectCourse(ctx context.Context, courseTopicName string) (course model.Course, b []byte, err error) {
	defer model.TimeTrack(time.Now(), "SelectCourse")
	span := mtrace.NewSpan(ctx, "database.SelectCourse")
	span.SetLabel("topicName", courseTopicName)
	defer span.Finish()

	d := store.Data{}
	m := map[string]interface{}{"topic_name": courseTopicName}
	if err = middleware.Get(ctx, store.SelectCourseQuery, &d, m); err != nil {
		return
	}
	b = d.Data
	err = course.Unmarshal(b)
	return
}
