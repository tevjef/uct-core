package main

import (
	"strings"
	"time"
	"uct/common/model"
	"uct/spike/middleware"

	"uct/spike/middleware/cache"

	"github.com/gin-gonic/gin"
)

func universityHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		topicName := strings.ToLower(c.ParamValue("topic"))

		if u, err := SelectUniversity(c, topicName); err != nil {
			middleware.ResolveErr(err, c)
		} else {
			response := model.Response{
				Data: &model.Data{University: &u},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}, expire)
}

func universitiesHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		if universities, err := SelectUniversities(c); err != nil {
			middleware.ResolveErr(err, c)
		} else {
			response := model.Response{
				Data: &model.Data{Universities: universities},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}, expire)
}

func subjectHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		subjectTopicName := strings.ToLower(c.ParamValue("topic"))

		if sub, _, err := SelectSubject(c, subjectTopicName); err != nil {
			middleware.ResolveErr(err, c)
		} else {
			response := model.Response{
				Data: &model.Data{Subject: &sub},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}, expire)
}

func subjectsHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		season := strings.ToLower(c.ParamValue("season"))
		year := c.ParamValue("year")
		uniTopicName := strings.ToLower(c.ParamValue("topic"))

		if subjects, err := SelectSubjects(c, uniTopicName, season, year); err != nil {
			middleware.ResolveErr(err, c)
		} else {
			response := model.Response{
				Data: &model.Data{Subjects: subjects},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}, expire)
}

func courseHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		courseTopicName := strings.ToLower(c.ParamValue("topic"))

		if course, _, err := SelectCourse(c, courseTopicName); err != nil {
			middleware.ResolveErr(err, c)
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
			middleware.ResolveErr(err, c)
		} else {
			response := model.Response{
				Data: &model.Data{Courses: courses},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}, expire)
}

func sectionHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		sectionTopicName := strings.ToLower(c.ParamValue("topic"))

		if s, _, err := SelectSection(c, sectionTopicName); err != nil {
			middleware.ResolveErr(err, c)
		} else {
			response := model.Response{
				Data: &model.Data{Section: &s},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}, expire)
}
