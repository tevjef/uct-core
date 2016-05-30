package main

import (
	"github.com/gin-gonic/gin"
	uct "uct/common"
)

func universityHandler(c *gin.Context) {
	topicName := c.Param("topic")
	if u, err := SelectUniversity(topicName); err != nil {

		c.Error(gin.Error{err, gin.ErrorTypePublic, c.Request.URL.String()})
	} else {
		if b, err := u.Marshal(); err != nil {

			c.Error(gin.Error{err, gin.ErrorTypePublic, c.Request.URL.String()})
		} else {
			c.Set("protobuf", b)
			c.Set("object", u)
		}
	}
}

func universitiesHandler(c *gin.Context) {
	if universityList, err := SelectUniversities(); err != nil {

		c.Error(gin.Error{err, gin.ErrorTypePublic, c.Request.URL.String()})
	} else {
		universities := uct.Universities{Universities: universityList}
		if b, err := universities.Marshal(); err != nil {

			c.Error(gin.Error{err, gin.ErrorTypePublic, c.Request.URL.String()})
		} else {
			c.Set("protobuf", b)
			c.Set("object", universities)
		}
	}
}

func subjectHandler(c *gin.Context) {
	subjectTopicName := c.Param("topic")

	if sub, b, err := SelectSubject(subjectTopicName); err != nil {

		c.Error(gin.Error{err, gin.ErrorTypePublic, c.Request.URL.String()})
	} else {
		c.Set("protobuf", b)
		c.Set("object", sub)
	}

}

func subjectsHandler(c *gin.Context) {
	season := c.Param("season")
	year := c.Param("year")
	uniTopicName := c.Param("topic")

	if subjects, err := SelectSubjects(uniTopicName, season, year); err != nil {
		c.Error(gin.Error{err, gin.ErrorTypePublic, c.Request.URL.String()})
	} else {
		s := uct.Subjects{Subjects: subjects}
		if b, err := s.Marshal(); err != nil {
			c.Error(gin.Error{err, gin.ErrorTypePublic, c.Request.URL.String()})
		} else {
			c.Set("protobuf", b)
			c.Set("object", s)
		}
	}

}

func courseHandler(c *gin.Context) {
	courseTopicName := c.Param("topic")

	if course, b, err := SelectCourse(courseTopicName); err != nil {
		c.Error(gin.Error{err, gin.ErrorTypePublic, c.Request.URL.String()})
	} else {
		c.Set("protobuf", b)
		c.Set("object", course)
	}

}

func coursesHandler(c *gin.Context) {
	subjectTopicName := c.Param("topic")
	if courseList, err := SelectCourses(subjectTopicName); err != nil {
		c.Error(gin.Error{err, gin.ErrorTypePublic, c.Request.URL.String()})
	} else {
		courses := uct.Courses{Courses: courseList}
		if b, err := courses.Marshal(); err != nil {
			c.Error(gin.Error{err, gin.ErrorTypePublic, c.Request.URL.String()})
		} else {
			c.Set("protobuf", b)
			c.Set("object", courses)
		}
	}

}

func sectionHandler(c *gin.Context) {
	sectionTopicName := c.Param("topic")
	if s, err := SelectSection(sectionTopicName); err != nil {
		c.Error(gin.Error{err, gin.ErrorTypePublic, c.Request.URL.String()})
	} else {
		if b, err := s.Marshal(); err != nil {
			c.Error(gin.Error{err, gin.ErrorTypePublic, c.Request.URL.String()})
		} else {
			c.Set("protobuf", b)
			c.Set("object", s)
		}
	}
}
