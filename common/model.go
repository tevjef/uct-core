package common

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"golang.org/x/exp/utf8string"
	"log"
	"net/url"
	"regexp"
	"sort"
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
	FALL   = "fall"
	SPRING = "spring"
	SUMMER = "summer"
	WINTER = "winter"
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

func (s Subject) hash(uni University) string {
	var buffer bytes.Buffer
	buffer.WriteString(uni.Name)
	buffer.WriteString(s.Name)
	buffer.WriteString(s.Number)
	buffer.WriteString(s.Season)
	buffer.WriteString(s.Year)

	h := fmt.Sprintf("%x", sha1.Sum([]byte(buffer.String())))
	/*	if len(h) >= 64 {
		return h[:64]
	}*/
	return h
}

func (c Course) hash(uni University, sub Subject) string {
	var buffer bytes.Buffer
	for _, section := range c.Sections {
		buffer.WriteString(uni.Name)
		buffer.WriteString(sub.Name)
		buffer.WriteString(section.CallNumber)
		buffer.WriteString(section.Number)
	}
	buffer.WriteString(c.Name)
	buffer.WriteString(c.Number)
	h := fmt.Sprintf("%x", sha1.Sum([]byte(buffer.String())))
	/*	if len(h) >= 64 {
		return h[:64]
	}*/
	return h
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
	regex, err = regexp.Compile("[^A-Za-z0-9]+")
	CheckError(err)
	u.TopicName = trim(regex.ReplaceAllString(u.TopicName, "."))
	limit := 25
	if len(u.TopicName) >= limit {
		u.TopicName = u.TopicName[:limit]
	}
}

func (sub *Subject) Validate(uni *University) {
	// Name
	if sub.Name == "" {
		log.Panic("Subject name == is empty")
	}
	sub.Name = strings.Title(strings.ToLower(sub.Name))
	sub.Name = strings.Replace(sub.Name, "_", " ", -1)
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
	sub.TopicName = sub.Number + " " + sub.Name + " " + sub.Season + " " + sub.Year
	sub.TopicName = regex.ReplaceAllString(sub.TopicName, ".")
	regex, err = regexp.Compile("[.]+")
	CheckError(err)
	sub.TopicName = trim(regex.ReplaceAllString(sub.TopicName, "."))

	sub.TopicName = uni.TopicName + "__" + sub.TopicName
	sort.Sort(courseSorter{sub.Courses})
}

func (course *Course) Validate(subject *Subject) {
	// Name
	if course.Name == "" {
		log.Panic("Course name == is empty", course)
	}

	course.Name = strings.Title(strings.ToLower(course.Name))
	course.Name = strings.Replace(course.Name, "_", " ", -1)
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
	course.TopicName = course.Number + " " + course.Name
	course.TopicName = regex.ReplaceAllString(course.TopicName, ".")
	regex, err = regexp.Compile("[.]+")
	CheckError(err)
	course.TopicName = trim(regex.ReplaceAllString(course.TopicName, "."))

	course.TopicName = subject.TopicName + "__" + course.TopicName
	sort.Stable(sectionSorter{course.Sections})
}

// Validate within the context for these enclosing objects
func (section *Section) Validate(course *Course) {
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

	section.TopicName = section.Number + "_" + section.CallNumber
	section.TopicName = course.TopicName + "__" + section.TopicName
	//sort.Stable(meetingSorter{section.Meetings})
	sort.Stable(instructorSorter{section.Instructors})

}

func (meeting *Meeting) Validate() {
	if meeting.StartTime != nil {
		a := TrimAll(*meeting.StartTime)
		if a == "" {
			meeting.StartTime = nil
		} else {
			meeting.StartTime = &a
		}
	}

	if meeting.EndTime != nil {
		b := TrimAll(*meeting.EndTime)
		if b == "" {
			meeting.EndTime = nil
		} else {
			meeting.EndTime = &b
		}
	}
}

