package spike

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gonum/stat"
	uctfirestore "github.com/tevjef/uct-backend/common/firestore"
	"github.com/tevjef/uct-backend/common/middleware"
	"github.com/tevjef/uct-backend/common/middleware/cache"
	"github.com/tevjef/uct-backend/common/middleware/httperror"
	"github.com/tevjef/uct-backend/common/model"
)

func hotnessHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		courseTopicName := strings.ToLower(c.Param("topic"))
		firestore := uctfirestore.FromContext(c)

		if course, err := firestore.GetCourse(c, courseTopicName); err != nil {
			httperror.ServerError(c, err)
			return
		} else {
			var subs []*model.SubscriptionView

			//Check if exists
			var shouldUpdate bool
			ch, lastUpdate, err := firestore.GetCourseHotness(c, course.TopicName)
			if err != nil || lastUpdate == nil {
				shouldUpdate = true
			} else {
				shouldUpdate = lastUpdate.Add(time.Hour * 6).Before(time.Now())

				if !shouldUpdate {
					for i := range ch.View {
						sv := ch.View[i]
						subs = append(subs, &model.SubscriptionView{
							IsHot:       sv.IsHot,
							TopicName:   sv.TopicName,
							Subscribers: sv.Subscribers,
						})
					}
				}
			}

			if shouldUpdate {
				for i := range course.Sections {
					count, _ := firestore.GetSubscriptionCount(c, course.Sections[i].TopicName)

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
					value.IsHot = deviation > float64(1) && value.Subscribers > 10

					x = append(x, float64(value.Subscribers))
					weights = append(weights, 1)
				}

				var fireStoreSubs []*uctfirestore.FireStoreSubscriptionView

				for key := range subs {
					value := subs[key]

					view := &uctfirestore.FireStoreSubscriptionView{
						TopicName:   value.TopicName,
						Subscribers: value.Subscribers,
						IsHot:       value.IsHot,
					}

					fireStoreSubs = append(fireStoreSubs, view)
				}

				m := uctfirestore.CourseHotness{
					View: fireStoreSubs,
				}

				if err := firestore.SetCourseHotness(c, courseTopicName, &m); err != nil {
					httperror.ServerError(c, err)
					return
				}
			}

			response := model.Response{
				Data: &model.Data{SubscriptionView: subs},
			}

			c.Set(middleware.ResponseKey, response)
		}

	}, expire)
}

func coursesHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		subjectTopicName := strings.ToLower(c.Param("topic"))
		firestore := uctfirestore.FromContext(c)

		if courses, err := firestore.GetCourses(c, subjectTopicName); err != nil {
			httperror.ServerError(c, err)
			return
		} else {
			response := model.Response{
				Data: &model.Data{Courses: courses},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}, expire)
}
