package model

import (
	"bytes"
	"sort"
	"strconv"
	"strings"
	"time"
	"uct/common/model"
)

type (
	MeetingByClass []RMeetingTime

	RSubject struct {
		Name    string    `json:"description,omitempty"`
		Number  string    `json:"code,omitempty"`
		Courses []RCourse `json:"courses,omitempty"`
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
		Sections          []RSection    `json:"sections"`
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
		CampusLocation  string  `json:"campusLocation"`
		BaClassHours    string  `json:"baClassHours"`
		RoomNumber      string  `json:"roomNumber"`
		PmCode          string  `json:"pmCode"`
		CampusAbbrev    string  `json:"campusAbbrev"`
		CampusName      string  `json:"campusName"`
		MeetingDay      string  `json:"meetingDay"`
		BuildingCode    string  `json:"buildingCode"`
		StartTime       string  `json:"startTime"`
		EndTime         string  `json:"endTime"`
		PStartTime      *string `json:"-"`
		PEndTime        *string `json:"-"`
		MeetingModeDesc string  `json:"meetingModeDesc"`
		MeetingModeCode string  `json:"meetingModeCode"`
	}
)

func (course *RCourse) Clean() {
	course.Sections = FilterSections(course.Sections, func(section RSection) bool {
		return section.Printed == "Y"
	})

	m := map[string]int{}

	// Filter duplicate sections, yes it happens e.g Fall 2016 NB Biochem Engin SR DESIGN I PROJECTS
	course.Sections = FilterSections(course.Sections, func(section RSection) bool {
		key := section.Index + section.Number
		m[key]++
		return m[key] <= 1
	})

	course.Title = model.TrimAll(course.Title)

	course.CourseNumber = model.TrimAll(course.CourseNumber)

	course.CourseDescription = model.TrimAll(course.CourseDescription)

	course.CourseNotes = model.TrimAll(course.CourseNotes)

	course.SubjectNotes = model.TrimAll(course.SubjectNotes)

	course.SynopsisURL = model.TrimAll(course.SynopsisURL)

	course.PreReqNotes = model.TrimAll(course.PreReqNotes)

	course.ExpandedTitle = model.TrimAll(course.ExpandedTitle)

	if len(course.ExpandedTitle) == 0 {
		course.ExpandedTitle = course.Title
	}
}

func (section *RSection) Clean() {
	section.Subtitle = model.TrimAll(section.Subtitle)
	section.SectionNotes = model.TrimAll(section.SectionNotes)
	section.CampusCode = model.TrimAll(section.CampusCode)
	section.SpecialPermissionAddCodeDescription = model.TrimAll(section.SpecialPermissionAddCodeDescription)

	for i := range section.MeetingTimes {
		section.MeetingTimes[i].Clean()
	}

	sort.Sort(MeetingByClass(section.MeetingTimes))
	sort.Sort(instructorSorter{section.Instructor})
}

func (meeting *RMeetingTime) Clean() {
	meeting.StartTime = model.TrimAll(meeting.StartTime)
	meeting.EndTime = model.TrimAll(meeting.EndTime)

	meeting.MeetingDay = meeting.day()
	meeting.StartTime = meeting.getMeetingHourBegin()
	meeting.EndTime = meeting.getMeetingHourEnd()

	if meeting.StartTime != "" {
		t := meeting.StartTime
		meeting.PStartTime = &t
	} else {
		meeting.PStartTime = nil
	}
	if meeting.EndTime != "" {
		t := meeting.EndTime
		meeting.PEndTime = &t
	} else {
		meeting.PEndTime = nil
	}
}

func (section *RSection) Status() string {
	if section.OpenStatus {
		return model.OPEN.String()
	} else {
		return model.CLOSED.String()
	}
}

func (section RSection) instructor() (instructors []*model.Instructor) {
	for _, instructor := range section.Instructor {
		instructors = append(instructors, &model.Instructor{Name: instructor.Name})
	}
	return
}

