package main

import (
	"testing"
	"time"

	"github.com/tevjef/uct-core/common/model"

	"github.com/stretchr/testify/assert"
)

func TestRequestSubjects(t *testing.T) {
	rr := rutgersRequest{
		host:     "http://sis.rutgers.edu/soc",
		semester: *model.ResolveSemesters(time.Now(), getRutgers("NK").Registrations).Current,
		campus:   "NK",
	}

	s := subjectRequest{rutgersRequest: rr}.requestSubjects()

	assert.True(t, len(s) > 10)
	for _, val := range s {
		assert.True(t, val.Name != "")
	}

	rr.campus = "NB"

	s = subjectRequest{rutgersRequest: rr}.requestSubjects()

	assert.True(t, len(s) > 10)
	for _, val := range s {
		assert.True(t, val.Name != "")
	}

	rr.campus = "CM"

	s = subjectRequest{rutgersRequest: rr}.requestSubjects()

	assert.True(t, len(s) > 10)
	for _, val := range s {
		assert.True(t, val.Name != "")
	}
}

func TestRequestCourses(t *testing.T) {
	rr := rutgersRequest{
		host:     "http://sis.rutgers.edu/soc",
		semester: *model.ResolveSemesters(time.Now(), getRutgers("NK").Registrations).Current,
		campus:   "NK",
	}

	s := subjectRequest{rutgersRequest: rr}.requestSubjects()

	c := courseRequest{rutgersRequest: rr, subject: s[10].Number}.requestCourses()

	assert.True(t, len(c) > 1)
	for _, val := range c {
		assert.True(t, val.CourseNumber != "")
	}
}

func TestGetCampus(t *testing.T) {
	u := getCampus("CM")
	assert.True(t, len(u.Subjects) > 1)
}
