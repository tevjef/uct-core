package model

import "strings"

type SemesterSorter []*Semester

func (a SemesterSorter) Len() int {
	return len(a)
}

func (a SemesterSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a SemesterSorter) Less(i, j int) bool {
	if a[j].Year < a[i].Year {
		return true
	} else if a[i].Year == a[j].Year {
		return rankSeason(a[j].Season) < rankSeason(a[i].Season)
	}
	return false
}

func rankSeason(seasonStr string) int {
	seasonStr = strings.ToLower(seasonStr)
	switch seasonStr {
	case "fall":
		return 4
	case "summer":
		return 3
	case "spring":
		return 2
	case "winter":
		return 1
	}
	return 0
}

type courseSorter struct {
	courses []*Course
}

func (a courseSorter) Len() int {
	return len(a.courses)
}

func (a courseSorter) Swap(i, j int) {
	a.courses[i], a.courses[j] = a.courses[j], a.courses[i]
}

func (a courseSorter) Less(i, j int) bool {
	c1 := a.courses[i]
	c2 := a.courses[j]

	left := c1.Name + c1.Number
	right := c2.Name + c2.Number

	if left == right {
		return len(c1.Sections) < len(c2.Sections)
	}

	return left < right
}

type sectionSorter struct {
	sections []*Section
}

func (a sectionSorter) Len() int {
	return len(a.sections)
}

func (a sectionSorter) Swap(i, j int) {
	a.sections[i], a.sections[j] = a.sections[j], a.sections[i]
}

func (a sectionSorter) Less(i, j int) bool {
	return a.sections[i].Number < a.sections[j].Number
}

type instructorSorter struct {
	instructors []*Instructor
}

func (a instructorSorter) Len() int {
	return len(a.instructors)
}

func (a instructorSorter) Swap(i, j int) {
	a.instructors[i], a.instructors[j] = a.instructors[j], a.instructors[i]
}

func (a instructorSorter) Less(i, j int) bool {
	return a.instructors[i].Name < a.instructors[j].Name
}

type metadataSorter struct {
	metadatas []Metadata
}

func (a metadataSorter) Len() int {
	return len(a.metadatas)
}

func (a metadataSorter) Swap(i, j int) {
	a.metadatas[i], a.metadatas[j] = a.metadatas[j], a.metadatas[i]
}

func (a metadataSorter) Less(i, j int) bool {
	m1 := a.metadatas[i]
	m2 := a.metadatas[j]
	return (m1.Title + m1.Content) < (m2.Title + m2.Content)
}

type meetingSorter struct {
	meetings []Meeting
}

func (a meetingSorter) Len() int {
	return len(a.meetings)
}

func (a meetingSorter) Swap(i, j int) {
	a.meetings[i], a.meetings[j] = a.meetings[j], a.meetings[i]
}

func (a meetingSorter) Less(i, j int) bool {
	/*	if a.meetings[i].dayRank() < a.meetings[j].dayRank() {
		return a.meetings[i].hash() < a.meetings[j].hash()
	}*/
	return a.meetings[i].dayRank() < a.meetings[j].dayRank()
}

type SectionByNumber []Section

func (a SectionByNumber) Len() int {
	return len(a)
}

func (a SectionByNumber) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a SectionByNumber) Less(i, j int) bool {
	return strings.Compare(a[i].Number, a[j].Number) < 0
}

type UniversityByName []*University

func (a UniversityByName) Len() int {
	return len(a)
}

func (a UniversityByName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a UniversityByName) Less(i, j int) bool {
	return strings.Compare(a[i].Name, a[j].Name) < 0
}

type CourseByName []Course

func (a CourseByName) Len() int {
	return len(a)
}

func (a CourseByName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a CourseByName) Less(i, j int) bool {
	return strings.Compare(a[i].Name, a[j].Name) < 0
}

type SubjectByName []Subject

func (a SubjectByName) Len() int {
	return len(a)
}

func (a SubjectByName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a SubjectByName) Less(i, j int) bool {
	return strings.Compare(a[i].Name, a[j].Name) < 0
}

func (meeting Meeting) dayRank() int {
	if meeting.Day == nil {
		return 8
	}
	switch *meeting.Day {
	case "Monday":
		return 1
	case "Tuesday":
		return 2
	case "Wednesday":
		return 3
	case "Thurdsday":
		return 4
	case "Friday":
		return 5
	case "Saturday":
		return 6
	case "Sunday":
		return 7
	}
	return 8
}
