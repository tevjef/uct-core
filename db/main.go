package main

import (
	"bufio"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pquerna/ffjson/ffjson"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime/pprof"
	"time"
	uct "uct/common"
)

var (
	database *sqlx.DB
	sem      = make(chan int, 50)
	input    *bufio.Reader
)

var (
	debugBool      bool
	app            = kingpin.New("uct-db", "A command-line application for inserting and updated university information")
	add            = app.Command("add", "A command-line application for inserting and updated university information").Hidden().Default()
	fullUpsert     = add.Flag("insert-all", "Full insert/update of all json objects.").Default("true").Short('i').Bool()
	file           = add.Flag("file", "File to read university data from.").Short('f').File()
	verbose        = add.Flag("verbose", "Verbose log of object representations.").Default("false").Short('v').Bool()
	debug          = app.Command("debug", "Enable debug mode.")
	server         = debug.Flag("server", "Debug server address to enable profiling.").PlaceHolder("hostname:port").Default("127.0.0.1:6060").TCP()
	cpuprofile     = debug.Flag("cpuprofile", "Write cpu profile to file.").PlaceHolder("cpu.pprof").String()
	memprofile     = debug.Flag("memprofile", "Write memory profile to file.").PlaceHolder("mem.pprof").String()
	memprofileRate = debug.Flag("memprofile-rate", "Rate at which memory is profiled.").Default("20s").Duration()
)

var (
	insertions    int
	updates       int
	upserts       int
	existential   int
	subjectCount  int
	courseCount   int
	sectionCount  int
	meetingCount  int
	metadataCount int
)


func initDB(connection string) *sqlx.DB {
	database, err := sqlx.Open("postgres", connection)
	if err != nil {
		log.Fatalln(err)
	}
	return database
}

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "debug":
		debugBool = true
		break
	case "add":
		debugBool = false
	}

	if *server != nil && debugBool {
		go func() {
			log.Println("**Starting debug server on...", (*server).String())
			log.Println(http.ListenAndServe((*server).String(), nil))
		}()
	}

	if *cpuprofile != "" && debugBool {
		if f, err := os.Create(*cpuprofile); err != nil {
			log.Println(err)
		} else {
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()

		}
	}

	if *memprofile != "" && debugBool {
		ticker := time.NewTicker(*memprofileRate)
		go func() {
			for t := range ticker.C {
				if f, err := os.Create(*memprofile + "-" + t.String()); err != nil {
					log.Println(err)
				} else {
					pprof.WriteHeapProfile(f)
				}
			}
		}()
	}

	database = initDB(uct.GetUniversityDB(true))

	if *file != nil {
		input = bufio.NewReader(*file)
	} else {
		input = bufio.NewReader(os.Stdin)
	}

	dec := ffjson.NewDecoder()
	var universities []uct.University
	if err := dec.DecodeReader(input, &universities); err != nil {
		log.Fatal(err)
	}

	for _, university := range universities {
		insertUniversity(database, university)
	}

	uct.Log("Insertions: ", insertions)
	uct.Log("Updates: ", updates)
	uct.Log("Upserts: ", upserts)
	uct.Log("Existential: ", existential)
	uct.Log("Subjects: ", subjectCount)
	uct.Log("Courses: ", courseCount)
	uct.Log("Sections: ", sectionCount)
	uct.Log("Meetings: ", metadataCount)
	uct.Log("Metadata: ", metadataCount)
	defer func() {
		database.Close()
		pprof.StopCPUProfile()
	}()

}

