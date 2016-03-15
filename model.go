package main

import (
	"database/sql"
	"net/url"
	"regexp"
	"strings"
	"time"
	"log"
)

type (
	University struct {
		Id               float64   `json:"id,omitempty" db:"id"`
		Name             string    `json:"name,omitempty" db:"name"`
		Abbr             string    `json:"abbr,omitempty" db:"abbr"`
		HomePage         string    `json:"home_page,omitempty" db:"home_page"`
		RegistrationPage string    `json:"registration_page,omitempty" db:"registration_page"`
		MainColor        string    `json:"main_color,omitempty" db:"main_color"`
		AccentColor      string    `json:"accent_color,omitempty" db:"accent_color"`
		TopicName        string    `json:"topic_name,omitempty" db:"topic_name"`
		CreatedAt        time.Time `json:"created_at,omitempty" db:"created_at"`
		UpdatedAt        time.Time `json:"updated_at,omitempty" db:"updated_at"`
		Subjects         []Subject `json:"subjects,omitempty"`
		Registration 	[]Registration `json:"registration,omitempty"`
		Metadata 	[]Metadata `json:"metadata,omitempty"`
	}

	Subject struct {
		Id           float64   `json:"id,omitempty" db:"id"`
		UniversityId float64   `json:"university_id,omitempty" db:"university_id"`
		Name         string    `json:"name,omitempty" db:"name"`
		Abbr         string    `json:"abbr,omitempty" db:"abbr"`
		Season Season    `json:"season,omitempty" db:"season"`
		Year   time.Time `json:"year,omitempty" db:"year"`
		TopicName    string    `json:"topic_name,omitempty" db:"topic_name"`
		CreatedAt    time.Time `json:"created_at,omitempty" db:"created_at"`
		UpdatedAt    time.Time `json:"updated_at,omitempty" db:"updated_at"`
		Courses      []Course  `json:"courses,omitempty"`
		Metadata 	[]Metadata `json:"metadata,omitempty"`
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
		Metadata 	[]Metadata `json:"metadata,omitempty"`
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
		Metadata 	[]Metadata `json:"metadata,omitempty"`
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
	}


	TimePeriod struct {
		Period     Period    `json:"period,omitempty" db:"period"`
		PeriodDate time.Time `json:"period_date,omitempty" db:"period_date"`
	}

	Period int
	Season int
	Status int
)

const (
	FALL Period = iota
	SPRING
	SUMMER
	WINTER
	PRE_FALL
	PRE_SPRING
	PRE_SUMMER
	PRE_WINTER
	DROP_FALL
	DROP_SPRING
	DROP_SUMMER
	DROP_WINTER
)

var seasons = [...]string{
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
	return seasons[s]
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
	if u.Name == nil {
		log.Panic("University name == nil")
	}
	u.Name = strings.Replace(u.Name, "_", " ", -1)
	u.Name = strings.Replace(u.Name, ".", " ", -1)
	u.Name = strings.Replace(u.Name, "~", " ", -1)
	u.Name = strings.Replace(u.Name, "%", " ", -1)

	// Abbr
	if u.Abbr == nil {
		regex, err := regexp.Compile("[^A-Z]")
		checkError(err)
		u.Abbr = trim(regex.ReplaceAllString(u.Name, ""))
	}

	// Homepage
	if u.HomePage == nil {
		log.Panic("HomePage == nil")
	}
	u.HomePage = trim(u.HomePage)
	url, err := url.ParseRequestURI(u.HomePage)
	checkError(err)
	u.HomePage = url.String()

	// RegistrationPage
	if u.RegistrationPage == nil {
		log.Panic("RegistrationPage == nil")
	}
	u.RegistrationPage = trim(u.RegistrationPage)
	url, err = url.ParseRequestURI(u.RegistrationPage)
	checkError(err)
	u.RegistrationPage = url.String()

	// MainColor
	if u.MainColor == nil {
		u.MainColor = "00000000"
	}

	// AccentColor
	if u.AccentColor == nil {
		u.AccentColor = "00000000"
	}

	// TopicName
	regex, err := regexp.Compile("\\s\\s+")
	checkError(err)
	u.TopicName = regex.ReplaceAllString(u.Name, ".")
	regex, err = regexp.Compile("[^A-Za-z.]")
	checkError(err)
	u.TopicName = trim(regex.ReplaceAllString(u.TopicName, ""))
}

