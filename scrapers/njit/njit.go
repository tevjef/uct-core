package main

import (
	"github.com/Sirupsen/logrus"
	"sort"
	"strconv"
	"time"
	"uct/common/model"
	"strings"
)

type NSemesters struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type NSubject struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	year        int
	season      string
	courses     []*NCourse
}

type NSearch struct {
	Success              bool       `json:"success"`
	TotalCount           int        `json:"totalCount"`
	Data                 []*NCourse `json:"data"`
	PageOffset           int        `json:"pageOffset"`
	PageMaxSize          int        `json:"pageMaxSize"`
	SectionsFetchedCount int        `json:"sectionsFetchedCount"`
	PathMode             string     `json:"pathMode"`
}

type NCourse struct {
	ID                      int               `json:"id"`
	Term                    string            `json:"term"`
	TermDesc                string            `json:"termDesc"`
	CourseReferenceNumber   string            `json:"courseReferenceNumber"`
	PartOfTerm              string            `json:"partOfTerm"`
	CourseNumber            string            `json:"courseNumber"`
	Subject                 string            `json:"subject"`
	SubjectDescription      string            `json:"subjectDescription"`
	SequenceNumber          string            `json:"sequenceNumber"`
	CampusDescription       string            `json:"campusDescription"`
	ScheduleTypeDescription string            `json:"scheduleTypeDescription"`
	CourseTitle             string            `json:"courseTitle"`
	CreditHours             float64           `json:"creditHours"`
	MaximumEnrollment       int               `json:"maximumEnrollment"`
	Enrollment              int               `json:"enrollment"`
	SeatsAvailable          int               `json:"seatsAvailable"`
	WaitCapacity            int               `json:"waitCapacity"`
	WaitCount               int               `json:"waitCount"`
	WaitAvailable           int               `json:"waitAvailable"`
	OpenSection             bool              `json:"openSection"`
	IsSectionLinked         bool              `json:"isSectionLinked"`
	SubjectCourse           string            `json:"subjectCourse"`
	Faculty                 []Faculty         `json:"faculty"`
	MeetingsFaculty         []MeetingsFaculty `json:"meetingsFaculty"`
	status                  string
	meetingTimes            []*MeetingTime
}

type MeetingsFaculty struct {
	Category              string      `json:"category"`
	Class                 string      `json:"class"`
	CourseReferenceNumber string      `json:"courseReferenceNumber"`
	Faculty               []Faculty   `json:"faculty"`
	MeetingTime           MeetingTime `json:"meetingTime"`
	Term                  string      `json:"term"`
}

type MeetingTime struct {
	BeginTime             string  `json:"beginTime"`
	Building              string  `json:"building"`
	BuildingDescription   string  `json:"buildingDescription"`
	Campus                string  `json:"campus"`
	CampusDescription     string  `json:"campusDescription"`
	Category              string  `json:"category"`
	Class                 string  `json:"class"`
	CourseReferenceNumber string  `json:"courseReferenceNumber"`
	CreditHourSession     float64 `json:"creditHourSession"`
	EndDate               string  `json:"endDate"`
	EndTime               string  `json:"endTime"`
	Friday                bool    `json:"friday"`
	HoursWeek             float64 `json:"hoursWeek"`
	MeetingScheduleType   string  `json:"meetingScheduleType"`
	Monday                bool    `json:"monday"`
	Room                  string  `json:"room"`
	Saturday              bool    `json:"saturday"`
	StartDate             string  `json:"startDate"`
	Sunday                bool    `json:"sunday"`
	Term                  string  `json:"term"`
	Thursday              bool    `json:"thursday"`
	Tuesday               bool    `json:"tuesday"`
	Wednesday             bool    `json:"wednesday"`
	day                   string
}

type Faculty struct {
	BannerID              string      `json:"bannerId"`
	Category              interface{} `json:"category"`
	Class                 string      `json:"class"`
	CourseReferenceNumber string      `json:"courseReferenceNumber"`
	DisplayName           string      `json:"displayName"`
	EmailAddress          string      `json:"emailAddress"`
	PrimaryIndicator      bool        `json:"primaryIndicator"`
	Term                  string      `json:"term"`
}

