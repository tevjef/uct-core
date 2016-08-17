package common

import (
	log "github.com/Sirupsen/logrus"
	"golang.org/x/exp/utf8string"
	"hash/fnv"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type (
	MeetingByDay    []Meeting
	SectionByNumber []Section
	CourseByName    []Course
	SubjectByName   []Subject

	Period int
	Status int

	DBResolvedSemester struct {
		Id            int64  `db:"id"`
		UniversityId  int64  `db:"university_id"`
		CurrentSeason string `db:"current_season"`
		CurrentYear   string `db:"current_year"`
		LastSeason    string `db:"last_season"`
		LastYear      string `db:"last_year"`
		NextSeason    string `db:"next_season"`
		NextYear      string `db:"next_year"`
	}
)

const (
	PROTOBUF = "protobuf"
	JSON     = "json"
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

func toTopicName(str string) string {
	regex, err := regexp.Compile("[^A-Za-z0-9-_.~% ]+")
	CheckError(err)
	str = trim(regex.ReplaceAllString(str, ""))

	regex, err = regexp.Compile("\\s+")
	CheckError(err)
	str = regex.ReplaceAllString(str, ".")

	regex, err = regexp.Compile("[.]+")
	CheckError(err)
	str = regex.ReplaceAllString(str, ".")

	str = strings.ToLower(str)

	return str
}

func toTopicId(str string) string {
	h := fnv.New64a()
	h.Write([]byte(str))
	return strconv.FormatUint(h.Sum64(), 10)
}

func toTitle(str string) string {
	str = strings.Title(strings.ToLower(str))

	for i := len(str) - 1; i != 0; i-- {
		if strings.LastIndex(str, "i") == i {
			str = swapChar(str, "I", i)
		} else {
			break
		}
	}
	return str
}

func swapChar(s, char string, index int) string {
	left := s[:index]
	right := s[index+1:]
	return left + char + right
}

func (u *University) Validate() {
	// Name
	if u.Name == "" {
		log.Panic("University name == is empty")
	}

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

	if u.ResolvedSemesters == nil {
		log.Panic("ResolvedSemesters is nil")
	}

	if u.ResolvedSemesters.Current == nil {
		log.Panic("ResolvedSemesters.Current is nil")
	}

	u.TopicName = toTopicName(u.Name)
	u.TopicId = toTopicId(u.TopicName)
}

func (sub *Subject) Validate(uni *University) {
	// Name
	if sub.Name == "" {
		log.Panic("Subject name == is empty")
	}
	sub.Name = toTitle(sub.Name)

	// TopicName
	sub.TopicName = uni.TopicName + "." + sub.Number + "." + sub.Name + "." + sub.Season + "." + sub.Year
	sub.TopicName = toTopicName(sub.TopicName)
	sub.TopicId = toTopicId(sub.TopicName)

	if len(sub.Courses) == 0 {
		log.WithField("subject", sub.TopicName).Errorln("No course in subject")
	} else {
		sort.Sort(courseSorter{sub.Courses})
	}
}

func (course *Course) Validate(subject *Subject) {
	// Name
	if course.Name == "" {
		log.Panic("Course name == is empty", course)
	}

	course.Name = TrimAll(course.Name)
	course.Name = toTitle(course.Name)

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
	course.TopicName = course.Number + "." + course.Name
	course.TopicName = subject.TopicName + "." + course.TopicName
	course.TopicName = toTopicName(course.TopicName)
	course.TopicId = toTopicId(course.TopicName)
	if len(course.Sections) == 0 {
		log.WithField("course", course.TopicName).Errorln("No section in course")
	} else {
		sort.Stable(sectionSorter{course.Sections})
	}
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

	section.TopicName = course.TopicName + "." + section.Number + "." + section.CallNumber
	section.TopicName = toTopicName(section.TopicName)
	section.TopicId = toTopicId(section.TopicName)
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

func ResolveSemesters(t time.Time, registration []*Registration) *ResolvedSemester {
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

	fall := &Semester{
		Year:   int32(year),
		Season: FALL}

	winter := &Semester{
		Year:   int32(year),
		Season: WINTER}

	spring := &Semester{
		Year:   int32(year),
		Season: SPRING}

	summer := &Semester{
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
		return &ResolvedSemester{
			Last:    winter,
			Current: spring,
			Next:    summer}

	} else if yearDay >= startFallReg.dayOfYear() && yearDay < endSummerReg.dayOfYear() {
		Log("StartFall: StartFall -- EndSummer ", startFallReg.dayOfYear(), "--", endSummerReg.dayOfYear(), "--", yearDay)
		return &ResolvedSemester{
			Last:    spring,
			Current: summer,
			Next:    fall,
		}
	} else if yearDay >= endSummerReg.dayOfYear() && yearDay < startSpringReg.dayOfYear() {
		Log("Fall: EndSummer -- StartSpring ", endSummerReg.dayOfYear(), "--", yearDay < startSpringReg.dayOfYear(), "--", yearDay)
		return &ResolvedSemester{
			Last:    summer,
			Current: fall,
			Next:    winter,
		}
	} else if yearDay >= startSpringReg.dayOfYear() && yearDay < winterReg.dayOfYear() {
		spring.Year = spring.Year + 1
		Log("StartSpring: StartSpring -- Winter ", startSpringReg.dayOfYear(), "--", winterReg.dayOfYear(), "--", yearDay)
		return &ResolvedSemester{
			Last:    fall,
			Current: winter,
			Next:    spring,
		}
	}

	return &ResolvedSemester{}
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
		log.WithFields(log.Fields{
			"index":       s,
			"old_subject": oldSubjects[s].Name,
			"old_season":  oldSubjects[s].Season,
			"new_subject": newSubjects[s].Name,
			"new_season":  newSubjects[s].Season,
			"same":        oldSubjects[s].Number == newSubjects[s].Number,
		}).Debug()
		if err := newSubjects[s].VerboseEqual(oldSubjects[s]); err != nil {
			log.Errorln("Subject differs")
			oldCourses := oldSubjects[s].Courses
			newCourses := newSubjects[s].Courses
			var filteredCourses []*Course
			for c := range newCourses {
				if c >= len(oldCourses) {
					filteredCourses = newCourses
					break
				}
				log.WithFields(log.Fields{
					"index":      c,
					"old_course": oldCourses[c].Name,
					"old_number": oldCourses[c].Number,
					"new_course": newCourses[c].Name,
					"new_number": oldCourses[c].Number,
					"same":       oldCourses[c].Number == newCourses[c].Number,
				}).Debug()
				if err := newCourses[c].VerboseEqual(oldCourses[c]); err != nil {
					log.Errorln("Course differs")
					oldSections := oldCourses[c].Sections
					newSections := newCourses[c].Sections
					var filteredSections []*Section
					for e := range newSections {
						if e >= len(oldSections) {
							filteredSections = newSections
							break
						}
						if err := newSections[e].VerboseEqual(oldSections[e]); err != nil {
							log.WithFields(log.Fields{
								"old_call_number": oldSections[e].CallNumber,
								"new_call_number": newSections[e].CallNumber,
							}).Errorln()
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