func insertUniversity(db *sqlx.DB, uni uct.University) {
	//uni.VetAndBuild()
	LogVerbose(fmt.Sprintln("In University:", uni.Name))
	var university_id int64

	insertQuery := `INSERT INTO university (name, abbr, home_page, registration_page, main_color, accent_color, topic_name)
                    VALUES (:name, :abbr, :home_page, :registration_page, :main_color, :accent_color, :topic_name)
                    RETURNING university.id`

	updateQuery := `UPDATE university SET (abbr, home_page, registration_page, main_color, accent_color, topic_name) =
	                (:abbr, :home_page, :registration_page, :main_color, :accent_color, :topic_name)
	                WHERE name = :name
	                RETURNING university.id`

	university_id = upsert(db, insertQuery, updateQuery, uni)

	for subjectIndex := 0; subjectIndex < len(uni.Subjects); subjectIndex++ {
		uni.Subjects[subjectIndex].UniversityId = university_id
		subId := insertSubject(db, uni.Subjects[subjectIndex])

		courses := uni.Subjects[subjectIndex].Courses
		for courseIndex := 0; courseIndex < len(courses); courseIndex++ {
			courses[courseIndex].SubjectId = subId
			course := courses[courseIndex]
			courseId := insertCourse(db, course)

			sections := course.Sections
			for sectionIndex := 0; sectionIndex < len(sections); sectionIndex++ {
				sections[sectionIndex].CourseId = courseId
				sectionId := insertSection(db, sections[sectionIndex])

				instructors := sections[sectionIndex].Instructors
				for instructorIndex := 0; instructorIndex < len(instructors); instructorIndex++ {
					instructors[instructorIndex].SectionId = sectionId
					insertInstructor(db, instructors[instructorIndex])
				}

				meetings := sections[sectionIndex].Meetings
				for meetingIndex := 0; meetingIndex > len(meetings); meetingIndex++ {
					meetings[meetingIndex].SectionId = sectionId
					meetings[meetingIndex].Index = meetingIndex
					meetingId := insertMeeting(db, meetings[meetingIndex])

					metadata := meetings[meetingIndex].Metadata
					for metadataIndex := 0; metadataIndex < len(metadata); metadataIndex++ {
						metadata[metadataIndex].MeetingId.Int64 = meetingId
						insertMetadata(db, metadata[metadataIndex])
					}
				}

				for _, book := range sections[sectionIndex].Books {
					book.SectionId = sectionId
					insertBook(db, book)
				}

				metadata := sections[sectionIndex].Metadata
				for metadataIndex := 0; metadataIndex < len(metadata); metadataIndex++ {
					metadata[metadataIndex].SectionId.Int64 = sectionId
					insertMetadata(db, metadata[metadataIndex])
				}
			}

			metadata := courses[courseIndex].Metadata
			for metadataIndex := 0; metadataIndex < len(metadata); metadataIndex++ {
				metadata[metadataIndex].CourseId.Int64 = courseId
				insertMetadata(db, metadata[metadataIndex])
			}
		}

		for _, metadata := range uni.Subjects[subjectIndex].Metadata {
			metadata.SubjectId.Int64 = subId
			insertMetadata(db, metadata)
		}
	}

	for _, registrations := range uni.Registrations {
		registrations.UniversityId = university_id
		insertRegistration(db, registrations)
	}

	for _, metadata := range uni.Metadata {
		metadata.UniversityId.Int64 = university_id
		insertMetadata(db, metadata)
	}
}

func insertSubject(db *sqlx.DB, sub uct.Subject) (subject_id int64) {
	//sub.VetAndBuild()
	subjectCount++
	LogVerbose(fmt.Sprintln("In Subjects:", sub.UniversityId, sub.Name, sub.Number, sub.Season, sub.Year, sub.Hash, "Course: len ", len(sub.Courses)))

	if !*fullUpsert {
		existsQuery := `SELECT id FROM course
						WHERE hash = :hash`

		if subject_id = exists(db, existsQuery, sub); subject_id != 0 {
			return
		}
	}

	insertQuery := `INSERT INTO subject (university_id, name, number, season, year, hash, topic_name)
                   	VALUES  (:university_id, :name, :number, :season, :year, :hash, :topic_name) 
                   	RETURNING subject.id`

	updateQuery := `UPDATE subject SET (name, season, year, topic_name) = (:name, :season, :year, :topic_name)
					WHERE hash = :hash
                   	RETURNING subject.id`

	subject_id = upsert(db, insertQuery, updateQuery, sub)

	return subject_id
}

func insertCourse(db *sqlx.DB, course uct.Course) (course_id int64) {
	//course.VetAndBuild()
	courseCount++
	LogVerbose(fmt.Sprintln("In Course:", course.SubjectId, course.Name, course.Number, course.Hash, "Section len: ", len(course.Sections)))

	if !*fullUpsert {
		existsQuery := `SELECT id FROM course
						WHERE hash = :hash`

		if course_id = exists(db, existsQuery, course); course_id != 0 {
			return
		}
	}

	updateQuery := `UPDATE course SET (name, synopsis, topic_name) = (:name, :synopsis, :topic_name)
					WHERE hash = :hash
                    RETURNING course.id`

	insertQuery := `INSERT INTO course (subject_id, name, number, synopsis, hash, topic_name)
                    VALUES  (:subject_id, :name, :number, :synopsis, :hash, :topic_name) 
                    RETURNING course.id`

	course_id = upsert(db, insertQuery, updateQuery, course)

	return course_id
}

