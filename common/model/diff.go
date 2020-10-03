package model

import (
	log "github.com/Sirupsen/logrus"
)

func DiffAndFilter(oldUni, newUni University) (filteredUniversity University) {
	filteredUniversity = newUni
	oldSubjects := oldUni.Subjects
	newSubjects := newUni.Subjects

	var filteredSubjects []*Subject
	// For each newer subject
	for s := range newSubjects {
		// If current index is out of range of the old subjects[] break and add every subject
		if s >= len(oldSubjects) {
			filteredSubjects = newSubjects
			break
		}

		if !newSubjects[s].Equal(oldSubjects[s]) {
			newSubjects[s].Courses = diffAndFilterCourses(oldSubjects[s].Courses, newSubjects[s].Courses)
			filteredSubjects = append(filteredSubjects, newSubjects[s])
		}
	}

	filteredUniversity.Subjects = filteredSubjects
	return
}

func diffAndFilterCourses(oldCourses, newCourses []*Course) []*Course {
	var filteredCourses []*Course
	for c := range newCourses {
		if c >= len(oldCourses) {
			filteredCourses = newCourses
			break
		}

		if !newCourses[c].Equal(oldCourses[c]) {
			newCourses[c].Sections = diffAndFilterSections(oldCourses[c].Sections, newCourses[c].Sections)
			filteredCourses = append(filteredCourses, newCourses[c])
		}
	}
	return filteredCourses
}

func diffAndFilterSections(oldSections, newSections []*Section) []*Section {
	oldSectionFields := logSection(oldSections, "old")
	newSectionFields := logSection(newSections, "new")
	var filteredSections []*Section
	for e := range newSections {
		if e >= len(oldSections) {
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
				}).Debugln("diff")
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
