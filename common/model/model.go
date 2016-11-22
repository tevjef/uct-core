package model

import (
	log "github.com/Sirupsen/logrus"
	//	"golang.org/x/exp/utf8string"
	"hash/fnv"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
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
	InFall Period = iota
	InSpring
	InSummer
	InWinter
	StartFall
	StartSpring
	StartSummer
	StartWinter
	EndFall
	EndSpring
	EndSummer
	EndWinter
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
	Fall   = "fall"
	Spring = "spring"
	Summer = "summer"
	Winter = "winter"
)

const (
	Open Status = 1 + iota
	Closed
	Cancelled
)

var status = [...]string{
	"Open",
	"Closed",
	"Cancelled",
}

func (s Status) String() string {
	return status[s-1]
}

func isValidTopicRune(char rune) bool {
	isDash := char == rune('-')
	isUnderscore := char == rune('_')
	isDot := char == rune('.')
	isTilde := char == rune('~')
	isPercent := char == rune('%')
	return unicode.IsLetter(char) || unicode.IsNumber(char) || unicode.IsSpace(char) || isDash || isUnderscore || isDot || isTilde || isPercent
}

func ToTopicName(str string) string {
	str = strings.Map(func(r rune) rune {
		if isValidTopicRune(r) {
			return r
		}
		return -1
	}, str)

	// replaces spaces with dots
	var lastRune rune
	dot := rune('.')
	str = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == dot {
			if unicode.IsSpace(lastRune) || lastRune == dot {
				return -1
			} else {
				lastRune = dot
				return dot
			}
		}
		// else keep it in the string
		lastRune = r
		return unicode.ToLower(r)
	}, str)

	return str
}

var topicHash = fnv.New64a()

func ToTopicId(str string) string {
	topicHash.Reset()
	topicHash.Write([]byte(str))
	return strconv.FormatUint(topicHash.Sum64(), 10)
}

func ToTitle(str string) string {
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

	u.TopicName = ToTopicName(u.Name)
	u.TopicId = ToTopicId(u.TopicName)
}

func (sub *Subject) Validate(uni *University) {
	// Name
	if sub.Name == "" {
		log.Panic("Subject name == is empty")
	}

	sub.Name = TrimAll(sub.Name)
	sub.Name = ToTitle(sub.Name)

	// TopicName
	sub.TopicName = uni.TopicName + "." + sub.Number + "." + sub.Name + "." + sub.Season + "." + sub.Year
	sub.TopicName = ToTopicName(sub.TopicName)
	sub.TopicId = ToTopicId(sub.TopicName)

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
	course.Name = ToTitle(course.Name)

	// Number
	if course.Number == "" {
		log.Panic("Number == is empty")
	}

	// Synopsis
	if course.Synopsis != nil {
		temp := TrimAll(*course.Synopsis)
		//temp = utf8string.NewString(*course.Synopsis).String()
		course.Synopsis = &temp
	}

	// TopicName
	course.TopicName = course.Number + "." + course.Name
	course.TopicName = subject.TopicName + "." + course.TopicName
	course.TopicName = ToTopicName(course.TopicName)
	course.TopicId = ToTopicId(course.TopicName)
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
	} else if strings.ToLower(section.Status) != "open" && strings.ToLower(section.Status) != "closed" {
		log.Panicf("Status != open || status != closed status=%s", section.Status)
	}

	// Max
	if section.Max == 0 {
		section.Max = 1
		if section.Status == Open.String() {
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
	section.TopicName = ToTopicName(section.TopicName)
	section.TopicId = ToTopicId(section.TopicName)
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
	case InFall.String():
		return Fall
	case InSpring.String():
		return Spring
	case InSummer.String():
		return Summer
	case InWinter.String():
		return Winter
	default:
		return Summer
	}
}

func ResolveSemesters(t time.Time, registration []*Registration) *ResolvedSemester {
	month := t.Month()
	day := t.Day()
	year := t.Year()

	yearDay := t.YearDay()

	//var springReg = registration[SEM_SPRING];
	var winterReg = registration[InWinter]
	//var summerReg = registration[SEM_SUMMER];
	//var fallReg  = registration[SEM_FALL];
	var startFallReg = registration[StartFall]
	var startSpringReg = registration[StartSpring]
	var endSummerReg = registration[EndSummer]
	//var startSummerReg  = registration[START_SUMMER];

	fall := &Semester{
		Year:   int32(year),
		Season: Fall}

	winter := &Semester{
		Year:   int32(year),
		Season: Winter}

	spring := &Semester{
		Year:   int32(year),
		Season: Spring}

	summer := &Semester{
		Year:   int32(year),
		Season: Summer}

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
	return (c1.Number + c1.Name) < (c2.Number + c2.Number)
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

func DiffAndFilter(oldUni, newUni University) (filteredUniversity University) {
	filteredUniversity = newUni
	oldSubjects := oldUni.Subjects
	newSubjects := newUni.Subjects

	var filteredSubjects []*Subject
	// For each newer subject
	for s := range newSubjects {
		// If current index is out of range of the old subjects[] break and add every subject
		if s >= len(oldSubjects) {
			filteredSubjects = newSubjects
			break
		}

		if !newSubjects[s].Equal(oldSubjects[s]) {
			oldCourses := oldSubjects[s].Courses
			newCourses := newSubjects[s].Courses
			var filteredCourses []*Course
			for c := range newCourses {
				if c >= len(oldCourses) {
					filteredCourses = newCourses
					break
				}

				if !newCourses[c].Equal(oldCourses[c]) {
					oldSections := oldCourses[c].Sections
					newSections := newCourses[c].Sections
					oldSectionFields := logSection(oldSections, "old")
					newSectionFields := logSection(newSections, "new")
					var filteredSections []*Section
					for e := range newSections {
						if e >= len(oldSections) {
							filteredSections = newSections
							break
						}
						if !newSections[e].Equal(oldSections[e]) {
							fullSection := log.Fields{"old_full_section": oldSections[e].String(), "new_full_section": newSections[e].String()}
							log.WithFields(log.Fields{
								"old_call_number": oldSections[e].CallNumber, "old_status": oldSections[e].Status,
								"new_call_number": newSections[e].CallNumber, "new_status": newSections[e].Status,
							}).WithFields(oldSectionFields).WithFields(newSectionFields).WithFields(fullSection).WithFields(log.Fields{"old_section": oldSections[e].TopicName, "new_section": newSections[e].TopicName}).Info("diff")
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

func logSection(section []*Section, prepend string) log.Fields {
	openCount := 0
	closedCount := 0
	for i := range section {
		if section[i].Status == "Open" {
			openCount++
		} else if section[i].Status == "Closed" {
			closedCount++
		}
	}
	return log.Fields{
		prepend + "_open_count":   openCount,
		prepend + "_closed_count": closedCount,
	}
}

type SemesterSorter []*Semester

func (a SemesterSorter) Len() int {
	return len(a)
}

func (a SemesterSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a SemesterSorter) Less(i, j int) bool {
	if a[j].Year < a[i].Year {
		return true
	} else if a[i].Year == a[j].Year {
		return rankSeason(a[j].Season) < rankSeason(a[i].Season)
	}
	return false
}

func rankSeason(seasonStr string) int {
	seasonStr = strings.ToLower(seasonStr)
	switch seasonStr {
	case "fall":
		return 4
	case "summer":
		return 3
	case "spring":
		return 2
	case "winter":
		return 1
	}
	return 0
}