func (sub *Subject) vetAndBuild() {
	// Name
	if sub.Name == nil {
		log.Panic("Subject name == nil")
	}
	sub.Name = strings.Replace(sub.Name, "_", " ", -1)
	sub.Name = strings.Replace(sub.Name, ".", " ", -1)
	sub.Name = strings.Replace(sub.Name, "~", " ", -1)
	sub.Name = strings.Replace(sub.Name, "%", " ", -1)

	// Abbr
	if sub.Abbr == nil {
		regex, err := regexp.Compile("[^A-Z]")
		checkError(err)
		sub.Abbr = trim(regex.ReplaceAllString(sub.Name, ""))
		if len(sub.Abbr) < 3 {
			sub.Abbr = sub.Abbr[:3]
		}
	}

	// Season
	if sub.Season == nil {
		log.Panic("Season == nil")
	}

	// Year
	if sub.Year == nil {
		log.Panic("Year == nil")
	}
	sub.Year = time.Date(sub.Year.Year(),1,1,0,0,0,0, time.UTC)

	// TopicName
	regex, err := regexp.Compile("\\s\\s+")
	checkError(err)
	sub.TopicName = regex.ReplaceAllString(sub.Name, ".")
	regex, err = regexp.Compile("[^A-Za-z.]")
	checkError(err)
	sub.TopicName = trim(regex.ReplaceAllString(sub.TopicName, ""))
}

func (course *Course) vetAndBuild() {
	// Name
	if course.Name == nil {
		log.Panic("Subject name == nil")
	}
	course.Name = strings.Replace(course.Name, "_", " ", -1)
	course.Name = strings.Replace(course.Name, ".", " ", -1)
	course.Name = strings.Replace(course.Name, "~", " ", -1)
	course.Name = strings.Replace(course.Name, "%", " ", -1)
	course.Name = trim(course.Name)

	// Number
	if course.Number == nil {
		log.Panic("Number == nil")
	}

	// Synopsis
	if course.Synopsis != nil {
		regex, err := regexp.Compile("\\s\\s+")
		checkError(err)
		course.Synopsis = regex.ReplaceAllString(course.Synopsis, " ")
	}

	// TopicName
	regex, err := regexp.Compile("\\s\\s+")
	checkError(err)
	course.TopicName = regex.ReplaceAllString(course.Name, ".")
	regex, err = regexp.Compile("[^A-Za-z.]")
	checkError(err)
	course.TopicName = trim(regex.ReplaceAllString(course.TopicName, ""))
}

func (section *Section) vetAndBuild() {
	// Number
	if section.Number == nil {
		log.Panic("Number == nil")
	}
	section.Number = trim(section.Number)

	// Call Number
	if section.CallNumber == nil {
		log.Panic("CallNumber == nil")
	}
	section.CallNumber = trim(section.CallNumber)

	// Max
	if section.Max == nil {
		section.Max = 0
	}

	// Now
	if section.Now == nil {
		section.Now = section.Max
	}

	// Status
	if section.Status == nil {
		log.Panic("Status == nil")
	}

	// Credits
	if section.Credits == nil {
		log.Panic("Credits == nil")
	}
}

func (meeting *Meeting) vetAndBuild() {
	// Number
	if meeting.Room == nil {
		log.Panic("Room == nil")
	}

	// StartTime
	if meeting.StartTime == nil {
		log.Panic("StartTime == nil")
	}

	// EndTime
	if meeting.EndTime == nil {
		log.Panic("EndTime == nil")
	}
}

func (instructor *Instructor) vetAndBuild() {
	// Name
	if instructor.Name == nil {
		log.Panic("Instructor name == nil")
	}
	instructor.Name = trim(instructor.Name)
}

func (book *Book) vetAndBuild() {
	// Name
	if book.Title == nil {
		log.Panic("Instructor name == nil")
	}
	book.Title = trim(book.Title)

	// RegistrationPage
	if book.Url == nil {
		log.Panic("RegistrationPage == nil")
	}
	book.Url = trim(book.Url)
	url, err := url.ParseRequestURI(book.Url)
	checkError(err)
	book.Url = url.String()
}

func (metaData *Metadata) vetAndBuild() {
	// Title
	if metaData.Title == nil {
		log.Panic("Title == nil")
	}
	metaData.Title = trim(metaData.Title)

	// Content
	if metaData.Content == nil {
		log.Panic("Content == nil")
	}
	metaData.Content = trim(metaData.Content)
}