package main

import (
	"bufio"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"log"
	"os"
	"strconv"
	uct "uct/common"
)

var (
	app    = kingpin.New("print", "An application to print and translate json and protobuf")
	format = app.Flag("format", "choose file input format").Short('f').HintOptions("protobuf", "json").PlaceHolder("[protobuf, json]").Required().String()
	out    = app.Flag("output", "output format").Short('o').HintOptions("protobuf", "json").PlaceHolder("[protobuf, json]").String()
	file   = app.Arg("input", "file to print").File()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *format != "json" && *format != "protobuf" {
		log.Fatalln("Invalid format:", *format)
	}

	var input *bufio.Reader
	if *file != nil {
		input = bufio.NewReader(*file)
	} else {
		input = bufio.NewReader(os.Stdin)
	}

	var university uct.University
	uct.UnmarshallMessage(*format, input, &university)

	clean(&university)

	if *out != "" {
		output := uct.MarshalMessage(*out, university)
		io.Copy(os.Stdout, output)
	}
}

func CheckUniqueSubject(subjects []*uct.Subject) {
	m := make(map[string]int)
	for subjectIndex := range subjects {
		subject := subjects[subjectIndex]
		key := subject.Season + subject.Year + subject.Name
		m[key]++
		if m[key] > 1 {
			uct.Log("Duplicate subject found:", key, " c:", m[key])
			subject.Name = subject.Name + "_" + strconv.Itoa(m[key])
		}
	}
}

func CheckUniqueCourse(subject *uct.Subject, courses []*uct.Course) {
	m := map[string]int{}
	for courseIndex := range courses {
		course := courses[courseIndex]
		key := course.Name + course.Number
		m[key]++
		if m[key] > 1 {
			uct.Log("subject", subject.Name, subject.Season)
			uct.Log("Duplicate course found: ", key, " c:", m[key])
			course.Name = course.Name + "_" + strconv.Itoa(m[key])
		}
	}
}

func pad(str string, num int) (out string) {
	for i := 0; i < num; i++ {
		out += str
	}
	return
}

func clean(uni *uct.University) {

	uni.Validate()
	CheckUniqueSubject(uni.Subjects)
	for subjectIndex := range uni.Subjects {
		subject := uni.Subjects[subjectIndex]
		subject.Validate(uni)

		// Old
		subject.UniversityName = uni.Name

		// New
		subject.SubjectName = subject.Name
		subject.SubjectNumber = subject.Number
		subject.SubjectSeason = subject.Season
		subject.SubjectYear = subject.Year

		courses := subject.Courses
		CheckUniqueCourse(subject, courses)
		for courseIndex := range courses {
			course := courses[courseIndex]
			course.Validate(subject)

			// Old
			course.UniversityName = uni.Name
			course.SubjectName = subject.Name
			course.SubjectNumber = subject.Number
			course.SubjectSeason = subject.Season
			course.SubjectYear = subject.Year

			// New
			course.CourseName = course.Name
			course.CourseNumber = course.Number

			sections := course.Sections
			for sectionIndex := range sections {
				section := sections[sectionIndex]
				section.Validate(course)

				// Old
				section.UniversityName = course.UniversityName
				section.SubjectName = course.SubjectName
				section.SubjectNumber = course.SubjectNumber
				section.SubjectSeason = course.SubjectSeason
				section.SubjectYear = course.SubjectYear
				section.CourseName = course.CourseName
				section.CourseNumber = course.CourseNumber

				//[]Instructors
				instructors := section.Instructors
				for instructorIndex := range instructors {
					instructor := instructors[instructorIndex]
					instructor.Validate()

				}

				//[]Meeting
				meetings := section.Meetings
				for meetingIndex := range meetings {
					meeting := meetings[meetingIndex]
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
}
