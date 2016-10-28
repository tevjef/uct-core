package main

import (
	"time"
	"strconv"
	"uct/common/model"
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
	courses     []NCourse
}

type NSearch struct {
	Success              bool      `json:"success"`
	TotalCount           int       `json:"totalCount"`
	Data                 []NCourse `json:"data"`
	PageOffset           int       `json:"pageOffset"`
	PageMaxSize          int       `json:"pageMaxSize"`
	SectionsFetchedCount int       `json:"sectionsFetchedCount"`
	PathMode             string    `json:"pathMode"`
}

type NCourse struct {
	ID                      int    `json:"id"`
	Term                    string `json:"term"`
	TermDesc                string `json:"termDesc"`
	CourseReferenceNumber   string `json:"courseReferenceNumber"`
	PartOfTerm              string `json:"partOfTerm"`
	CourseNumber            string `json:"courseNumber"`
	Subject                 string `json:"subject"`
	SubjectDescription      string `json:"subjectDescription"`
	SequenceNumber          string `json:"sequenceNumber"`
	CampusDescription       string `json:"campusDescription"`
	ScheduleTypeDescription string `json:"scheduleTypeDescription"`
	CourseTitle             string `json:"courseTitle"`
	CreditHours             int    `json:"creditHours"`
	MaximumEnrollment       int    `json:"maximumEnrollment"`
	Enrollment              int    `json:"enrollment"`
	SeatsAvailable          int    `json:"seatsAvailable"`
	WaitCapacity            int    `json:"waitCapacity"`
	WaitCount               int    `json:"waitCount"`
	WaitAvailable           int    `json:"waitAvailable"`
	OpenSection             bool   `json:"openSection"`
	IsSectionLinked         bool              `json:"isSectionLinked"`
	SubjectCourse           string            `json:"subjectCourse"`
	Faculty                 []Faculty         `json:"faculty"`
	MeetingsFaculty         []MeetingsFaculty `json:"meetingsFaculty"`
	status                  string
	meetingTimes []MeetingTime
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
	day string
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

func (c *NCourse) clean() {
	c.status = parseStatus(*c)
	c.meetingTimes = parseMeetingFaculty(c.MeetingsFaculty)
}

func (m *MeetingTime) clean() {
	m.day = parseDay(*m)
	m.Room = parseRoom(*m)
	m.StartDate = formatMeetingHour(m.StartDate)
	m.EndTime = formatMeetingHour(m.EndTime)
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

func parseMeetingFaculty(meetingFaculty []MeetingsFaculty) (meetingTimes []MeetingTime) {
	for _, val := range meetingFaculty {
		meetingTimes = append(meetingTimes, extractMeetings(val.MeetingTime)...)
	}
	return
}

func extractMeetings(njitMeeting MeetingTime) (meetings []MeetingTime) {
	meeting := njitMeeting
	meeting.Sunday, meeting.Monday, meeting.Tuesday, meeting.Wednesday, meeting.Thursday, meeting.Friday, meeting.Saturday = false, false, false, false, false, false, false
	if meeting.Sunday {
		n := meeting
		n.Sunday = true
		meetings = append(meetings, n)
	}

	if meeting.Monday {
		n := meeting
		n.Monday = true
		meetings = append(meetings, n)
	}

	if meeting.Tuesday {
		n := meeting
		n.Tuesday = true
		meetings = append(meetings, n)
	}

	if meeting.Wednesday {
		n := meeting
		n.Wednesday = true
		meetings = append(meetings, n)
	}

	if meeting.Thursday {
		n := meeting
		n.Thursday = true
		meetings = append(meetings, n)
	}

	if meeting.Friday {
		n := meeting
		n.Friday = true
		meetings = append(meetings, n)
	}

	if meeting.Saturday {
		n := meeting
		n.Saturday = true
		meetings = append(meetings, n)
	}

	return
}

func collapseCourses(njitCourses []NCourse) (courses [][]NCourse) {
	a := map[string][]NCourse{}

	for i := range njitCourses {
		key := njitCourses[i].CourseTitle
		a[key] = append(a[key], njitCourses[i])
	}

	for _, val := range a {
		courses = append(courses, val)
	}

	return
}