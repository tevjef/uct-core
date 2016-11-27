package main

import (
	"github.com/gin-gonic/gin"
	"strings"
	"uct/common/model"
	"time"
	"github.com/tevjef/contrib/cache"
	"uct/servers/spike/middleware"
)

func universityHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(expire, func(c *gin.Context) {
		topicName := strings.ToLower(c.ParamValue("topic"))

		if u, err := SelectUniversity(topicName); err != nil {
			middleware.ResolveErr(err, c)
		} else {
			response := model.Response{
				Data: &model.Data{University: &u},
			}
			c.Set(middleware.ResponseKey, response)
		}
	})
}

func universitiesHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(expire, func(c *gin.Context) {
		if universities, err := SelectUniversities(); err != nil {
			middleware.ResolveErr(err, c)
		} else {
			response := model.Response{
				Data: &model.Data{Universities: universities},
			}
			c.Set(middleware.ResponseKey, response)
		}
	})
}

func subjectHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(expire, func(c *gin.Context) {
		subjectTopicName := strings.ToLower(c.ParamValue("topic"))

		if sub, _, err := SelectSubject(subjectTopicName); err != nil {
			middleware.ResolveErr(err, c)
		} else {
			response := model.Response{
				Data: &model.Data{Subject: &sub},
			}
			c.Set(middleware.ResponseKey, response)
		}
	})
}

func subjectsHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(expire, func(c *gin.Context) {
		season := strings.ToLower(c.ParamValue("season"))
		year := c.ParamValue("year")
		uniTopicName := strings.ToLower(c.ParamValue("topic"))

		if subjects, err := SelectSubjects(uniTopicName, season, year); err != nil {
			middleware.ResolveErr(err, c)
		} else {
			response := model.Response{
				Data: &model.Data{Subjects: subjects},
			}
			c.Set(middleware.ResponseKey, response)
		}
	})
}

func courseHandler(expire time.Duration)  gin.HandlerFunc {
	return cache.CachePage(expire, func(c *gin.Context) {
		courseTopicName := strings.ToLower(c.ParamValue("topic"))

		if course, _, err := SelectCourse(courseTopicName); err != nil {
			middleware.ResolveErr(err, c)
		} else {
			response := model.Response{
				Data: &model.Data{Course: &course},
			}
			c.Set(middleware.ResponseKey, response)
		}
	})
}

func coursesHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(expire, func(c *gin.Context) {
		subjectTopicName := strings.ToLower(c.ParamValue("topic"))

		if courses, err := SelectCourses(subjectTopicName); err != nil {
			middleware.ResolveErr(err, c)
		} else {
			response := model.Response{
				Data: &model.Data{Courses: courses},
			}
			c.Set(middleware.ResponseKey, response)
		}
	})
}

func sectionHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(expire, func(c *gin.Context) {
		sectionTopicName := strings.ToLower(c.ParamValue("topic"))

		if s, _, err := SelectSection(sectionTopicName); err != nil {
			middleware.ResolveErr(err, c)
		} else {
			response := model.Response{
				Data: &model.Data{Section: &s},
			}
			c.Set(middleware.ResponseKey, response)
		}
	})
}