func (s *NSubject) clean() {
	s.Description = strings.NewReplacer("&amp;", "&", "&#39;", "'").Replace(s.Description)
}

func (c *NCourse) clean() {
	c.CourseTitle = strings.NewReplacer("&amp;", "&", "&#39;", "'").Replace(c.CourseTitle)
	c.status = parseStatus(*c)
	c.meetingTimes = c.parseMeetingFaculty(c.MeetingsFaculty)
	for i := range c.meetingTimes {
		c.meetingTimes[i].clean()
	}

	sort.Sort(meetingByClass{c.meetingTimes})
	sort.Sort(instructorSorter{c.Faculty})
}

func (m *MeetingTime) clean() {
	m.day = parseDay(*m)
	m.MeetingScheduleType = parseClassType(*m)
	m.Room = parseRoom(*m)
	m.BeginTime = formatMeetingHour(m.BeginTime)
	m.EndTime = formatMeetingHour(m.EndTime)
}

func parseClassType(meetingTime MeetingTime) string {
	switch meetingTime.MeetingScheduleType {
	case "LEC":
		return "Lecture"
	case "STU":
		return "Studio"
	case "LAB":
		return "Lab"
	case "ADV":
		return "Advising"
	default:
		return meetingTime.MeetingScheduleType
	}
}

func (meetingTime MeetingTime) classRank() int {
	if meetingTime.MeetingScheduleType == "Lecture" {
		return 1
	} else if meetingTime.MeetingScheduleType == "Studio" {
		return 2
	} else if meetingTime.MeetingScheduleType == "Lab" {
		return 3
	} else if meetingTime.MeetingScheduleType == "Advising" {
		return 4
	}
	return 99
}

func formatMeetingHour(timeStr string) string {
	if len(timeStr) == 4 {
		var meridan string
		hourStr := timeStr[:2]
		minuteStr := timeStr[2:]

		hour, _ := strconv.Atoi(hourStr)
		hourStr = strconv.Itoa(hour)

		// get meridian
		if hour >= 12 {
			meridan = "PM"
		} else {
			meridan = "AM"
		}

		if hour > 12 {
			hourStr = strconv.Itoa(hour - 12)
		}

		return hourStr + ":" + minuteStr + " " + meridan
	}

	return ""
}

func parseStatus(c NCourse) string {
	if c.OpenSection {
		return model.Open.String()
	} else {
		return model.Closed.String()
	}
	return ""
}

func parseRoom(m MeetingTime) string {
	if m.Room == "" {
		return m.Building
	} else if m.Building != "" {
		return m.Building + " " + m.Room
	}
	return ""
}

func parseDay(m MeetingTime) string {
	if m.Sunday {
		return time.Sunday.String()
	} else if m.Monday {
		return time.Monday.String()
	} else if m.Tuesday {
		return time.Tuesday.String()
	} else if m.Wednesday {
		return time.Wednesday.String()
	} else if m.Thursday {
		return time.Thursday.String()
	} else if m.Friday {
		return time.Friday.String()
	} else if m.Saturday {
		return time.Saturday.String()
	}
	return ""
}

func (c *NCourse) parseMeetingFaculty(meetingFaculty []MeetingsFaculty) (meetingTimes []*MeetingTime) {
	for i := range meetingFaculty {
		meetingTimes = append(meetingTimes, c.extractMeetings(&meetingFaculty[i].MeetingTime)...)
	}
	return
}

