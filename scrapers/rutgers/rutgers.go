package main

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/tevjef/uct-backend/common/model"
)

var toTitle = model.ToTitle
var trim = strings.TrimSpace

type (
	RSubject struct {
		Name    string     `json:"description,omitempty"`
		Number  string     `json:"code,omitempty"`
		Courses []*RCourse `json:"courses,omitempty"`
		Season  string
		Year    int
	}

	RCourse struct {
		SubjectNotes      string        `json:"subjectNotes"`
		CourseNumber      string        `json:"courseNumber"`
		Subject           string        `json:"subject"`
		CampusCode        string        `json:"campusCode"`
		OpenSections      int           `json:"openSections"`
		SynopsisURL       string        `json:"synopsisUrl"`
		SubjectGroupNotes string        `json:"subjectGroupNotes"`
		OfferingUnitCode  string        `json:"offeringUnitCode"`
		OfferingUnitTitle string        `json:"offeringUnitTitle"`
		Title             string        `json:"title"`
		CourseDescription string        `json:"courseDescription"`
		PreReqNotes       string        `json:"preReqNotes"`
		Sections          []*RSection   `json:"sections"`
		SupplementCode    string        `json:"supplementCode"`
		Credits           float64       `json:"credits"`
		UnitNotes         string        `json:"unitNotes"`
		CoreCodes         []interface{} `json:"coreCodes"`
		CourseNotes       string        `json:"courseNotes"`
		ExpandedTitle     string        `json:"expandedTitle"`
	}

	RSection struct {
		SectionEligibility                   string                 `json:"sectionEligibility"`
		SessionDatePrintIndicator            string                 `json:"sessionDatePrintIndicator"`
		ExamCode                             string                 `json:"examCode"`
		SpecialPermissionAddCode             string                 `json:"specialPermissionAddCode"`
		CrossListedSections                  []RCrossListedSections `json:"crossListedSections"`
		SectionNotes                         string                 `json:"sectionNotes"`
		SpecialPermissionDropCode            string                 `json:"specialPermissionDropCode"`
		Instructor                           []RInstructor          `json:"instructors"`
		Number                               string                 `json:"number"`
		Majors                               []RMajor               `json:"majors"`
		SessionDates                         string                 `json:"sessionDates"`
		SpecialPermissionDropCodeDescription string                 `json:"specialPermissionDropCodeDescription"`
		Subtopic                             string                 `json:"subtopic"`
		SynopsisUrl                          string                 `json:"synopsisUrl"`
		OpenStatus                           bool                   `json:"openStatus"`
		Comments                             []RComment             `json:"comments"`
		Minors                               []interface{}          `json:"minors"`
		CampusCode                           string                 `json:"campusCode"`
		Index                                string                 `json:"index"`
		UnitMajors                           []interface{}          `json:"unitMajors"`
		Printed                              string                 `json:"printed"`
		SpecialPermissionAddCodeDescription  string                 `json:"specialPermissionAddCodeDescription"`
		Subtitle                             string                 `json:"subtitle"`
		MeetingTimes                         []RMeetingTime         `json:"meetingTimes"`
		LegendKey                            string                 `json:"legendKey"`
		HonorPrograms                        []interface{}          `json:"honorPrograms"`
		status                               string
		creditsFloat                         string
		course                               RCourse
	}

	RInstructor struct {
		Name string `json:"name"`
	}

	RMajor struct {
		IsMajorCode bool   `json:"isMajorCode"`
		IsUnitCode  bool   `json:"isUnitCode"`
		Code        string `json:"code"`
	}

	RComment struct {
		Code        string `json:"code"`
		Description string `json:"description"`
	}

	RCrossListedSections struct {
		CourseNumber             string `json:"courseNumber"`
		SupplementCode           string `json:"supplementCode"`
		SectionNumber            string `json:"sectionNumber"`
		OfferingUnitCampus       string `json:"offeringUnitCampus"`
		PrimaryRegistrationIndex string `json:"primaryRegistrationIndex"`
		OfferingUnitCode         string `json:"offeringUnitCode"`
		RegistrationIndex        string `json:"registrationIndex"`
		SubjectCode              string `json:"subjectCode"`
	}

	RMeetingTime struct {
		CampusLocation  string `json:"campusLocation"`
		BaClassHours    string `json:"baClassHours"`
		RoomNumber      string `json:"roomNumber"`
		PmCode          string `json:"pmCode"`
		CampusAbbrev    string `json:"campusAbbrev"`
		CampusName      string `json:"campusName"`
		MeetingDay      string `json:"meetingDay"`
		BuildingCode    string `json:"buildingCode"`
		StartTime       string `json:"startTime"`
		EndTime         string `json:"endTime"`
		MeetingModeDesc string `json:"meetingModeDesc"`
		MeetingModeCode string `json:"meetingModeCode"`
		ClassType       string
	}
)

