package main

import (
	"github.com/gin-gonic/gin"
	uct "uct/common"
	"uct/servers"
	"strings"
)

func universityHandler(c *gin.Context) {
	topicName := strings.ToLower(c.ParamValue("topic"))

	if u, err := SelectUniversity(topicName); err != nil {
		servers.ResolveErr(err, c)
	} else {
		response := uct.Response{
			Data: &uct.Data{University: &u},
		}
		c.Set(servers.ResponseKey, response)
	}
}

func universitiesHandler(c *gin.Context) {
	if universities, err := SelectUniversities(); err != nil {
		servers.ResolveErr(err, c)
	} else {
		response := uct.Response{
			Data: &uct.Data{Universities: universities},
		}
		c.Set(servers.ResponseKey, response)
	}
}

func subjectHandler(c *gin.Context) {
	subjectTopicName := strings.ToLower(c.ParamValue("topic"))

	if sub, _, err := SelectSubject(subjectTopicName); err != nil {
		servers.ResolveErr(err, c)
	} else {
		response := uct.Response{
			Data: &uct.Data{Subject: &sub},
		}
		c.Set(servers.ResponseKey, response)
	}
}

func subjectsHandler(c *gin.Context) {
	season := strings.ToLower(c.ParamValue("season"))
	year := c.ParamValue("year")
	uniTopicName := strings.ToLower(c.ParamValue("topic"))

	if subjects, err := SelectSubjects(uniTopicName, season, year); err != nil {
		servers.ResolveErr(err, c)
	} else {
		response := uct.Response{
			Data: &uct.Data{Subjects: subjects},
		}
		c.Set(servers.ResponseKey, response)
	}
}

func courseHandler(c *gin.Context) {
	courseTopicName := strings.ToLower(c.ParamValue("topic"))

	if course, _, err := SelectCourse(courseTopicName); err != nil {
		servers.ResolveErr(err, c)
	} else {
		response := uct.Response{
			Data: &uct.Data{Course: &course},
		}
		c.Set(servers.ResponseKey, response)
	}

}

func coursesHandler(c *gin.Context) {
	subjectTopicName := strings.ToLower(c.ParamValue("topic"))

	if courses, err := SelectCourses(subjectTopicName); err != nil {
		servers.ResolveErr(err, c)
	} else {
		response := uct.Response{
			Data: &uct.Data{Courses: courses},
		}
		c.Set(servers.ResponseKey, response)
	}
}

func sectionHandler(c *gin.Context) {
	sectionTopicName := strings.ToLower(c.ParamValue("topic"))

	if s, _, err := SelectSection(sectionTopicName); err != nil {
		servers.ResolveErr(err, c)
	} else {
		response := uct.Response{
			Data: &uct.Data{Section: &s},
		}
		c.Set(servers.ResponseKey, response)
	}
}