func (c *NCourse) extractMeetings(njitMeeting *MeetingTime) (meetings []*MeetingTime) {
	meeting := *njitMeeting
	meeting.Sunday, meeting.Monday, meeting.Tuesday, meeting.Wednesday, meeting.Thursday, meeting.Friday, meeting.Saturday = false, false, false, false, false, false, false
	if njitMeeting.Sunday {
		n := meeting
		n.Sunday = true
		meetings = append(meetings, &n)
	}

	if njitMeeting.Monday {
		n := meeting
		n.Monday = true
		meetings = append(meetings, &n)
	}

	if njitMeeting.Tuesday {
		n := meeting
		n.Tuesday = true
		meetings = append(meetings, &n)
	}

	if njitMeeting.Wednesday {
		n := meeting
		n.Wednesday = true
		meetings = append(meetings, &n)
	}

	if njitMeeting.Thursday {
		n := meeting
		n.Thursday = true
		meetings = append(meetings, &n)
	}

	if njitMeeting.Friday {
		n := meeting
		n.Friday = true
		meetings = append(meetings, &n)
	}

	if njitMeeting.Saturday {
		n := meeting
		n.Saturday = true
		meetings = append(meetings, &n)
	}

	if !njitMeeting.Sunday && !njitMeeting.Monday && !njitMeeting.Tuesday && !njitMeeting.Wednesday && !njitMeeting.Thursday && !njitMeeting.Friday && !njitMeeting.Saturday {
		n := meeting
		meetings = append(meetings, &n)
	}

	return
}

func collapseCourses(njitCourses []*NCourse) (courses [][]*NCourse) {
	a := map[string][]*NCourse{}

	for i := range njitCourses {
		key := njitCourses[i].CourseTitle
		a[key] = append(a[key], njitCourses[i])
	}

	for _, val := range a {
		sort.Sort(CourseSorter{val})
		courses = append(courses, val)
	}

	sort.Sort(multiCourse{courses})
	return
}

func cleanCourseList(njitCourses []*NCourse) (courses []*NCourse) {
	hash := func(c NCourse) string {
		return c.CourseTitle + c.CampusDescription + c.SubjectCourse + c.SubjectDescription + c.CourseNumber + c.SequenceNumber + c.CourseReferenceNumber + c.Term
	}

	var courseMap = make(map[string]*NCourse)

	// Find duplicate courses
	for i := range njitCourses {
		c := njitCourses[i]

		// Clean
		c.clean()

		// Create hash
		h := hash(*c)
		if courseMap[h] == nil {
			courseMap[h] = c
		} else {
			logrus.Debugln("Found duplicate", h)
		}
	}

	for _, v := range courseMap {
		courses = append(courses, v)
	}

	sort.Sort(CourseSorter{courses})

	return
}

type CourseSorter struct {
	Courses []*NCourse
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

	hash := func(c *NCourse) string {
		return c.CourseNumber + c.CourseTitle + c.SequenceNumber + c.CourseReferenceNumber
	}

	return hash(c1) < hash(c2)
}

type multiCourse struct {
	Courses [][]*NCourse
}

func (a multiCourse) Len() int {
	return len(a.Courses)
}

func (a multiCourse) Swap(i, j int) {
	a.Courses[i], a.Courses[j] = a.Courses[j], a.Courses[i]
}

func (a multiCourse) Less(i, j int) bool {
	c1 := a.Courses[i][0]
	c2 := a.Courses[j][0]

	hash := func(c *NCourse) string {
		return c.CourseNumber + c.CourseTitle + c.SequenceNumber + c.CourseReferenceNumber
	}

	return hash(c1) < hash(c2)
}

type meetingByClass struct {
	Meeting []*MeetingTime
}

func (meeting meetingByClass) Len() int {
	return len(meeting.Meeting)
}

func (meeting meetingByClass) Swap(i, j int) {
	meeting.Meeting[i], meeting.Meeting[j] = meeting.Meeting[j], meeting.Meeting[i]
}

func (meeting meetingByClass) Less(i, j int) bool {
	left, right := meeting.Meeting[i], meeting.Meeting[j]

	// Both have a day
	if left.day != "" && right.day != "" {
		if left.dayRank() < right.dayRank() {
			return true
		} else if left.dayRank() == right.dayRank() && left.BeginTime != "" && right.EndTime != "" {
			return isAfter(left.BeginTime, right.EndTime)
		}
	}

	// Neither have a day
	if left.day == "" && right.day == "" {
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
	instructors []Faculty
}

func (a instructorSorter) Len() int {
	return len(a.instructors)
}

func (a instructorSorter) Swap(i, j int) {
	a.instructors[i], a.instructors[j] = a.instructors[j], a.instructors[i]
}

func (a instructorSorter) Less(i, j int) bool {
	return a.instructors[i].DisplayName < a.instructors[j].DisplayName
}

func (meeting MeetingTime) dayRank() int {
	switch meeting.day {
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
