package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	uct "uct/common"
)

func resolveErr(err error, c *gin.Context) {
	if err == sql.ErrNoRows {
		c.Set(metaKey, errResNotFound(c.Request.RequestURI))
	} else {
		c.Error(&gin.Error{err, gin.ErrorTypePublic, c.Request.URL.String()})
	}
}

func universityHandler(c *gin.Context) {
	topicName := c.ParamValue("topic")
	if u, err := SelectUniversity(topicName); err != nil {
		resolveErr(err, c)
	} else {
		response := uct.Response{
			Data: &uct.Data{University: &u},
		}
		c.Set(responseKey, response)
	}
}

func errResNotFound(str string) uct.Meta {
	code := int32(404)
	message := "Resource: " + str
	errtype := "Not Found"
	return uct.Meta{Code: &code, ErrorMessage: &message, ErrorType: &errtype}
}

func universitiesHandler(c *gin.Context) {
	if universities, err := SelectUniversities(); err != nil {
		resolveErr(err, c)
	} else {
		response := uct.Response{
			Data: &uct.Data{Universities: universities},
		}
		c.Set(responseKey, response)
	}
}

func subjectHandler(c *gin.Context) {
	subjectTopicName := c.ParamValue("topic")

	if sub, _, err := SelectSubject(subjectTopicName); err != nil {
		resolveErr(err, c)
	} else {
		response := uct.Response{
			Data: &uct.Data{Subject: &sub},
		}
		c.Set(responseKey, response)
	}
}

func subjectsHandler(c *gin.Context) {
	season := c.ParamValue("season")
	year := c.ParamValue("year")
	uniTopicName := c.ParamValue("topic")

	if subjects, err := SelectSubjects(uniTopicName, season, year); err != nil {
		resolveErr(err, c)
	} else {
		response := uct.Response{
			Data: &uct.Data{Subjects: subjects},
		}
		c.Set(responseKey, response)
	}
}

func courseHandler(c *gin.Context) {
	courseTopicName := c.ParamValue("topic")
	if course, _, err := SelectCourse(courseTopicName); err != nil {
		resolveErr(err, c)
	} else {
		response := uct.Response{
			Data: &uct.Data{Course: &course},
		}
		c.Set(responseKey, response)
	}

}

func coursesHandler(c *gin.Context) {
	subjectTopicName := c.ParamValue("topic")
	if courses, err := SelectCourses(subjectTopicName); err != nil {
		resolveErr(err, c)
	} else {
		response := uct.Response{
			Data: &uct.Data{Courses: courses},
		}
		c.Set(responseKey, response)
	}
}

func sectionHandler(c *gin.Context) {
	sectionTopicName := c.ParamValue("topic")
	if s, err := SelectSection(sectionTopicName); err != nil {
		resolveErr(err, c)
	} else {
		response := uct.Response{
			Data: &uct.Data{Section: &s},
		}
		c.Set(responseKey, response)
	}
}
