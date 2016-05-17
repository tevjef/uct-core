package common

import (
	"bytes"
	"fmt"
	"golang.org/x/exp/utf8string"
	"hash/fnv"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type (
	MeetingByDay    []Meeting
	SectionByNumber []Section
	CourseByName    []Course
	SubjectByName   []Subject

	ResolvedSemester struct {
		Last    Semester `json:"last,omitempty"`
		Current Semester `json:"current,omitempty"`
		Next    Semester `json:"next,omitempty"`
	}

	Period int
	Status int

	PostgresNotify struct {
		Payload    string     `json:"topic_name,omitempty"`
		Status     string     `json:"status,omitempty"`
		Max        float64    `json:"max,omitempty"`
		Now        float64    `json:"now,omitempty"`
		University University `json:"university,omitempty"`
	}

	GCMMessage struct {
		To              string              `json:"to,omitempty"`
		RegistrationIds []string            `json:"registration_ids,omitempty"`
		CollapseKey     string              `json:"collapse_key,omitempty"`
		Priority        string              `json:"priority,omitempty"`
		DelayWhileIdle  bool                `json:"delay_while_idle,omitempty"`
		TimeToLive      int                 `json:"time_to_live,omitempty"`
		Data            Data                `json:"data,omitempty"`
		Notification    *MobileNotification `json:"notification,omitempty"`
		DryRun          bool                `json:"dry_run,omitempty"`
	}

	GCMResponse struct {
		MessageId int64  `json:"message_id,omitempty"`
		Error     string `json:"error, omitempty"`
	}

	Data struct {
		Message string `json:"message,omitempty"`
	}

	MobileNotification struct {
		Title string `json:"title,omitempty"`
		Body  string `json:"body,omitempty"`
		Icon  string `json:"icon,omitempty"`
	}
)

const (
	SEM_FALL Period = iota
	SEM_SPRING
	SEM_SUMMER
	SEM_WINTER
	START_FALL
	START_SPRING
	START_SUMMER
	START_WINTER
	END_FALL
	END_SPRING
	END_SUMMER
	END_WINTER
)

var period = [...]string{
	"fall",
	"spring",
	"summer",
	"winter",
	"start_fall",
	"start_spring",
	"start_summer",
	"start_winter",
	"end_fall",
	"end_spring",
	"end_summer",
	"end_winter",
}

func (s Period) String() string {
	return period[s]
}

const (
	FALL Season = iota
	SPRING
	SUMMER
	WINTER
)

const (
	OPEN Status = 1 + iota
	CLOSED
	CANCELLED
)

var status = [...]string{
	"Open",
	"Closed",
	"Cancelled",
}

func (s Status) String() string {
	return status[s-1]
}

func (s Subject) hash() string {
	var buffer bytes.Buffer
	for _, course := range s.Courses {
		buffer.WriteString("a" + course.Hash)
	}

	return fmt.Sprintf("%x", fnv.New64a().Sum([]byte(s.Name+s.Number+s.Season+s.Year+buffer.String())))
}

func (c Course) hash() string {
	var buffer bytes.Buffer
	for _, section := range c.Sections {
		buffer.WriteString(section.CallNumber + section.Number)
	}
	return fmt.Sprintf("%x", fnv.New64a().Sum([]byte(c.Name+c.Number+buffer.String())))
}

func (u *University) Validate() {
	// Name
	if u.Name == "" {
		log.Panic("University name == is empty")
	}
	u.Name = strings.Replace(u.Name, "_", " ", -1)
	u.Name = strings.Replace(u.Name, ".", " ", -1)
	u.Name = strings.Replace(u.Name, "~", " ", -1)
	u.Name = strings.Replace(u.Name, "%", " ", -1)

	// Abbr
	if u.Abbr == "" {
		regex, err := regexp.Compile("[^A-Z]")
		CheckError(err)
		u.Abbr = trim(regex.ReplaceAllString(u.Name, ""))
	}

	// Homepage
	if u.HomePage == "" {
		log.Panic("HomePage == is empty")
	}
	u.HomePage = trim(u.HomePage)
	nUrl, err := url.ParseRequestURI(u.HomePage)
	CheckError(err)
	u.HomePage = nUrl.String()

	// RegistrationPage
	if u.RegistrationPage == "" {
		log.Panic("RegistrationPage == is empty")
	}
	u.RegistrationPage = trim(u.RegistrationPage)
	nUrl, err = url.ParseRequestURI(u.RegistrationPage)
	CheckError(err)
	u.RegistrationPage = nUrl.String()

	// MainColor
	if u.MainColor == "" {
		u.MainColor = "00000000"
	}

	// AccentColor
	if u.AccentColor == "" {
		u.AccentColor = "00000000"
	}

	// Registration
	if len(u.Registrations) != 12 {
		log.Panic("Registration != 12 ")
	}

	// TopicName
	regex, err := regexp.Compile("\\s+")
	CheckError(err)
	u.TopicName = regex.ReplaceAllString(u.Name, ".")
	regex, err = regexp.Compile("[^A-Za-z.]")
	CheckError(err)
	u.TopicName = trim(regex.ReplaceAllString(u.TopicName, ""))
	if len(u.TopicName) > 26 {
		u.TopicName = u.TopicName[:25]
	}
}

func (sub *Subject) Validate() {
	// Name
	if sub.Name == "" {
		log.Panic("Subject name == is empty")
	}
	sub.Name = strings.Title(strings.ToLower(sub.Name))
	sub.Name = strings.Replace(sub.Name, "_", " ", -1)
	sub.Name = strings.Replace(sub.Name, ".", " ", -1)
	sub.Name = strings.Replace(sub.Name, "~", " ", -1)
	sub.Name = strings.Replace(sub.Name, "%", " ", -1)

	// Abbr
	/*if sub.Number == "" {
		regex, err := regexp.Compile("[^A-Z]")
		CheckError(err)
		sub.Number = trim(regex.ReplaceAllString(sub.Name, ""))
	}*/
	if len(sub.Number) > 3 {
		sub.Number = sub.Number[:3]
	}

	if len(sub.Courses) == 0 {
		Log("No courses in subject", sub)
	}

	// TopicName
	regex, err := regexp.Compile("\\s+")
	CheckError(err)
	sub.TopicName = regex.ReplaceAllString(sub.Name, ".")
	regex, err = regexp.Compile("[^A-Za-z.]")
	CheckError(err)
	sub.TopicName = trim(regex.ReplaceAllString(sub.TopicName, "."))

	sub.Hash = sub.hash()
}

func (course *Course) Validate() {
	// Name

	if course.Name == "" {
		log.Panic("Course name == is empty", course)
	}

	course.Name = strings.Title(strings.ToLower(course.Name))
	course.Name = strings.Replace(course.Name, "_", " ", -1)
	course.Name = strings.Replace(course.Name, ".", " ", -1)
	course.Name = strings.Replace(course.Name, "~", " ", -1)
	course.Name = strings.Replace(course.Name, "%", " ", -1)

	course.Name = TrimAll(course.Name)
	// Number
	if course.Number == "" {
		log.Panic("Number == is empty")
	}

	// Synopsis
	if course.Synopsis != nil {
		regex, err := regexp.Compile("\\s\\s+")
		CheckError(err)
		temp := regex.ReplaceAllString(*course.Synopsis, " ")
		course.Synopsis = &temp
		temp = utf8string.NewString(*course.Synopsis).String()
		course.Synopsis = &temp

	}

	// TopicName
	regex, err := regexp.Compile("\\s+")
	CheckError(err)
	course.TopicName = regex.ReplaceAllString(course.Name, ".")
	regex, err = regexp.Compile("[^A-Za-z.]")
	CheckError(err)
	course.TopicName = trim(regex.ReplaceAllString(course.TopicName, "."))

	course.Hash = course.hash()
}

func (section *Section) Validate() {
	// Number
	if section.Number == "" {
		log.Panic("Number == is empty")
	}
	section.Number = trim(section.Number)

	// Call Number
	if section.CallNumber == "" {
		log.Panic("CallNumber == is empty")
	}
	section.CallNumber = trim(section.CallNumber)

	// Status
	if section.Status == "" {
		log.Panic("Status == is empty")
	}

	// Max
	if section.Max == 0 {
		section.Max = 1
		if section.Status == OPEN.String() {
			section.Now = 1
		} else {
			section.Now = 0
		}
	}

	// Credits
	if section.Credits == "" {
		log.Panic("Credits == is empty")
	}
}

func (meeting *Meeting) Validate() {

	if meeting.StartTime == "" {
		meeting.StartTime = "00:00 AM"
		meeting.EndTime = "00:00 AM"
	}

	/*meeting.StartTime.String = TrimAll(meeting.StartTime.String)
	meeting.EndTime.String = TrimAll(meeting.EndTime.String)*/
}

func (instructor *Instructor) Validate() {
	// Name
	if instructor.Name == "" {
		log.Panic("Instructor name == is empty")
	}
	instructor.Name = trim(instructor.Name)
}

func (book *Book) Validate() {
	// Title
	if book.Title == "" {
		log.Panic("Title  == is empty")
	}
	book.Title = trim(book.Title)

	// Url
	if book.Url == "" {
		log.Panic("Url == is empty")
	}
	book.Url = trim(book.Url)
	url, err := url.ParseRequestURI(book.Url)
	CheckError(err)
	book.Url = url.String()
}

func (metaData *Metadata) Validate() {
	// Title
	if metaData.Title == "" {
		log.Panic("Title == is empty")
	}
	metaData.Title = trim(metaData.Title)

	// Content
	if metaData.Content == "" {
		log.Panic("Content == is empty")
	}
	metaData.Content = trim(metaData.Content)
}

func (r Registration) month() time.Month {
	return time.Unix(r.PeriodDate, 0).Month()
}

func (r Registration) day() int {
	return time.Unix(r.PeriodDate, 0).Day()
}

func (r Registration) dayOfYear() int {
	return time.Unix(r.PeriodDate, 0).YearDay()
}

func (r Registration) season() string {
	switch r.Period {
	case SEM_FALL.String():
		return FALL.String()
	case SEM_SPRING.String():
		return SPRING.String()
	case SEM_SUMMER.String():
		return SUMMER.String()
	case SEM_WINTER.String():
		return WINTER.String()
	default:
		return SUMMER.String()
	}
}

func ResolveSemesters(t time.Time, registration []Registration) ResolvedSemester {
	month := t.Month()
	day := t.Day()
	year := t.Year()

	yearDay := t.YearDay()

	//var springReg = registration[SEM_SPRING];
	var winterReg = registration[SEM_WINTER]
	//var summerReg = registration[SEM_SUMMER];
	//var fallReg  = registration[SEM_FALL];
	var startFallReg = registration[START_FALL]
	var startSpringReg = registration[START_SPRING]
	var endSummerReg = registration[END_SUMMER]
	//var startSummerReg  = registration[START_SUMMER];

	fall := Semester{
		Year:   int32(year),
		Season: FALL}

	winter := Semester{
		Year:   int32(year),
		Season: WINTER}

	spring := Semester{
		Year:   int32(year),
		Season: SPRING}

	summer := Semester{
		Year:   int32(year),
		Season: SUMMER}

	// Spring: Winter - StartFall
	if (month >= winterReg.month() && day >= winterReg.day()) || (month <= startFallReg.month() && day < startFallReg.day()) {
		if winterReg.month()-month <= 0 {
			spring.Year = spring.Year + 1
			summer.Year = summer.Year + 1
		} else {
			winter.Year = winter.Year - 1
			fall.Year = fall.Year - 1
		}
		Log("Spring: Winter - StartFall ", winterReg.month(), winterReg.day(), "--", startFallReg.month(), startFallReg.day(), "--", month, day)

		return ResolvedSemester{
			Last:    winter,
			Current: spring,
			Next:    summer}

	} else if yearDay >= startFallReg.dayOfYear() && yearDay < endSummerReg.dayOfYear() {
		Log("StartFall: StartFall -- EndSummer ", startFallReg.dayOfYear(), "--", endSummerReg.dayOfYear(), "--", yearDay)
		return ResolvedSemester{
			Last:    spring,
			Current: summer,
			Next:    fall,
		}
	} else if yearDay >= endSummerReg.dayOfYear() && yearDay < startSpringReg.dayOfYear() {

		Log("Fall: EndSummer -- StartSpring ", endSummerReg.dayOfYear(), "--", yearDay < startSpringReg.dayOfYear(), "--", yearDay)

		return ResolvedSemester{
			Last:    summer,
			Current: fall,
			Next:    winter,
		}
	} else if yearDay >= startSpringReg.dayOfYear() && yearDay < winterReg.dayOfYear() {
		spring.Year = spring.Year + 1
		Log("StartSpring: StartSpring -- Winter ", startSpringReg.dayOfYear(), "--", winterReg.dayOfYear(), "--", yearDay)

		return ResolvedSemester{
			Last:    fall,
			Current: winter,
			Next:    spring,
		}
	}

	return ResolvedSemester{}
}

func (meeting Meeting) dayRank() int {
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

func (a MeetingByDay) Len() int {
	return len(a)
}

func (a MeetingByDay) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a MeetingByDay) Less(i, j int) bool {
	return a[i].dayRank() < a[j].dayRank()
}

func (a SectionByNumber) Len() int {
	return len(a)
}

func (a SectionByNumber) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a SectionByNumber) Less(i, j int) bool {
	return strings.Compare(a[i].Number, a[j].Number) < 0
}

func (a CourseByName) Len() int {
	return len(a)
}

func (a CourseByName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a CourseByName) Less(i, j int) bool {
	return strings.Compare(a[i].Name, a[j].Name) < 0
}

func (a SubjectByName) Len() int {
	return len(a)
}

func (a SubjectByName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a SubjectByName) Less(i, j int) bool {
	return strings.Compare(a[i].Name, a[j].Name) < 0
}
