package model

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"unicode"
	"bytes"
	"hash/fnv"
)

var trim = strings.TrimSpace

func swapChar(s, char string, index int) string {
	left := s[:index]
	right := s[index+1:]
	return left + char + right
}


func stripSpaces(s string) string {
	var lastRune rune
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) && unicode.IsSpace(lastRune) {
			// if the character is a space, drop it
			return -1
		}
		lastRune = r
		// else keep it in the string
		return r
	}, s)
}

const NUL = []byte("\x00")
const SOH = []byte("\x01")

func TrimAll(str string) string {
	str = stripSpaces(str)
	temp := []byte(str)

	// Remove NUL and Heading bytes from string, cannot be inserted into postgresql
	temp = bytes.Replace(temp, NUL, []byte{}, -1)
	str = string(bytes.Replace(temp, SOH, []byte{}, -1))
	return trim(str)
}

func isValidTopicRune(r rune) bool {
	isLetter := unicode.IsLetter(r)
	isNumber := unicode.IsNumber(r)
	isSpace := unicode.IsSpace(r)
	isDash := r == rune('-')
	isUnderscore := r == rune('_')
	isDot := r == rune('.')
	isTilde := r == rune('~')
	isPercent := r == rune('%')
	return  isLetter || isNumber || isSpace || isDash || isUnderscore || isDot || isTilde || isPercent
}

func ToTopicName(topic string) string {
	// replaces spaces with dots
	var lastRune rune
	dot := rune('.')

	return strings.Map(func(r rune) rune {
		if !isValidTopicRune(r) {
			return -1
		}

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
	}, topic)
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

func ValidateAll(uni *University) (err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in ValidateAll", r)
			err = fmt.Errorf("%v+", r)
		}
	}()

	uni.Validate()
	CheckUniqueSubject(uni.Subjects)
	for subjectIndex := range uni.Subjects {
		subject := uni.Subjects[subjectIndex]
		subject.Validate(uni)

		courses := subject.Courses
		CheckUniqueCourse(subject, courses)
		for courseIndex := range courses {
			course := courses[courseIndex]
			course.Validate(subject)

			sections := course.Sections
			for sectionIndex := range sections {
				section := sections[sectionIndex]
				section.Validate(course)

				//[]Instructors
				instructors := section.Instructors
				for instructorIndex := range instructors {
					instructor := instructors[instructorIndex]
					instructor.Index = int32(instructorIndex)
					instructor.Validate()
				}

				//[]Meeting
				meetings := section.Meetings
				for meetingIndex := range meetings {
					meeting := meetings[meetingIndex]
					meeting.Index = int32(meetingIndex)
					meeting.Validate()

					// Meeting []Metadata
					metadatas := meeting.Metadata
					for metadataIndex := range metadatas {
						metadata := metadatas[metadataIndex]
						metadata.Validate()
					}
				}

				//[]Books
				books := section.Books
				for bookIndex := range books {
					book := books[bookIndex]
					book.Validate()
				}

				// Section []Metadata
				metadatas := section.Metadata
				for metadataIndex := range metadatas {
					metadata := metadatas[metadataIndex]
					metadata.Validate()
				}
			}

			// Course []Metadata
			metadatas := course.Metadata
			for metadataIndex := range metadatas {
				metadata := metadatas[metadataIndex]
				metadata.Validate()
			}
		}
	}

	for registrations := range uni.Registrations {
		_ = uni.Registrations[registrations]

	}

	// university []Metadata
	metadatas := uni.Metadata
	for metadataIndex := range metadatas {
		metadata := metadatas[metadataIndex]
		metadata.Validate()

	}

	return nil
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

func CheckUniqueSubject(subjects []*Subject) {
	m := make(map[string]int)
	for subjectIndex := range subjects {
		subject := subjects[subjectIndex]
		key := subject.Season + subject.Year + subject.Name + subject.Number
		m[key]++
		if m[key] > 1 {
			log.WithFields(log.Fields{"key": key, "count": m[key]}).Debugln("Duplicate subject")
			subject.Name = subject.Name + "_" + strconv.Itoa(m[key])
		}
	}
}

func CheckUniqueCourse(subject *Subject, courses []*Course) {
	m := map[string]int{}
	for courseIndex := range courses {
		course := courses[courseIndex]
		key := course.Name + course.Number
		m[key]++
		if m[key] > 1 {
			log.WithFields(log.Fields{"subject": subject.Name,
				"season": subject.Season,
				"year":   subject.Year,
				"key":    key,
				"count":  m[key]}).Debugln("Duplicate course")
			course.Name = course.Name + "_" + strconv.Itoa(m[key])
		}
	}
}