func (instructor *Instructor) Validate() {
	// Name
	if instructor.Name == "" {
		log.Panic("Instructor name == is empty")
	}
	if instructor.Name[len(instructor.Name)-1:] == "-" {
		instructor.Name = instructor.Name[:len(instructor.Name)-1]
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
	return time.Unix(r.PeriodDate, 0).UTC().Month()
}

func (r Registration) day() int {
	return time.Unix(r.PeriodDate, 0).UTC().Day()
}

func (r Registration) dayOfYear() int {
	return time.Unix(r.PeriodDate, 0).UTC().YearDay()
}

func (r Registration) season() string {
	switch r.Period {
	case SEM_FALL.String():
		return FALL
	case SEM_SPRING.String():
		return SPRING
	case SEM_SUMMER.String():
		return SUMMER
	case SEM_WINTER.String():
		return WINTER
	default:
		return SUMMER
	}
}

func ResolveSemesters(t time.Time, registration []*Registration) ResolvedSemester {
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
		//Log("Spring: Winter - StartFall ", winterReg.month(), winterReg.day(), "--", startFallReg.month(), startFallReg.day(), "--", month, day)
		return ResolvedSemester{
			Last:    winter,
			Current: spring,
			Next:    summer}

	} else if yearDay >= startFallReg.dayOfYear() && yearDay < endSummerReg.dayOfYear() {
		//Log("StartFall: StartFall -- EndSummer ", startFallReg.dayOfYear(), "--", endSummerReg.dayOfYear(), "--", yearDay)
		return ResolvedSemester{
			Last:    spring,
			Current: summer,
			Next:    fall,
		}
	} else if yearDay >= endSummerReg.dayOfYear() && yearDay < startSpringReg.dayOfYear() {
		//Log("Fall: EndSummer -- StartSpring ", endSummerReg.dayOfYear(), "--", yearDay < startSpringReg.dayOfYear(), "--", yearDay)
		return ResolvedSemester{
			Last:    summer,
			Current: fall,
			Next:    winter,
		}
	} else if yearDay >= startSpringReg.dayOfYear() && yearDay < winterReg.dayOfYear() {
		spring.Year = spring.Year + 1
		//Log("StartSpring: StartSpring -- Winter ", startSpringReg.dayOfYear(), "--", winterReg.dayOfYear(), "--", yearDay)
		return ResolvedSemester{
			Last:    fall,
			Current: winter,
			Next:    spring,
		}
	}

	return ResolvedSemester{}
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
	return (c1.Number + c1.GoString()) < (c2.Number + c1.GoString())
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

func DiffAndFilter(uni, uni2 University) (filteredUniversity University) {
	filteredUniversity = uni2
	oldSubjects := uni.Subjects
	newSubjects := uni2.Subjects
	var filteredSubjects []*Subject
	// For each newer subject

	for s := range newSubjects {
		// If current index is out of range of the old subjects[] break and add every subject
		if s >= len(oldSubjects) {
			filteredSubjects = newSubjects
			break
		}
		Log(fmt.Sprintf("%s Subject index: %d \t %s | %s %s\n", oldSubjects[s].Season, s, oldSubjects[s].Name, newSubjects[s].Name, oldSubjects[s].Number == newSubjects[s].Number))
		// If newSubject != oldSubject, log why, then drill deeper to filter out ones tht haven't changed
		if err := newSubjects[s].VerboseEqual(oldSubjects[s]); err != nil {
			Log("Subject differs!")
			oldCourses := oldSubjects[s].Courses
			newCourses := newSubjects[s].Courses
			var filteredCourses []*Course
			for c := range newCourses {
				if c >= len(oldCourses) {
					filteredCourses = newCourses
					break
				}
				Log(fmt.Sprintf("Course index: %d \t %s | %s %s\n", c, oldCourses[c].Name, newCourses[c].Name, oldCourses[c].Number == newCourses[c].Number))
				if oldCourses[c].Number != newCourses[c].Number {
					fmt.Printf("%s %s", oldCourses[c].Name, newCourses[c].Name)
				}
				if err := newCourses[c].VerboseEqual(oldCourses[c]); err != nil {
					Log("Course differs!")
					oldSections := oldCourses[c].Sections
					newSections := newCourses[c].Sections
					var filteredSections []*Section
					for e := range newSections {
						//Log(fmt.Sprintf("Section index: %d \t %s | %s %s\n", e, oldSections[e].Number, newSections[e].Number, oldSections[e].Number == newSections[e].Number))
						if e >= len(oldSections) {
							filteredSections = newSections
							break
						}
						if err := newSections[e].VerboseEqual(oldSections[e]); err != nil {
							Log("Section: ", newSections[e].CallNumber, " | ", oldSections[e].CallNumber, err)
							filteredSections = append(filteredSections, newSections[e])
						}
					}
					newCourses[c].Sections = filteredSections
					filteredCourses = append(filteredCourses, newCourses[c])
				}
			}
			newSubjects[s].Courses = filteredCourses
			filteredSubjects = append(filteredSubjects, newSubjects[s])
		}
	}
	filteredUniversity.Subjects = filteredSubjects
	return
}