func (subject *RSubject) clean() {
	subject.Name = toTitle(subject.Name)
}

func (course *RCourse) clean() {
	m := map[string]int{}

	// Filter duplicate sections
	course.Sections = filterSections(course.Sections, func(section *RSection) bool {
		key := section.Index + section.Number
		m[key]++
		return m[key] <= 1 && section.Printed == "Y"
	})

	for j := range course.Sections {
		course.Sections[j].course = *course
		course.Sections[j].clean()
	}

	sort.Sort(SectionSorter{course.Sections})

	course.Title = toTitle(trimAll(course.Title))
	course.CourseNumber = trimAll(course.CourseNumber)
	course.CourseDescription = trimAll(course.CourseDescription)
	course.CourseNotes = trimAll(course.CourseNotes)
	course.SubjectNotes = trimAll(course.SubjectNotes)
	course.SynopsisURL = trimAll(course.SynopsisURL)
	course.PreReqNotes = trimAll(course.PreReqNotes)
	course.ExpandedTitle = trimAll(course.ExpandedTitle)

	if len(course.ExpandedTitle) == 0 {
		course.ExpandedTitle = course.Title
	}
}

func (section *RSection) clean() {
	section.Subtitle = trimAll(section.Subtitle)
	section.SectionNotes = trimAll(section.SectionNotes)
	section.CampusCode = trimAll(section.CampusCode)
	section.SpecialPermissionAddCodeDescription = trimAll(section.SpecialPermissionAddCodeDescription)

	for i := range section.MeetingTimes {
		section.MeetingTimes[i].clean()
	}

	if section.OpenStatus {
		section.status = model.Open.String()
	} else {
		section.status = model.Closed.String()
	}

	section.creditsFloat = fmt.Sprintf("%.1f", section.course.Credits)

	sort.Sort(MeetingByClass(section.MeetingTimes))
	sort.Sort(instructorSorter{section.Instructor})
}

func (meeting *RMeetingTime) clean() {
	meeting.StartTime = trimAll(meeting.StartTime)
	meeting.EndTime = trimAll(meeting.EndTime)

	meeting.MeetingDay = meeting.day()
	meeting.StartTime = meeting.getMeetingHourStart()
	meeting.EndTime = meeting.getMeetingHourEnd()
	meeting.ClassType = meeting.classType()

	// Some meetings may not have a type, default to lecture
	if meeting.MeetingModeCode == "" {
		meeting.MeetingModeCode = "02"
		meeting.MeetingModeDesc = "LEC"
	}

	if meeting.BuildingCode != "" {
		meeting.RoomNumber = meeting.BuildingCode + "-" + meeting.RoomNumber
	} else {
		meeting.RoomNumber = ""
	}
}

func (section RSection) instructor() (instructors []*model.Instructor) {
	for _, instructor := range section.Instructor {
		instructors = append(instructors, &model.Instructor{Name: instructor.Name})
	}
	return
}

