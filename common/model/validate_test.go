package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckUniqueSubject(t *testing.T) {
	subjects := []*Subject{
		{Season: Fall, Year: "2012", Name: "a", Number: "1"},
		{Season: Fall, Year: "2012", Name: "a", Number: "1"},
		{Season: Fall, Year: "2012", Name: "a", Number: "1"},
	}

	makeUniqueSubjects(subjects)

	assert.True(t, subjects[0].Name == "a")
	assert.True(t, subjects[1].Name == "a_2")
	assert.True(t, subjects[2].Name == "a_3")
}

func TestCheckUniqueCourse(t *testing.T) {
	courses := []*Course{
		{Name: "a", Number: "1"},
		{Name: "a", Number: "1"},
		{Name: "a", Number: "1"},
	}

	makeUniqueCourses(&Subject{}, courses)

	assert.True(t, courses[0].Name == "a")
	assert.True(t, courses[1].Name == "a_2")
	assert.True(t, courses[2].Name == "a_3")
}
