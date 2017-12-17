package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tevjef/uct-core/common/model"
	"github.com/tevjef/uct-core/spike/middleware"
	"github.com/tevjef/uct-core/spike/middleware/cache"
	"github.com/tevjef/uct-core/spike/middleware/httperror"
	mtrace "github.com/tevjef/uct-core/spike/middleware/trace"
	"github.com/tevjef/uct-core/spike/store"
	"golang.org/x/net/context"
)

func courseHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		courseTopicName := strings.ToLower(c.ParamValue("topic"))

		if course, _, err := SelectCourse(c, courseTopicName); err != nil {
			if err == sql.ErrNoRows {
				httperror.NotFound(c, err)
				return
			}
			httperror.ServerError(c, err)
			return
		} else {
			response := model.Response{
				Data: &model.Data{Course: &course},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}, expire)
}

func coursesHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		subjectTopicName := strings.ToLower(c.ParamValue("topic"))

		if courses, err := SelectCourses(c, subjectTopicName); err != nil {
			if err == sql.ErrNoRows {
				httperror.NotFound(c, err)
				return
			}
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

func SelectCourse(ctx context.Context, courseTopicName string) (course model.Course, b []byte, err error) {
	defer model.TimeTrack(time.Now(), "SelectCourse")
	span := mtrace.NewSpan(ctx, "database.SelectCourse")
	span.SetLabel("topicName", courseTopicName)
	defer span.Finish()

	d := store.Data{}
	m := map[string]interface{}{"topic_name": courseTopicName}
	if err = store.Get(ctx, store.SelectCourseQuery, &d, m); err != nil {
		return
	}
	b = d.Data
	err = course.Unmarshal(b)
	return
}

func SelectCourses(ctx context.Context, subjectTopicName string) (courses []*model.Course, err error) {
	defer model.TimeTrack(time.Now(), "SelectCourses")
	span := mtrace.NewSpan(ctx, "database.SelectCourses")
	span.SetLabel("topicName", subjectTopicName)
	defer span.Finish()

	var d []store.Data
	m := map[string]interface{}{"topic_name": subjectTopicName}
	if err = store.Select(ctx, store.ListCoursesQuery, &d, m); err != nil {
		return
	}
	if err == nil && len(courses) == 0 {
		err = httperror.NoDataFound(fmt.Sprintf("No courses found for %s", subjectTopicName))
	}
	for i := range d {
		c := model.Course{}
		if err = c.Unmarshal(d[i].Data); err != nil {
			return
		}
		courses = append(courses, &c)
	}

	return
}