func (section RSection) metadata() (metadata []*model.Metadata) {

	if len(section.CrossListedSections) > 0 {
		crossListedSections := []string{}
		for _, cls := range section.CrossListedSections {
			crossListedSections = append(crossListedSections, cls.OfferingUnitCode+":"+cls.SubjectCode+":"+cls.CourseNumber+":"+cls.SectionNumber)
		}
		metadata = append(metadata, &model.Metadata{
			Title:   "Cross-listed Sections",
			Content: strings.Join(crossListedSections, ", "),
		})

	}

	if len(section.Comments) > 0 {
		sort.Sort(commentSorter{section.Comments})
		comments := []string{}
		for _, comment := range section.Comments {
			comments = append(comments, comment.Description)
		}

		metadata = append(metadata, &model.Metadata{
			Title:   "Comments",
			Content: strings.Join(comments, ", "),
		})
	}

	if len(section.Majors) > 0 {
		var openTo []string
		var majors []string
		var schools []string

		for _, unit := range section.Majors {
			if unit.IsMajorCode {
				majors = append(majors, unit.Code)
			} else if unit.IsUnitCode {
				schools = append(schools, unit.Code)
			}
		}

		sort.Strings(majors)
		sort.Strings(schools)

		if len(majors) > 0 {
			openTo = append(openTo, "Majors: "+strings.Join(majors, ", "))
		}

		if len(schools) > 0 {
			openTo = append(openTo, "Schools: "+strings.Join(schools, ", "))
		}

		if len(openTo) > 0 {
			metadata = append(metadata, &model.Metadata{
				Title:   "Open To",
				Content: strings.Join(openTo, ", "),
			})
		}
	}

	if len(section.SectionNotes) > 0 {
		metadata = append(metadata, &model.Metadata{
			Title:   "Section Notes",
			Content: section.SectionNotes,
		})
	}

	if len(section.SynopsisUrl) > 0 {
		metadata = append(metadata, &model.Metadata{
			Title:   "Synopsis Url",
			Content: section.SynopsisUrl,
		})
	}

	if len(section.ExamCode) > 0 {
		metadata = append(metadata, &model.Metadata{
			Title:   "Exam Code",
			Content: getExamCode(section.ExamCode),
		})
	}

	if len(section.SpecialPermissionAddCodeDescription) > 0 {
		metadata = append(metadata, &model.Metadata{
			Title:   "Special Permission",
			Content: "Code: " + section.SpecialPermissionAddCode + "\n" + section.SpecialPermissionAddCodeDescription,
		})
	}

	if len(section.Subtitle) > 0 {
		metadata = append(metadata, &model.Metadata{
			Title:   "Subtitle",
			Content: section.Subtitle,
		})
	}

	return
}

type SectionSorter struct {
	Sections []*RSection
}

func (a SectionSorter) Len() int {
	return len(a.Sections)
}

func (a SectionSorter) Swap(i, j int) {
	a.Sections[i], a.Sections[j] = a.Sections[j], a.Sections[i]
}

func (a SectionSorter) Less(i, j int) bool {
	return a.Sections[i].Index < a.Sections[j].Index
}

type CourseSorter struct {
	Courses []*RCourse
}

func (a CourseSorter) Len() int {
	return len(a.Courses)
}

func (a CourseSorter) Swap(i, j int) {
	a.Courses[i], a.Courses[j] = a.Courses[j], a.Courses[i]
}

func (a CourseSorter) Less(i, j int) bool {
	c1 := a.Courses[i]
	c2 := a.Courses[j]
	var hash = func(s []*RSection) string {
		var buffer bytes.Buffer
		for i := range s {
			buffer.WriteString(s[i].Index)
			buffer.WriteString(s[i].SectionNotes)
			buffer.WriteString(s[i].Subtitle)
		}
		return buffer.String()
	}
	return (c1.Title + c1.CourseNumber + hash(c1.Sections) + strconv.Itoa(int(c1.Credits))) < (c2.Title + c2.CourseNumber + hash(c2.Sections) + strconv.Itoa(int(c2.Credits)))
}

type commentSorter struct {
	comments []RComment
}

func (a commentSorter) Len() int {
	return len(a.comments)
}

func (a commentSorter) Swap(i, j int) {
	a.comments[i], a.comments[j] = a.comments[j], a.comments[i]
}

func (a commentSorter) Less(i, j int) bool {
	return a.comments[i].Code < a.comments[j].Code
}

type MeetingByClass []RMeetingTime

func (meeting MeetingByClass) Len() int {
	return len(meeting)
}