func (section RSection) Metadata() (metadata []*model.Metadata) {

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
	Sections []RSection
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
	Courses []RCourse
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
	var hash = func(s []RSection) string {
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

func (meeting MeetingByClass) Len() int {
	return len(meeting)
}

func (meeting MeetingByClass) Swap(i, j int) {
	meeting[i], meeting[j] = meeting[j], meeting[i]
}

func (meeting MeetingByClass) Less(i, j int) bool {
	left, right := meeting[i], meeting[j]

	if left.dayRank() < right.dayRank() {
		return true
	} else if left.dayRank() == right.dayRank()  && left.StartTime != "" && right.StartTime != "" {
		return IsAfter(left.StartTime, right.StartTime)
	}

	if left.MeetingDay == "" || right.MeetingDay == "" {
		b := left.classRank() < right.classRank()
		return b
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

func (meeting RMeetingTime) Room() *string {
	if meeting.BuildingCode != "" {
		room := meeting.BuildingCode + "-" + meeting.RoomNumber
		return &room
	}
	return nil
}

func (meetingTime RMeetingTime) getMeetingHourBegin() string {
	if len(meetingTime.StartTime) > 1 || len(meetingTime.EndTime) > 1 {

		meridian := ""

		if meetingTime.PmCode != "" {
			if meetingTime.PmCode == "A" {
				meridian = "AM"
			} else {
				meridian = "PM"
			}
		}
		return FormatMeetingHours(meetingTime.StartTime) + " " + meridian
	}
	return ""
}

func (meetingTime RMeetingTime) getMeetingHourEnd() string {
	if len(meetingTime.StartTime) > 1 && len(meetingTime.EndTime) > 1 {
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

		return FormatMeetingHours(meetingTime.EndTime) + " " + meridian
	}

	return ""
}

func (meetingTime RMeetingTime) getMeetingHourBeginTime() time.Time {
	if len(model.TrimAll(meetingTime.StartTime)) > 1 || len(model.TrimAll(meetingTime.EndTime)) > 1 {

		meridian := ""

		if meetingTime.PmCode != "" {
			if meetingTime.PmCode == "A" {
				meridian = "AM"
			} else {
				meridian = "PM"
			}
		}

		kitchenTime := model.TrimAll(FormatMeetingHours(meetingTime.StartTime) + meridian)
		time, err := time.Parse(time.Kitchen, kitchenTime)
		model.CheckError(err)
		return time
	}
	return time.Unix(0, 0)
}

func (meeting RMeetingTime) Metadata() (metadata []*model.Metadata) {

	return
}

func (course RCourse) Synopsis() *string {
	if course.CourseDescription == "" {
		return nil
	} else {
		return &course.CourseDescription
	}
}

func (course RCourse) Metadata() (metadata []*model.Metadata) {

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

func FilterSubjects(vs []RSubject, f func(RSubject) bool) []RSubject {
	vsf := make([]RSubject, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func FilterCourses(vs []RCourse, f func(RCourse) bool) []RCourse {
	vsf := make([]RCourse, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func FilterSections(vs []RSection, f func(RSection) bool) []RSection {
	vsf := make([]RSection, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func AppendRSubjects(subjects []RSubject, toAppend []RSubject) []RSubject {
	for _, val := range toAppend {
		subjects = append(subjects, val)
	}
	return subjects
}

func FormatMeetingHours(time string) string {
	if len(time) > 1 {
		if time[:1] == "0" {
			return time[1:2] + ":" + time[2:]
		}
		return time[:2] + ":" + time[2:]
	}
	return ""
}

func (meetingTime RMeetingTime) getMeetingHourEndTime() time.Time {
	if len(model.TrimAll(meetingTime.StartTime)) > 1 || len(model.TrimAll(meetingTime.EndTime)) > 1 {
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
			meridian = "AM"
		} else {
			meridian = "AM"
		}

		time, err := time.Parse(time.Kitchen, FormatMeetingHours(meetingTime.EndTime)+meridian)
		model.CheckError(err)
		return time
	}
	return time.Unix(0, 0)
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

func (meeting RMeetingTime) DayPointer() *string {
	if meeting.MeetingDay == "" {
		return nil
	} else {
		return &meeting.MeetingDay
	}
}

func (meeting RMeetingTime) ClassType() *string {
	mtype := ""
	if meeting.isLab() {
		mtype = "Lab"
	} else if meeting.isStudio() {
		mtype = "Studio"
	} else if meeting.isByArrangement() {
		mtype = "Hours By Arrangement"
	} else if meeting.isLecture() {
		mtype = "Lecture"
	} else if meeting.isOnline() {
		mtype = "Online"
	} else if meeting.isInternship() {
		mtype = "Insternship"
	} else if meeting.isSeminar() {
		mtype = "Seminar"
	} else if meeting.isHybrid() {
		mtype = "Hybrid"
	} else if meeting.isRecitation() {
		mtype = "Recitation"
	} else {
		mtype = meeting.MeetingModeDesc
	}
	if mtype == "" {
		return nil
	} else {
		return &mtype
	}
}

// Determines if one date string is after another e.g 1:00PM is after 9:00AM
func IsAfter(t1, t2 string) bool {
	if l1 := len(t1); l1 == 7 {
		t1 = "0" + t1
	} else if l1 == 0 {
		return false
	}
	if l2 := len(t2); l2 == 7 {
		t2 = "0" + t2
	} else if l2 == 0 {
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
