package model

import (
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func MakeSemesterKey(semester *Semester) string {
	return semester.Season + fmt.Sprint(semester.Year)
}

func MakeSemester(season, year string) *Semester {
	yearInt, _ := strconv.Atoi(year)

	semester := &Semester{
		Year:   int32(yearInt),
		Season: season,
	}
	return semester
}

func GroupSubjectsForSemester(subjects []*Subject) map[string][]*Subject {
	m := map[string][]*Subject{}
	for subjectIndex := range subjects {
		subject := subjects[subjectIndex]
		key := MakeSemesterKey(MakeSemester(subject.Season, subject.Year))
		m[key] = append(m[key], subject)
	}

	return m
}

func DiffAndFilter(oldUni, newUni University) (filteredUniversity University) {
	oldSeasonedSubjects := GroupSubjectsForSemester(oldUni.Subjects)
	newSeasonedSubjects := GroupSubjectsForSemester(newUni.Subjects)

	var filteredSubjects []*Subject

	for seasonKey := range newSeasonedSubjects {
		newSubjects := newSeasonedSubjects[seasonKey]
		oldSubjects, ok := oldSeasonedSubjects[seasonKey]
		// A new semester is found

		logKey := oldUni.TopicName + "." + seasonKey

		if !ok {
			log.Debugf("diff: %v: new semester found %v", oldUni.TopicName, seasonKey)
			filteredSubjects = append(filteredSubjects, newSubjects...)
			continue
		}

		if len(newSubjects) != len(oldSubjects) {
			log.Debugf("diff: %v: newSubjects: %v != oldSubjects: %v", logKey, len(newSubjects), len(oldSubjects))
			filteredSubjects = append(filteredSubjects, newSubjects...)
			continue
		}

		for s := range newSubjects {
			if !newSubjects[s].Equal(oldSubjects[s]) {
				log.Debugf("diff: %v: change detected in subject: %v", logKey, newSubjects[s].TopicName)
				filteredSubject := *newSubjects[s]
				filteredSubject.Courses = diffAndFilterCourses(logKey, oldSubjects[s].Courses, newSubjects[s].Courses)
				filteredSubjects = append(filteredSubjects, &filteredSubject)
			}
		}
	}

	filteredUniversity = newUni
	filteredUniversity.Subjects = filteredSubjects
	return
}

func diffAndFilterCourses(logKey string, oldCourses, newCourses []*Course) []*Course {
	var filteredCourses []*Course
	if len(newCourses) != len(oldCourses) {
		log.Debugf("diff: %v: newCourses: %v != oldCourses: %v", logKey, len(newCourses), len(oldCourses))
		return newCourses
	}

	for c := range newCourses {
		if !newCourses[c].Equal(oldCourses[c]) {
			log.Debugf("diff: %v: change detected in course: %v", logKey, newCourses[c].TopicName)
			filteredCourse := *newCourses[c]
			filteredCourse.Sections = diffAndFilterSections(logKey, oldCourses[c].Sections, newCourses[c].Sections)
			filteredCourses = append(filteredCourses, &filteredCourse)
		}
	}
	return filteredCourses
}

func diffAndFilterSections(logKey string, oldSections, newSections []*Section) []*Section {
	oldSectionFields := logSection(oldSections, "old")
	newSectionFields := logSection(newSections, "new")
	var filteredSections []*Section
	for e := range newSections {
		if len(newSections) != len(oldSections) {
			filteredSections = newSections
			break
		}
		if !newSections[e].Equal(oldSections[e]) {
			log.WithFields(log.Fields{
				"old_call_number": oldSections[e].CallNumber,
				"old_status":      oldSections[e].Status,
				"new_call_number": newSections[e].CallNumber,
				"new_status":      newSections[e].Status,
			}).WithFields(oldSectionFields).WithFields(newSectionFields).WithFields(
				log.Fields{
					"old_section": oldSections[e].TopicName,
					"new_section": newSections[e].TopicName,
				}).Debugf("diff: %v: change detected in section: %v", logKey, newSections[e].TopicName)
			filteredSections = append(filteredSections, newSections[e])
		}
	}
	return filteredSections
}

func logSection(section []*Section, prepend string) log.Fields {
	openCount := 0
	closedCount := 0
	for i := range section {
		if section[i].Status == "Open" {
			openCount++
		} else if section[i].Status == "Closed" {
			closedCount++
		}
	}
	return log.Fields{
		prepend + "_open_count":   openCount,
		prepend + "_closed_count": closedCount,
	}
}
