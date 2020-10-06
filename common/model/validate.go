package model

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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

var NUL []byte = []byte("\x00")
var SOH []byte = []byte("\x01")

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
	return isLetter || isNumber || isSpace || isDash || isUnderscore || isDot || isTilde || isPercent
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

func ValidateAll(university *University) (err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in ValidateAll", r)
			err = fmt.Errorf("%v+", r)
		}
	}()

	if err := university.Validate(); err != nil {
		return err
	}

	if err := ValidateAllSubjects(university); err != nil {
		return err
	}

	// university []Metadata
	metadata := university.Metadata
	for metadataIndex := range metadata {
		metadata := metadata[metadataIndex]
		if err := metadata.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func ValidateAllSubjects(university *University) error {
	makeUniqueSubjects(university.Subjects)

	if len(university.Subjects) == 0 {
		log.Warningln("validate: no subjects in university: %v", university.TopicName)
	} else {
		sort.Sort(subjectSorter{university.Subjects})
	}

	var duplicateCourses []string

	for subjectIndex := range university.Subjects {
		subject := university.Subjects[subjectIndex]
		if err := subject.Validate(university); err != nil {
			return err
		}

		duplicateCourses = append(duplicateCourses, makeUniqueCourses(subject, subject.Courses)...)

		if err := ValidateAllCourses(subject); err != nil {
			return err
		}
	}

	if len(duplicateCourses) > 0 {
		log.WithFields(log.Fields{
			"courses": duplicateCourses,
		}).Debugf("validate: duplicate courses found: %v", len(duplicateCourses))
	}

	return nil
}

func ValidateAllCourses(subject *Subject) error {
	courses := subject.Courses
	for courseIndex := range courses {
		course := courses[courseIndex]
		if err := course.Validate(subject); err != nil {
			return err
		}

		if err := ValidateAllSections(course); err != nil {
			return err
		}

		// Course []Metadata
		metadata := course.Metadata
		for metadataIndex := range metadata {
			metadata := metadata[metadataIndex]
			if err := metadata.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

func ValidateAllSections(course *Course) error {
	sections := course.Sections

	for sectionIndex := range sections {
		section := sections[sectionIndex]
		if err := section.Validate(course); err != nil {
			return err
		}

		//[]Instructors
		instructors := section.Instructors
		for instructorIndex := range instructors {
			instructor := instructors[instructorIndex]
			instructor.Index = int32(instructorIndex)
			if err := instructor.Validate(); err != nil {
				return err
			}
		}

		//[]Meeting
		meetings := section.Meetings
		for meetingIndex := range meetings {
			meeting := meetings[meetingIndex]
			meeting.Index = int32(meetingIndex)
			if err := meeting.Validate(); err != nil {
				return err
			}

			// Meeting []Metadata
			metadatas := meeting.Metadata
			for metadataIndex := range metadatas {
				metadata := metadatas[metadataIndex]
				if err := metadata.Validate(); err != nil {
					return err
				}
			}
		}

		//[]Books
		books := section.Books
		for bookIndex := range books {
			book := books[bookIndex]
			if err := book.Validate(); err != nil {
				return err
			}
		}

		// Section []Metadata
		metadatas := section.Metadata
		for metadataIndex := range metadatas {
			metadata := metadatas[metadataIndex]
			if err := metadata.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (u *University) Validate() error {
	// Name
	if u.Name == "" {
		return errors.New("University name == is empty")
	}

	// Abbr
	if u.Abbr == "" {
		u.Abbr = strings.Map(func(r rune) rune {
			if unicode.IsUpper(r) {
				return r
			}

			return -1
		}, u.Name)
	}

	// Homepage
	if u.HomePage == "" {
		return errors.New("HomePage == is empty")
	}

	u.HomePage = trim(u.HomePage)

	if homePageUrl, err := url.ParseRequestURI(u.HomePage); err != nil {
		return err
	} else {
		u.HomePage = homePageUrl.String()
	}

	// RegistrationPage
	if u.RegistrationPage == "" {
		return errors.New("RegistrationPage == is empty")
	}

	u.RegistrationPage = trim(u.RegistrationPage)

	if registrationPageUrl, err := url.ParseRequestURI(u.RegistrationPage); err != nil {
		return err
	} else {
		u.RegistrationPage = registrationPageUrl.String()
	}

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
		return errors.New("Registration != 12 ")
	}

	if u.ResolvedSemesters == nil {
		return errors.New("ResolvedSemesters is nil")
	}

	if u.ResolvedSemesters.Current == nil {
		return errors.New("ResolvedSemesters.Current is nil")
	}

	u.TopicName = ToTopicName(u.Name)
	u.TopicId = ToTopicId(u.TopicName)

	return nil
}

func (sub *Subject) Validate(uni *University) error {
	// Name
	if sub.Name == "" {
		return errors.New("Subject name == is empty")
	}

	sub.Name = TrimAll(sub.Name)
	sub.Name = ToTitle(sub.Name)

	// TopicName
	sub.TopicName = strings.Join([]string{uni.TopicName, sub.Number, sub.Name, sub.Season, sub.Year}, ".")
	sub.TopicName = ToTopicName(sub.TopicName)
	sub.TopicId = ToTopicId(sub.TopicName)

	if len(sub.Courses) == 0 {
		log.Warningf("validate: no course in subject: %v", sub.TopicName)
	} else {
		sort.Sort(courseSorter{sub.Courses})
	}

	return nil
}

func (course *Course) Validate(subject *Subject) error {
	// Name
	if course.Name == "" {
		return errors.New("Course name == is empty: " + course.Name)
	}

	course.Name = TrimAll(course.Name)
	course.Name = ToTitle(course.Name)

	// Number
	if course.Number == "" {
		return errors.New("Number == is empty")
	}

	// Synopsis
	if course.Synopsis != nil {
		temp := TrimAll(*course.Synopsis)
		//temp = utf8string.NewString(*course.Synopsis).String()
		course.Synopsis = &temp
	}

	// TopicName
	course.TopicName = strings.Join([]string{subject.TopicName, course.Number, course.Name}, ".")
	course.TopicName = ToTopicName(course.TopicName)
	course.TopicId = ToTopicId(course.TopicName)
	if len(course.Sections) == 0 {
		log.Errorln("validate: no section in course: %v", course.TopicName)
	} else {
		sort.Stable(sectionSorter{course.Sections})
	}

	return nil
}

// Validate within the context for these enclosing objects
func (section *Section) Validate(course *Course) error {
	// Number
	if section.Number == "" {
		return errors.New("Number == is empty")
	}

	section.Number = trim(section.Number)

	// Call Number
	if section.CallNumber == "" {
		return errors.New("CallNumber == is empty")
	}

	section.CallNumber = trim(section.CallNumber)

	// Status
	if section.Status == "" {
		return errors.New("Status == is empty")
	} else if strings.ToLower(section.Status) != "open" && strings.ToLower(section.Status) != "closed" {
		return errors.New("Status != open || status != closed status=" + section.Status)
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

	// Credits must be a numeric type
	if section.Credits == "" {
		return errors.New("Credits == is empty")
	} else if _, err := strconv.ParseFloat(section.Credits, 64); err != nil {
		return err
	}

	section.TopicName = strings.Join([]string{course.TopicName, section.Number, section.CallNumber}, ".")
	section.TopicName = ToTopicName(section.TopicName)
	section.TopicId = ToTopicId(section.TopicName)
	sort.Stable(instructorSorter{section.Instructors})

	return nil
}

func (meeting *Meeting) Validate() error {
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

	return nil
}

func (instructor *Instructor) Validate() error {

	if instructor.Name == "" {
		return errors.New("Instructor name == is empty")
	}

	if instructor.Name[len(instructor.Name)-1:] == "-" {
		instructor.Name = instructor.Name[:len(instructor.Name)-1]
	}

	instructor.Name = trim(instructor.Name)

	return nil
}

func (book *Book) Validate() error {
	if book.Title == "" {
		return errors.New("Title  == is empty")
	}

	book.Title = trim(book.Title)

	if book.Url == "" {
		return errors.New("Url == is empty")
	}

	book.Url = trim(book.Url)

	if url, err := url.ParseRequestURI(book.Url); err != nil {
		return err
	} else {
		book.Url = url.String()
	}

	return nil
}

func (metaData *Metadata) Validate() error {
	// Title
	if metaData.Title == "" {
		return errors.New("Title == is empty")
	}

	metaData.Title = trim(metaData.Title)

	// Content
	if metaData.Content == "" {
		return errors.New("Content == is empty")
	}

	metaData.Content = trim(metaData.Content)

	return nil
}

func makeUniqueSubjects(subjects []*Subject) {
	m := make(map[string]int)
	var duplicates []string
	for subjectIndex := range subjects {
		subject := subjects[subjectIndex]
		key := strings.Join([]string{subject.Season, subject.Year, subject.Name, subject.Number}, "")
		m[key]++
		if m[key] > 1 {
			subject.Name = subject.Name + "_" + strconv.Itoa(m[key])
			duplicates = append(duplicates, fmt.Sprintf("count: %v | %v ", m[key], key))
		}
	}
	if len(duplicates) > 0 {
		log.WithFields(log.Fields{"subjects": duplicates}).Debugf("validate: duplicate subjects found: %v", len(duplicates))
	}
}

func makeUniqueCourses(subject *Subject, courses []*Course) []string {
	m := map[string]int{}
	var duplicates []string
	for courseIndex := range courses {
		course := courses[courseIndex]
		key := strings.Join([]string{course.Name, course.Number}, "")
		m[key]++
		if m[key] > 1 {
			course.Name = course.Name + "_" + strconv.Itoa(m[key])
			duplicates = append(duplicates, fmt.Sprintf("subject: %v sectionCount:%v | %v", subject.TopicName, len(course.Sections), key))
		}
	}

	return duplicates
}
