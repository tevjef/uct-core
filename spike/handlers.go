package main

import (
	"strings"
	"time"
	"uct/common/model"
	"uct/spike/cache"
	"uct/spike/middleware"

	"github.com/gin-gonic/gin"
)

func universityHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		topicName := strings.ToLower(c.ParamValue("topic"))

		if u, err := SelectUniversity(topicName); err != nil {
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
		if universities, err := SelectUniversities(); err != nil {
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

		if sub, _, err := SelectSubject(subjectTopicName); err != nil {
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

		if subjects, err := SelectSubjects(uniTopicName, season, year); err != nil {
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

		if course, _, err := SelectCourse(courseTopicName); err != nil {
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

		if courses, err := SelectCourses(subjectTopicName); err != nil {
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

		if s, _, err := SelectSection(sectionTopicName); err != nil {
			middleware.ResolveErr(err, c)
		} else {
			response := model.Response{
				Data: &model.Data{Section: &s},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}, expire)
}