func insertSection(db *sqlx.DB, section uct.Section) (section_id int64) {
	//section.VetAndBuild()
	sectionCount++
	LogVerbose(fmt.Sprintln("In Section:", section.CourseId, section.Number, section.CallNumber, section.Status, section.Meetings, section.Instructors, section.Books))

	updateQuery := `UPDATE section SET (max, now, status, credits, topic_name) = (:max, :now, :status, :credits, :topic_name)
					WHERE course_id = :course_id AND call_number = :call_number AND number = :number
                    RETURNING section.id`

	insertQuery := `INSERT INTO section (course_id, number, call_number, max, now, status, credits, topic_name)
                    VALUES  (:course_id, :number, :call_number, :max, :now, :status, :credits, :topic_name)
                    RETURNING section.id`

	section_id = upsert(db, insertQuery, updateQuery, section)

	return section_id
}

func insertMeeting(db *sqlx.DB, meeting uct.Meeting) (meeting_id int64) {
	//meeting.VetAndBuild()
	meetingCount++
	LogVerbose(fmt.Sprintln("In Meeting:", meeting))

	if !*fullUpsert {
		existsQuery := `SELECT id FROM meeting
						WHERE section_id = :section_id AND index = :index`

		if meeting_id = exists(db, existsQuery, meeting); meeting_id != 0 {
			return
		}
	}

	updateQuery := `UPDATE meeting SET (room, day, start_time, end_time) = (:room, :day, :start_time, :end_time)
					WHERE section_id = :section_id AND index = :index
                    RETURNING meeting.id`

	insertQuery := `INSERT INTO meeting (section_id, room, day, start_time, end_time, index)
                    VALUES  (:section_id, :room, :day, :start_time, :end_time, :index)
                    RETURNING meeting.id`

	meeting_id = upsert(db, insertQuery, updateQuery, meeting)

	return meeting_id
}

func insertInstructor(db *sqlx.DB, instructor uct.Instructor) (instructor_id int64) {
	//instructor.VetAndBuild()
	LogVerbose(fmt.Sprintln("In Instructor:", instructor))

	existsQuery := `SELECT id FROM instructor
				WHERE section_id = :section_id AND name = :name`

	if instructor_id = exists(db, existsQuery, instructor); instructor_id != 0 {
		return
	}

	insertQuery := `INSERT INTO instructor (section_id, name)
				VALUES  (:section_id, :name)
				RETURNING instructor.id`
	instructor_id = insert(db, insertQuery, instructor)

	return instructor_id
}

func insertBook(db *sqlx.DB, book uct.Book) (book_id int64) {
	LogVerbose(fmt.Sprintln("In Book:", book))

	updateQuery := `UPDATE book SET (title, url) = (:title, :url)
					WHERE section_id = :section_id AND title = :title
                    RETURNING book.id`

	insertQuery := `INSERT INTO meeting (section_id, title, url)
                    VALUES  (:section_id, :title, :url)
                    RETURNING book.id`

	book_id = upsert(db, insertQuery, updateQuery, book)

	return book_id
}

func insertRegistration(db *sqlx.DB, registration uct.Registration) int64 {
	LogVerbose(fmt.Sprintln("In Registration:", registration))
	var registration_id int64

	updateQuery := `UPDATE registration SET (period_date) = (:period_date) 
					WHERE university_id = :university_id AND period = :period
	                RETURNING registration.id`

	insertQuery := `INSERT INTO registration (university_id, period, period_date)
                    VALUES (:university_id, :period, :period_date) 
                    RETURNING registration.id`

	registration_id = upsert(db, insertQuery, updateQuery, registration)

	return registration_id
}

