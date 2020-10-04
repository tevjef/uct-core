package uctfirestore

import (
	"fmt"

	"github.com/tevjef/uct-backend/common/model"
)

func MakeSemesterKey(semester *model.Semester) string {
	return semester.Season + fmt.Sprint(semester.Year)
}

func getSubjectsForSemester(subjects []*model.Subject, semester *model.Semester) []*model.Subject {
	var current []*model.Subject

	for subjectIndex := range subjects {
		subject := subjects[subjectIndex]

		if subject.Year == fmt.Sprint(semester.Year) && subject.Season == semester.Season {
			current = append(current, subject)
		}
	}

	return current
}
