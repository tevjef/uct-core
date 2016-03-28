package common

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"golang.org/x/exp/utf8string"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type (
	/*	Subjects []Subject
		Courses  []Course
		Sections []Section
		Meetings []Meeting*/

	MeetingByDay    []Meeting
	SectionByNumber []Section
	CourseByName    []Course
	SubjectByName   []Subject

	University struct {
		Id               int64          `json:"-" db:"id"`
		Name             string         `json:"name,omitempty" db:"name"`
		Abbr             string         `json:"abbr,omitempty" db:"abbr"`
		HomePage         string         `json:"home_page,omitempty" db:"home_page"`
		RegistrationPage string         `json:"registration_page,omitempty" db:"registration_page"`
		MainColor        string         `json:"main_color,omitempty" db:"main_color"`
		AccentColor      string         `json:"accent_color,omitempty" db:"accent_color"`
		TopicName        string         `json:"topic_name,omitempty" db:"topic_name"`
		Subjects         []Subject      `json:"subjects,omitempty"`
		Registrations    []Registration `json:"registration,omitempty"`
		Metadata         []Metadata     `json:"metadata,omitempty"`
		CreatedAt        time.Time      `json:"-" db:"created_at"`
		UpdatedAt        time.Time      `json:"-" db:"updated_at"`
	}

	// Sort by name
	Subject struct {
		Id           int64      `json:"-" db:"id"`
		UniversityId int64      `json:"-" db:"university_id"`
		Name         string     `json:"name,omitempty" db:"name"`
		Number       string     `json:"number,omitempty" db:"number"`
		Season       string     `json:"season" db:"season"`
		Year         int        `json:"year,omitempty" db:"year"`
		Hash         string     `json:"hash,omitempty" db:"hash"`
		TopicName    string     `json:"topic_name,omitempty" db:"topic_name"`
		Courses      []Course   `json:"courses,omitempty"`
		Metadata     []Metadata `json:"metadata,omitempty"`
		CreatedAt    time.Time  `json:"-" db:"created_at"`
		UpdatedAt    time.Time  `json:"-" db:"updated_at"`
	}

	// Sort by name
	Course struct {
		Id        int64      `json:"-" db:"id"`
		SubjectId int64      `json:"-" db:"subject_id"`
		Name      string     `json:"name,omitempty" db:"name"`
		Number    string     `json:"number,omitempty" db:"number"`
		Synopsis  *string    `json:"synopsis,omitempty" db:"synopsis"`
		Hash      string     `json:"hash,omitempty" db:"hash"`
		TopicName string     `json:"topic_name,omitempty" db:"topic_name"`
		Sections  []Section  `json:"sections,omitempty"`
		Metadata  []Metadata `json:"metadata,omitempty"`
		CreatedAt time.Time  `json:"-" db:"created_at"`
		UpdatedAt time.Time  `json:"-" db:"updated_at"`
	}

	// Sort by number
	Section struct {
		Id          int64        `json:"-" db:"id"`
		CourseId    int64        `json:"-" db:"course_id"`
		Number      string       `json:"number,omitempty" db:"number"`
		CallNumber  string       `json:"call_number,omitempty" db:"call_number"`
		Max         float64      `json:"max,omitempty" db:"max"`
		Now         float64      `json:"now,omitempty" db:"now"`
		Status      string       `json:"status,omitempty" db:"status"`
		Credits     string       `json:"credits,omitempty" db:"credits"`
		TopicName   string       `json:"topic_name,omitempty" db:"topic_name"`
		Meetings    []Meeting    `json:"meeting,omitempty"`
		Instructors []Instructor `json:"instructors,omitempty"`
		Books       []Book       `json:"books,omitempty"`
		Metadata    []Metadata   `json:"metadata,omitempty"`
		CreatedAt   time.Time    `json:"-" db:"created_at"`
		UpdatedAt   time.Time    `json:"-" db:"updated_at"`
	}

	// Sort by day
	Meeting struct {
		Id        int64      `json:"-" db:"id"`
		SectionId int64      `json:"-" db:"section_id"`
		Room      *string    `json:"room,omitempty" db:"room"`
		Day       *string    `json:"day,omitempty" db:"day"`
		StartTime string     `json:"start_time,omitempty" db:"start_time"`
		EndTime   string     `json:"end_time,omitempty" db:"end_time"`
		Index     int        `json:"index" db:"index"`
		Metadata  []Metadata `json:"metadata,omitempty"`
		CreatedAt time.Time  `json:"-" db:"created_at"`
		UpdatedAt time.Time  `json:"-" db:"updated_at"`
	}

	Instructor struct {
		Id        int64     `json:"-" db:"id"`
		SectionId int64     `json:"-" db:"section_id"`
		Name      string    `json:"name,omitempty" db:"name"`
		CreatedAt time.Time `json:"-" db:"created_at"`
		UpdatedAt time.Time `json:"-" db:"updated_at"`
	}

	Book struct {
		Id        int64     `json:"-" db:"id"`
		SectionId int64     `json:"-" db:"section_id"`
		Title     string    `json:"title,omitempty" db:"title"`
		Url       string    `json:"url,omitempty" db:"url"`
		CreatedAt time.Time `json:"-" db:"created_at"`
		UpdatedAt time.Time `json:"-" db:"updated_at"`
	}

	Metadata struct {
		Id           int64     `json:"-" db:"id"`
		UniversityId *int64    `json:"-" db:"university_id"`
		SubjectId    *int64    `json:"-" db:"subject_id"`
		CourseId     *int64    `json:"-" db:"course_id"`
		SectionId    *int64    `json:"-" db:"section_id"`
		MeetingId    *int64    `json:"-" db:"meeting_id"`
		Title        string    `json:"title,omitempty" db:"title"`
		Content      string    `json:"content,omitempty" db:"content"`
		CreatedAt    time.Time `json:"-" db:"created_at"`
		UpdatedAt    time.Time `json:"-" db:"updated_at"`
	}

	Registration struct {
		Id           int64     `json:"-" db:"id"`
		UniversityId int64     `json:"-" db:"university_id"`
		Period       string    `json:"period,omitempty" db:"period"`
		PeriodDate   time.Time `json:"period_date,omitempty" db:"period_date"`
		CreatedAt    time.Time `json:"-" db:"created_at"`
		UpdatedAt    time.Time `json:"-" db:"updated_at"`
	}

	Semester struct {
		Year   int
		Season Season
	}

	ResolvedSemester struct {
		Last    Semester
		Current Semester
		Next    Semester
	}

	Period int
	Season int
	Status int

	PostgresNotify struct {
		Payload    string     `json:"topic_name,omitempty"`
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

var seasons = [...]string{
	"fall",
	"spring",
	"summer",
	"winter",
}

func (s Season) String() string {
	return seasons[s]
}

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
	return fmt.Sprintf("%x", sha1.Sum([]byte(s.Name+s.Number+s.Season+strconv.Itoa(s.Year)+buffer.String())))
}

func (c Course) hash() string {
	var buffer bytes.Buffer
	for _, section := range c.Sections {
		buffer.WriteString(section.CallNumber + section.Number)
	}
	return fmt.Sprintf("%x", sha1.Sum([]byte(c.Name+c.Number+buffer.String())))
}

func (u *University) VetAndBuild() {
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

func (sub *Subject) VetAndBuild() {
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

func (course *Course) VetAndBuild() {
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
	if  course.Synopsis != nil {
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

func (section *Section) VetAndBuild() {
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

	// Max
	if section.Max == 0 {
		section.Now = section.Max
	}

	// Status
	if section.Status == "" {
		log.Panic("Status == is empty")
	}

	// Credits
	if section.Credits == "" {
		log.Panic("Credits == is empty")
	}
}

func (meeting *Meeting) VetAndBuild() {

	if meeting.StartTime == "" {
		meeting.StartTime = "00:00 AM"
		meeting.EndTime = "00:00 AM"
	}

	/*meeting.StartTime.String = TrimAll(meeting.StartTime.String)
	meeting.EndTime.String = TrimAll(meeting.EndTime.String)*/
}

func (instructor *Instructor) VetAndBuild() {
	// Name
	if instructor.Name == "" {
		log.Panic("Instructor name == is empty")
	}
	instructor.Name = trim(instructor.Name)
}

func (book *Book) vetAndBuild() {
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

func (metaData *Metadata) vetAndBuild() {
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
	return r.PeriodDate.Month()
}

func (r Registration) day() int {
	return r.PeriodDate.Day()
}

func (r Registration) dayOfYear() int {
	return r.PeriodDate.YearDay()
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
		Year:   year,
		Season: FALL}

	winter := Semester{
		Year:   year,
		Season: WINTER}

	spring := Semester{
		Year:   year,
		Season: SPRING}

	summer := Semester{
		Year:   year,
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
