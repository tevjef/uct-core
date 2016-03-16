package common

import (
	"database/sql"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type (
	University struct {
		Id               float64        `json:"id,omitempty" db:"id"`
		Name             string         `json:"name,omitempty" db:"name"`
		Abbr             string         `json:"abbr,omitempty" db:"abbr"`
		HomePage         string         `json:"home_page,omitempty" db:"home_page"`
		RegistrationPage string         `json:"registration_page,omitempty" db:"registration_page"`
		MainColor        string         `json:"main_color,omitempty" db:"main_color"`
		AccentColor      string         `json:"accent_color,omitempty" db:"accent_color"`
		TopicName        string         `json:"topic_name,omitempty" db:"topic_name"`
		CreatedAt        time.Time      `json:"created_at,omitempty" db:"created_at"`
		UpdatedAt        time.Time      `json:"updated_at,omitempty" db:"updated_at"`
		Subjects         []Subject      `json:"subjects,omitempty"`
		Registrations    []Registration `json:"registration,omitempty"`
		Metadata         []Metadata     `json:"metadata,omitempty"`
	}

	Subject struct {
		Id           float64    `json:"id,omitempty" db:"id"`
		UniversityId float64    `json:"university_id,omitempty" db:"university_id"`
		Name         string     `json:"name,omitempty" db:"name"`
		Abbr         string     `json:"abbr,omitempty" db:"abbr"`
		Season       Season     `json:"season,omitempty" db:"season"`
		Year         time.Time  `json:"year,omitempty" db:"year"`
		TopicName    string     `json:"topic_name,omitempty" db:"topic_name"`
		CreatedAt    time.Time  `json:"created_at,omitempty" db:"created_at"`
		UpdatedAt    time.Time  `json:"updated_at,omitempty" db:"updated_at"`
		Courses      []Course   `json:"courses,omitempty"`
		Metadata     []Metadata `json:"metadata,omitempty"`
	}

	Course struct {
		Id        float64        `json:"id,omitempty" db:"id"`
		SubjectId float64        `json:"subject_id,omitempty" db:"subject_id"`
		Name      string         `json:"name,omitempty" db:"name"`
		Number    string         `json:"number,omitempty" db:"number"`
		Synopsis  sql.NullString `json:"synopsis,omitempty" db:"synopsis"`
		TopicName string         `json:"topic_name,omitempty" db:"topic_name"`
		CreatedAt time.Time      `json:"created_at,omitempty" db:"created_at"`
		UpdatedAt time.Time      `json:"updated_at,omitempty" db:"updated_at"`
		Sections  []Section      `json:"sections,omitempty"`
		Metadata  []Metadata     `json:"metadata,omitempty"`
	}

	Section struct {
		Id          float64      `json:"id,omitempty" db:"id"`
		CourseId    float64      `json:"course_id,omitempty" db:"course_id"`
		Number      string       `json:"number,omitempty" db:"number"`
		CallNumber  string       `json:"call_number,omitempty" db:"call_number"`
		Max         float64      `json:"max,omitempty" db:"max"`
		Now         float64      `json:"now,omitempty" db:"now"`
		Status      string       `json:"status,omitempty" db:"status"`
		Credits     string       `json:"credits,omitempty" db:"credits"`
		TopicName   string       `json:"topic_name,omitempty" db:"topic_name"`
		CreatedAt   time.Time    `json:"created_at,omitempty" db:"created_at"`
		UpdatedAt   time.Time    `json:"updated_at,omitempty" db:"updated_at"`
		Meetings    []Meeting    `json:"meeting,omitempty"`
		Instructors []Instructor `json:"instructors,omitempty"`
		Books       []Book       `json:"books,omitempty"`
		Metadata    []Metadata   `json:"metadata,omitempty"`
	}

	Meeting struct {
		Id        float64   `json:"id,omitempty" db:"id"`
		SectionId float64   `json:"section_id,omitempty" db:"section_id"`
		Room      string    `json:"room,omitempty" db:"room"`
		StartTime string    `json:"start_time,omitempty" db:"start_time"`
		EndTime   string    `json:"end_time,omitempty" db:"section_id"`
		CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
		UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
		Meetings  []Meeting `json:"meeting,omitempty"`
	}

	Instructor struct {
		Id        float64   `json:"id,omitempty" db:"id"`
		SectionId float64   `json:"section_id,omitempty" db:"section_id"`
		Name      string    `json:"name,omitempty" db:"name"`
		CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
		UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
	}

	Book struct {
		Id        float64   `json:"id,omitempty" db:"id"`
		SectionId float64   `json:"section_id,omitempty" db:"section_id"`
		Title     string    `json:"title,omitempty" db:"title"`
		Url       string    `json:"url,omitempty" db:"url"`
		CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
		UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
	}

	Metadata struct {
		Id           float64   `json:"id,omitempty" db:"id"`
		UniversityId float64   `json:"university_id,omitempty" db:"university_id"`
		SubjectId    float64   `json:"subject_id,omitempty" db:"subject_id"`
		CourseId     float64   `json:"course_id,omitempty" db:"course_id"`
		SectionId    float64   `json:"section_id,omitempty" db:"section_id"`
		Title        string    `json:"title,omitempty" db:"title"`
		Content      string    `json:"content,omitempty" db:"content"`
		CreatedAt    time.Time `json:"created_at,omitempty" db:"created_at"`
		UpdatedAt    time.Time `json:"updated_at,omitempty" db:"updated_at"`
	}

	Registration struct {
		Id           float64   `json:"id,omitempty" db:"id"`
		UniversityId float64   `json:"university_id,omitempty" db:"university_id"`
		Period       Period    `json:"period,omitempty" db:"period"`
		PeriodDate   time.Time `json:"period_date,omitempty" db:"period_date"`
		CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
		UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
	}

	TimePeriod struct {
		Period     Period    `json:"period,omitempty" db:"period"`
		PeriodDate time.Time `json:"period_date,omitempty" db:"period_date"`
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
	OPEN Status = iota
	CLOSED
	CANCELLED
)

var status = [...]string{
	"Open",
	"Closed",
	"Cancelled",
}

func (s Status) String() string {
	return status[s]
}

func (u *University) vetAndBuild() {
	// Name
	if u.Name == ""  {
		log.Panic("University name == is empty")
	}
	u.Name = strings.Replace(u.Name, "_", " ", -1)
	u.Name = strings.Replace(u.Name, ".", " ", -1)
	u.Name = strings.Replace(u.Name, "~", " ", -1)
	u.Name = strings.Replace(u.Name, "%", " ", -1)

	// Abbr
	if u.Abbr == ""  {
		regex, err := regexp.Compile("[^A-Z]")
		CheckError(err)
		u.Abbr = trim(regex.ReplaceAllString(u.Name, ""))
	}

	// Homepage
	if u.HomePage == ""  {
		log.Panic("HomePage == is empty")
	}
	u.HomePage = trim(u.HomePage)
	nUrl, err := url.ParseRequestURI(u.HomePage)
	CheckError(err)
	u.HomePage = nUrl.String()

	// RegistrationPage
	if u.RegistrationPage == ""  {
		log.Panic("RegistrationPage == is empty")
	}
	u.RegistrationPage = trim(u.RegistrationPage)
	nUrl, err = url.ParseRequestURI(u.RegistrationPage)
	CheckError(err)
	u.RegistrationPage = nUrl.String()

	// MainColor
	if u.MainColor == ""  {
		u.MainColor = "00000000"
	}

	// AccentColor
	if u.AccentColor == ""  {
		u.AccentColor = "00000000"
	}

	// Registration
	if len(u.Registrations) != 12  {
		log.Panic("Registration != 12 ")
	}

	// TopicName
	regex, err := regexp.Compile("\\s\\s+")
	CheckError(err)
	u.TopicName = regex.ReplaceAllString(u.Name, ".")
	regex, err = regexp.Compile("[^A-Za-z.]")
	CheckError(err)
	u.TopicName = trim(regex.ReplaceAllString(u.TopicName, ""))
	u.TopicName = u.TopicName[:26]
}

func (sub *Subject) vetAndBuild() {
	// Name
	if sub.Name == ""  {
		log.Panic("Subject name == is empty")
	}
	sub.Name = strings.Replace(sub.Name, "_", " ", -1)
	sub.Name = strings.Replace(sub.Name, ".", " ", -1)
	sub.Name = strings.Replace(sub.Name, "~", " ", -1)
	sub.Name = strings.Replace(sub.Name, "%", " ", -1)

	// Abbr
	if sub.Abbr == ""  {
		regex, err := regexp.Compile("[^A-Z]")
		CheckError(err)
		sub.Abbr = trim(regex.ReplaceAllString(sub.Name, ""))
		if len(sub.Abbr) < 3 {
			sub.Abbr = sub.Abbr[:3]
		}
	}

	sub.Year = time.Date(sub.Year.Year(), 1, 1, 0, 0, 0, 0, time.UTC)

	// TopicName
	regex, err := regexp.Compile("\\s\\s+")
	CheckError(err)
	sub.TopicName = regex.ReplaceAllString(sub.Name, ".")
	regex, err = regexp.Compile("[^A-Za-z.]")
	CheckError(err)
	sub.TopicName = trim(regex.ReplaceAllString(sub.TopicName, ""))
}

func (course *Course) vetAndBuild() {
	// Name
	if course.Name == ""  {
		log.Panic("Subject name == is empty")
	}
	course.Name = strings.Replace(course.Name, "_", " ", -1)
	course.Name = strings.Replace(course.Name, ".", " ", -1)
	course.Name = strings.Replace(course.Name, "~", " ", -1)
	course.Name = strings.Replace(course.Name, "%", " ", -1)
	course.Name = trim(course.Name)

	// Number
	if course.Number == ""  {
		log.Panic("Number == is empty")
	}

	// Synopsis
	if course.Synopsis.String != "" {
		regex, err := regexp.Compile("\\s\\s+")
		CheckError(err)
		course.Synopsis.String = regex.ReplaceAllString(course.Synopsis.String, " ")
	}

	// TopicName
	regex, err := regexp.Compile("\\s\\s+")
	CheckError(err)
	course.TopicName = regex.ReplaceAllString(course.Name, ".")
	regex, err = regexp.Compile("[^A-Za-z.]")
	CheckError(err)
	course.TopicName = trim(regex.ReplaceAllString(course.TopicName, ""))
}

func (section *Section) vetAndBuild() {
	// Number
	if section.Number == ""  {
		log.Panic("Number == is empty")
	}
	section.Number = trim(section.Number)

	// Call Number
	if section.CallNumber == ""  {
		log.Panic("CallNumber == is empty")
	}
	section.CallNumber = trim(section.CallNumber)

	// Max
	if section.Max == 0  {
		section.Now = section.Max
	}

	// Status
	if section.Status == ""  {
		log.Panic("Status == is empty")
	}

	// Credits
	if section.Credits == ""  {
		log.Panic("Credits == is empty")
	}
}

func (meeting *Meeting) vetAndBuild() {
	// Number
	if meeting.Room == ""  {
		log.Panic("Room == is empty")
	}

	// StartTime
	if meeting.StartTime == ""  {
		log.Panic("StartTime == is empty")
	}

	// EndTime
	if meeting.EndTime == ""  {
		log.Panic("EndTime == is empty")
	}
}

func (instructor *Instructor) vetAndBuild() {
	// Name
	if instructor.Name == ""  {
		log.Panic("Instructor name == is empty")
	}
	instructor.Name = trim(instructor.Name)
}

func (book *Book) vetAndBuild() {
	// Title
	if book.Title == ""  {
		log.Panic("Title  == is empty")
	}
	book.Title = trim(book.Title)

	// Url
	if book.Url == ""  {
		log.Panic("Url == is empty")
	}
	book.Url = trim(book.Url)
	url, err := url.ParseRequestURI(book.Url)
	CheckError(err)
	book.Url = url.String()
}

func (metaData *Metadata) vetAndBuild() {
	// Title
	if metaData.Title == ""  {
		log.Panic("Title == is empty")
	}
	metaData.Title = trim(metaData.Title)

	// Content
	if metaData.Content == ""  {
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

func (r Registration) season() Season {
	switch r.Period {
	case SEM_FALL:
		return FALL
	case SEM_SPRING:
		return SPRING
	case SEM_SUMMER:
		return SUMMER
	case SEM_WINTER:
		return WINTER
	default:
		return SUMMER
	}
}

func ResolveSemesters(t time.Time, registration []Registration) ResolvedSemester {
	month := t.Month()
	day := t.Day()
	year := t.Year()

	yearDay := t.YearDay()

	//var springReg = registration[SEM_SPRING];
	var winterReg = registration[SEM_WINTER];
	//var summerReg = registration[SEM_SUMMER];
	//var fallReg  = registration[SEM_FALL];
	var startFallReg  = registration[START_FALL];
	var startSpringReg  = registration[START_SPRING];
	var endSummerReg  = registration[END_SUMMER];
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
		if winterReg.month() - month <= 0 {
			spring.Year = spring.Year + 1
			summer.Year = summer.Year + 1
		} else {
			winter.Year = winter.Year - 1
			fall.Year = fall.Year - 1
		}
		Log("Spring: Winter - StartFall ", winterReg.month(), winterReg.day(), "--", startFallReg.month(), startFallReg.day(),"--", month, day)

		return ResolvedSemester{
			Last: winter,
			Current: spring,
			Next: summer}

	} else if (yearDay >= startFallReg.dayOfYear() && yearDay < endSummerReg.dayOfYear()) {
		Log("StartFall: StartFall -- EndSummer ", startFallReg.dayOfYear(), "--",endSummerReg.dayOfYear(), "--", yearDay)
		return ResolvedSemester{
			Last:    spring,
			Current: summer,
			Next:    fall,
		}
	} else if (yearDay >= endSummerReg.dayOfYear() &&  yearDay < startSpringReg.dayOfYear()) {

		Log("Fall: EndSummer -- StartSpring ",  endSummerReg.dayOfYear(), "--", yearDay < startSpringReg.dayOfYear(), "--", yearDay)

		return ResolvedSemester{
			Last:    summer,
			Current: fall,
			Next:    winter,
		}
	} else if (yearDay >= startSpringReg.dayOfYear() && yearDay < winterReg.dayOfYear()) {
		spring.Year = spring.Year + 1
		Log("StartSpring: StartSpring -- Winter ", startSpringReg.dayOfYear(), "--", winterReg.dayOfYear(), "--", yearDay)

		return ResolvedSemester{
			Last:    fall,
			Current: winter,
			Next:    spring,
		}
	}

	return ResolvedSemester{

	}
}