func (meeting MeetingByClass) Swap(i, j int) {
	meeting[i], meeting[j] = meeting[j], meeting[i]
}

func (meeting MeetingByClass) Less(i, j int) bool {
	left, right := meeting[i], meeting[j]

	// Both have a day
	if left.MeetingDay != "" && right.MeetingDay != "" {
		if left.dayRank() < right.dayRank() {
			return true
		} else if left.dayRank() == right.dayRank() && left.StartTime != "" && right.StartTime != "" {
			return isAfter(left.StartTime, right.StartTime)
		}
	}

	// Neither have a day
	if left.MeetingDay == "" && right.MeetingDay == "" {
		return left.classRank() < right.classRank()
	}

	// One is missing their day
	if left.dayRank() < right.dayRank() {
		return true
	} else if left.dayRank() == right.dayRank() {
		return left.classRank() < right.classRank()
	}

	return false
}

type instructorSorter struct {
	instructors []RInstructor
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

func (meeting RMeetingTime) classRank() int {
	if meeting.isLecture() {
		return 1
	} else if meeting.isStudio() {
		return 2
	} else if meeting.isRecitation() {
		return 3
	} else if meeting.isByArrangement() {
		return 4
	} else if meeting.isLab() {
		return 5
	} else if meeting.isHybrid() {
		return 6
	} else if meeting.isOnline() {
		return 7
	} else if meeting.isInternship() {
		return 8
	} else if meeting.isSeminar() {
		return 9
	}

	return 99
}

func (meeting RMeetingTime) dayRank() int {
	switch meeting.MeetingDay {
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

func (meetingTime RMeetingTime) getMeetingHourStart() string {
	if len(meetingTime.StartTime) > 3 || len(meetingTime.EndTime) > 3 {

		meridian := ""

		if meetingTime.PmCode != "" {
			if meetingTime.PmCode == "A" {
				meridian = "AM"
			} else {
				meridian = "PM"
			}
		}
		return formatMeetingHours(meetingTime.StartTime) + " " + meridian
	}
	return ""
}

func (meetingTime RMeetingTime) getMeetingHourEnd() string {
	if len(meetingTime.StartTime) > 3 && len(meetingTime.EndTime) > 3 {
		var meridian string
		starttime := meetingTime.StartTime
		endtime := meetingTime.EndTime
		pmcode := meetingTime.PmCode

		end, _ := strconv.Atoi(endtime[:2])
		start, _ := strconv.Atoi(starttime[:2])

		if pmcode != "A" {
			meridian = "PM"
		} else if end < start {
			meridian = "PM"
		} else if endtime[:2] == "12" {
			meridian = "PM"
		} else {
			meridian = "AM"
		}

		return formatMeetingHours(meetingTime.EndTime) + " " + meridian
	}

	return ""
}

func formatMeetingHours(time string) string {
	if len(time) == 4 {
		if time[:1] == "0" {
			return time[1:2] + ":" + time[2:]
		}
		return time[:2] + ":" + time[2:]
	}
	return ""
}

func (course RCourse) metadata() (metadata []*model.Metadata) {

	if course.UnitNotes != "" {
		metadata = append(metadata, &model.Metadata{
			Title:   "School Notes",
			Content: course.UnitNotes,
		})
	}

	if course.SubjectNotes != "" {
		metadata = append(metadata, &model.Metadata{
			Title:   "Subject Notes",
			Content: course.SubjectNotes,
		})
	}

	if course.CourseNotes != "" {
		metadata = append(metadata, &model.Metadata{
			Title:   "Course Notes",
			Content: course.CourseNotes,
		})
	}

	if course.PreReqNotes != "" {
		metadata = append(metadata, &model.Metadata{
			Title:   "Prequisites",
			Content: course.PreReqNotes,
		})
	}
	if course.SynopsisURL != "" {
		metadata = append(metadata, &model.Metadata{
			Title:   "Synopsis Url",
			Content: course.SynopsisURL,
		})
	}

	return metadata
}

func filterSubjects(vs []*RSubject, f func(*RSubject) bool) []*RSubject {
	vsf := make([]*RSubject, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func filterCourses(vs []*RCourse, f func(*RCourse) bool) []*RCourse {
	vsf := make([]*RCourse, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func filterSections(vs []*RSection, f func(*RSection) bool) []*RSection {
	vsf := make([]*RSection, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func (meeting RMeetingTime) isSeminar() bool {
	return meeting.MeetingModeCode == "04"
}

func (meeting RMeetingTime) isInternship() bool {
	return meeting.MeetingModeCode == "15"
}

func (meeting RMeetingTime) isOnline() bool {
	return meeting.MeetingModeCode == "90"
}

func (meeting RMeetingTime) isHybrid() bool {
	return meeting.MeetingModeCode == "91"
}

func (meeting RMeetingTime) isByArrangement() bool {
	return meeting.MeetingModeCode == "B"
}

func (meeting RMeetingTime) isStudio() bool {
	return meeting.MeetingModeCode == "07"
}

func (meeting RMeetingTime) isLab() bool {
	return meeting.MeetingModeCode == "05"
}

func (meeting RMeetingTime) isRecitation() bool {
	return meeting.MeetingModeCode == "03"
}

func (meeting RMeetingTime) isLecture() bool {
	return meeting.MeetingModeCode == "02"
}

func (meeting RMeetingTime) day() string {
	switch meeting.MeetingDay {
	case "M":
		return "Monday"
	case "T":
		return "Tuesday"
	case "W":
		return "Wednesday"
	case "TH":
		return "Thursday"
	case "F":
		return "Friday"
	case "S":
		return "Saturday"
	case "U":
		return "Sunday"
	default:
		return ""
	}
}

func (meeting RMeetingTime) classType() string {
	if meeting.isLab() {
		return "Lab"
	} else if meeting.isStudio() {
		return "Studio"
	} else if meeting.isByArrangement() {
		return "Hours By Arrangement"
	} else if meeting.isLecture() {
		return "Lecture"
	} else if meeting.isOnline() {
		return "Online"
	} else if meeting.isInternship() {
		return "Insternship"
	} else if meeting.isSeminar() {
		return "Seminar"
	} else if meeting.isHybrid() {
		return "Hybrid"
	} else if meeting.isRecitation() {
		return "Recitation"
	} else {
		return meeting.MeetingModeDesc
	}
}

var emptyByteArray = make([]byte, 0)
var nullByte = []byte("\x00")
var headingBytes = []byte("\x01")

func trimAll(str string) string {
	str = stripSpaces(str)
	temp := []byte(str)

	// Remove NUL and Heading bytes from string, cannot be inserted into postgresql
	temp = bytes.Replace(temp, nullByte, emptyByteArray, -1)
	str = string(bytes.Replace(temp, headingBytes, emptyByteArray, -1))

	return trim(str)
}

func stripSpaces(str string) string {
	var lastRune rune
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) && unicode.IsSpace(lastRune) {
			// if the character is a space, drop it
			return -1
		}
		lastRune = r
		// else keep it in the string
		return r
	}, str)
}

// Determines if t2 is after t1
func isAfter(t1, t2 string) bool {
	if length := len(t1); length == 7 {
		t1 = "0" + t1
	} else if length == 0 {
		return false
	}
	if length := len(t2); length == 7 {
		t2 = "0" + t2
	} else if length == 0 {
		return false
	}

	if t1[:2] == "12" {
		t1 = t1[2:]
		t1 = "00" + t1
	}
	if t2[:2] == "12" {
		t2 = t2[2:]
		t2 = "00" + t2
	}

	if t1[6:] == "AM" && t2[6:] == "PM" {
		return true
	} else if t1[6:] == "PM" && t2[6:] == "AM" {
		return false
	}

	if t1[:2] == t2[:2] {
		return t1[3:5] < t2[3:5]
	}
	return t1[:2] < t2[:2]
}

func getExamCode(code string) string {
	switch code {
	case "A":
		return "By Arrangement"
	case "O":
		return "No Exam"
	case "S":
		return "Single Day starting 6:00pm or later and Saturday Courses"
	default:
		return "Group Exam"
	}
}