func insertMetadata(db *sqlx.DB, metadata uct.Metadata) (metadata_id int64) {
	metadataCount++
	LogVerbose(fmt.Sprintln("In Metadata:", metadata))

	var insertQuery string
	var updateQuery string
	if metadata.UniversityId.Int64 != 0 {
		if !*fullUpsert {
			existsQuery := `SELECT id FROM metadata
						WHERE metadata.university_id = :university_id AND metadata.title = :title`

			if metadata_id = exists(db, existsQuery, metadata); metadata_id != 0 {
				return
			}
		}

		updateQuery = `UPDATE metadata SET (title, content) = (:title, :content) 
					   WHERE metadata.university_id = :university_id AND metadata.title = :title
		               RETURNING metadata.id`

		insertQuery = `INSERT INTO metadata (university_id, title, content)
                       VALUES (:university_id, :title, :content) 
                       RETURNING metadata.id`

	} else if metadata.SubjectId.Int64 != 0 {
		if !*fullUpsert {
			existsQuery := `SELECT id FROM metadata
						WHERE metadata.subject_id = :subject_id AND metadata.title = :title`

			if metadata_id = exists(db, existsQuery, metadata); metadata_id != 0 {
				return
			}
		}

		updateQuery = `UPDATE metadata SET (title, content) = (:title, :content) 
					   WHERE metadata.subject_id = :subject_id AND metadata.title = :title
		               RETURNING metadata.id`

		insertQuery = `INSERT INTO metadata (subject_id, title, content)
                       VALUES (:subject_id, :title, :content) 
                       RETURNING metadata.id`

	} else if metadata.CourseId.Int64 != 0 {
		if !*fullUpsert {
			existsQuery := `SELECT id FROM metadata
						WHERE metadata.course_id = :course_id AND metadata.title = :title`

			if metadata_id = exists(db, existsQuery, metadata); metadata_id != 0 {
				return
			}
		}

		updateQuery = `UPDATE metadata SET (title, content) = (:title, :content) 
					   WHERE metadata.course_id = :course_id AND metadata.title = :title
		               RETURNING metadata.id`

		insertQuery = `INSERT INTO metadata (course_id, title, content)
                       VALUES (:course_id, :title, :content) 
                       RETURNING metadata.id`

	} else if metadata.SectionId.Int64 != 0 {
		if !*fullUpsert {

			existsQuery := `SELECT id FROM metadata
						WHERE metadata.section_id = :section_id AND metadata.title = :title`

			if metadata_id = exists(db, existsQuery, metadata); metadata_id != 0 {
				return
			}
		}

		updateQuery = `UPDATE metadata SET (title, content) = (:title, :content) 
					   WHERE metadata.section_id = :section_id AND metadata.title = :title
		               RETURNING metadata.id`

		insertQuery = `INSERT INTO metadata (section_id, title, content)
                       VALUES (:section_id, :title, :content) 
                       RETURNING metadata.id`

	} else if metadata.MeetingId.Int64 != 0 {
		if !*fullUpsert {
			existsQuery := `SELECT id FROM metadata
						WHERE metadata.meeting_id = :meeting_id AND metadata.title = :title`

			if metadata_id = exists(db, existsQuery, metadata); metadata_id != 0 {
				return
			}
		}

		updateQuery = `UPDATE metadata SET (title, content) = (:title, :content) 
					   WHERE metadata.meeting_id = :meeting_id AND metadata.title = :title
		               RETURNING metadata.id`

		insertQuery = `INSERT INTO metadata (meeting_id, title, content)
                       VALUES (:meeting_id, :title, :content) 
                       RETURNING metadata.id`
	}

	metadata_id = upsert(db, insertQuery, updateQuery, metadata)

	return metadata_id
}

func insert(db *sqlx.DB, query string, data interface{}) (id int64) {
	insertions++
	typeName := reflect.TypeOf(data).Name()


	if rows, err := db.NamedQuery(query, data); err != nil {
		log.Panicln(err, typeName)
	} else {
		for rows.Next() {
			if err = rows.Scan(&id); err != nil {
				log.Panicln(err, typeName)
			}
			rows.Close()
			uct.Log("Insert: ", typeName, " ", id)
		}
	}
	return id
}

func update(db *sqlx.DB, query string, data interface{}) (id int64) {
	updates++
	typeName := reflect.TypeOf(data).Name()

	if rows, err := db.NamedQuery(query, data); err != nil {
		log.Panicln(err, typeName)
	} else {
		for rows.Next() {
			if err = rows.Scan(&id); err != nil {
				log.Panicln(err, typeName)
			}
			rows.Close()
			uct.Log("Update: ", typeName, " ", id)
		}
	}

	return id
}

func upsert(db *sqlx.DB, insertQuery, updateQuery string, data interface{}) (id int64) {
	upserts++
	if id = update(db, updateQuery, data); id != 0 {
	} else if id == 0 {
		id = insert(db, insertQuery, data)
	}
	return
}

func exists(db *sqlx.DB, query string, data interface{}) (id int64) {
	existential++
	if rows, err := db.NamedQuery(query, data); err != nil {
		log.Panicln(err)
	} else {
		for rows.Next() {
			if err = rows.Scan(&id); err != nil {
				log.Panicln(err)
			}
		}
	}

	return
}

func LogVerbose(v ...interface{}) {
	if *verbose {
		uct.LogVerbose(v)
	}
}

type DataType int

const (
	SUBJECT DataType = iota
	COURSE
	SECTION
)

var datatype = [...]string{
	"subject",
	"course",
	"section",
}

func (s DataType) String() string {
	return datatype[s]
}
