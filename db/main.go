package main

import (
	"bufio"
	"github.com/golang/protobuf/proto"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pquerna/ffjson/ffjson"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"reflect"
	"runtime/pprof"
	"time"
	uct "uct/common"
)

type App struct {
	dbHandler DatabaseHandler
}

var (
	debugBool      bool
	app            = kingpin.New("uct-db", "A command-line application for inserting and updated university information")
	add            = app.Command("add", "A command-line application for inserting and updated university information").Hidden().Default()
	fullUpsert     = add.Flag("insert-all", "Full insert/update of all json objects.").Default("true").Short('a').Bool()
	file           = add.Flag("input", "File to read university data from.").Short('i').File()
	format         = app.Flag("format", "Choose input format").Short('f').HintOptions("protobuf", "json").PlaceHolder("[protobuf, json]").Required().String()
	debug          = app.Command("debug", "Enable debug mode.")
	cpuprofile     = debug.Flag("cpuprofile", "Write cpu profile to file.").PlaceHolder("cpu.pprof").String()
	memprofile     = debug.Flag("memprofile", "Write memory profile to file.").PlaceHolder("mem.pprof").String()
	memprofileRate = debug.Flag("memprofile-rate", "Ratae at which memory is profiled.").Default("20s").Duration()
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

	if *format != "json" && *format != "protobuf" {
		log.Fatalln("Invalid format:", *format)
	}

	go func() {
		log.Println("**Starting debug server on...", uct.DB_DEBUG_SERVER)
		log.Println(http.ListenAndServe(uct.DB_DEBUG_SERVER, nil))
	}()

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

	var input *bufio.Reader
	if *file != nil {
		input = bufio.NewReader(*file)
	} else {
		input = bufio.NewReader(os.Stdin)
	}

	var university uct.University

	if *format == "json" {
		dec := ffjson.NewDecoder()
		if err := dec.DecodeReader(input, &university); err != nil {
			log.Fatalln("Failed to parse university:", err)
		}
	} else if *format == "protobuf" {
		data, err := ioutil.ReadAll(input)
		if err = proto.Unmarshal(data, &university); err != nil {
			log.Fatalln("Failed to parse university:", err)
		}
	}

	// Initialize database connection
	database := initDB(uct.GetUniversityDB())
	dbHandler := DatabaseHandlerImpl{Database: database}
	dbHandler.PrepareAllStmts()
	app := App{dbHandler: dbHandler}

	// Start logging with influx
	go audit()

	// Was originally designed to insert an array of universities.
	startAudit <- true
	app.insertUniversity(university)
	endAudit <- true

	// Before main ends, close the database and stop writing profile
	defer func() {
		database.Close()
		pprof.StopCPUProfile()
	}()

}

func (app App) insertUniversity(uni uct.University) {
	university_id := app.dbHandler.upsert(UniversityInsertQuery, UniversityUpdateQuery, uni)

	subjectCountCh <- len(uni.Subjects)
	for subjectIndex := range uni.Subjects {
		subject := uni.Subjects[subjectIndex]
		subject.UniversityId = university_id

		subjectId := app.insertSubject(subject)

		courses := subject.Courses
		courseCountCh <- len(courses)
		for courseIndex := range courses {
			course := courses[courseIndex]

			course.SubjectId = subjectId
			courseId := app.insertCourse(course)

			sections := course.Sections
			sectionCountCh <- len(sections)
			for sectionIndex := range sections {
				section := sections[sectionIndex]

				section.CourseId = courseId
				sectionId := app.insertSection(section)

				//[]Instructors
				instructors := section.Instructors
				for instructorIndex := range instructors {
					instructor := instructors[instructorIndex]

					instructor.SectionId = sectionId
					app.insertInstructor(instructor)
				}

				//[]Meeting
				meetings := section.Meetings
				meetingCountCh <- len(meetings)
				for meetingIndex := range meetings {
					meeting := meetings[meetingIndex]

					meeting.SectionId = sectionId
					meeting.Index = int32(meetingIndex)
					meetingId := app.insertMeeting(meeting)

					// Meeting []Metadata
					metadatas := meeting.Metadata
					metadataCountCh <- len(metadatas)
					for metadataIndex := range metadatas {
						metadata := metadatas[metadataIndex]

						metadata.MeetingId = &meetingId
						app.insertMetadata(metadata)
					}
				}

				//[]Books
				books := section.Books
				for bookIndex := range books {
					book := books[bookIndex]

					book.SectionId = sectionId
					app.insertBook(book)
				}

				// Section []Metadata
				metadatas := section.Metadata
				metadataCountCh <- len(metadatas)
				for metadataIndex := range metadatas {
					metadata := metadatas[metadataIndex]

					metadata.SectionId = &sectionId
					app.insertMetadata(metadata)
				}
			}

			// Course []Metadata
			metadatas := course.Metadata
			metadataCountCh <- len(metadatas)
			for metadataIndex := range metadatas {
				metadata := metadatas[metadataIndex]

				metadata.CourseId = &courseId
				app.insertMetadata(metadata)
			}
		}
	}

	for _, registrations := range uni.Registrations {
		registrations.UniversityId = university_id
		app.insertRegistration(registrations)
	}

	// university []Metadata
	metadatas := uni.Metadata
	for metadataIndex := range metadatas {
		metadata := metadatas[metadataIndex]

		metadata.UniversityId = &university_id
		app.insertMetadata(metadata)
	}
}

func (app App) insertSubject(sub uct.Subject) (subject_id int64) {
	//sub.VetAndBuild()
	if !*fullUpsert {

		if subject_id = app.dbHandler.exists(SubjectExistQuery, sub); subject_id != 0 {
			return
		}
	}
	subject_id = app.dbHandler.upsert(SubjectInsertQuery, SubjectUpdateQuery, sub)

	// Subject []Metadata
	metadatas := sub.Metadata
	for metadataIndex := range metadatas {
		metadata := metadatas[metadataIndex]

		metadata.SubjectId = &subject_id
		app.insertMetadata(metadata)
	}
	return subject_id
}

func (app App) insertCourse(course uct.Course) (course_id int64) {
	if !*fullUpsert {

		if course_id = app.dbHandler.exists(CourseExistQuery, course); course_id != 0 {
			return
		}
	}
	course_id = app.dbHandler.upsert(CourseInsertQuery, CourseUpdateQuery, course)

	return course_id
}

func (app App) insertSection(section uct.Section) (section_id int64) {
	section_id = app.dbHandler.upsert(SectionInsertQuery, SectionUpdateQuery, section)
	return section_id
}

func (app App) insertMeeting(meeting uct.Meeting) (meeting_id int64) {
	if !*fullUpsert {
		if meeting_id = app.dbHandler.exists(MeetingExistQuery, meeting); meeting_id != 0 {
			return
		}
	}
	meeting_id = app.dbHandler.upsert(MeetingInsertQuery, MeetingUpdateQuery, meeting)
	return meeting_id
}

func (app App) insertInstructor(instructor uct.Instructor) (instructor_id int64) {
	if instructor_id = app.dbHandler.exists(InstructorExistQuery, instructor); instructor_id != 0 {
		return
	}

	instructor_id = app.dbHandler.insert(InstructorInsertQuery, instructor)

	return instructor_id
}

func (app App) insertBook(book uct.Book) (book_id int64) {
	book_id = app.dbHandler.upsert(BookInsertQuery, BookUpdateQuery, book)

	return book_id
}

func (app App) insertRegistration(registration uct.Registration) int64 {
	var registration_id int64
	registration_id = app.dbHandler.upsert(RegistrationInsertQuery, RegistrationUpdateQuery, registration)

	return registration_id
}

func (app App) insertMetadata(metadata uct.Metadata) (metadata_id int64) {
	var insertQuery string
	var updateQuery string

	if metadata.UniversityId != nil {
		if !*fullUpsert {
			if metadata_id = app.dbHandler.exists(MetaUniExistQuery, metadata); metadata_id != 0 {
				return
			}
		}
		updateQuery = MetaUniUpdateQuery
		insertQuery = MetaUniInsertQuery

	} else if metadata.SubjectId != nil {
		if !*fullUpsert {
			if metadata_id = app.dbHandler.exists(MetaSubjectExistQuery, metadata); metadata_id != 0 {
				return
			}
		}
		updateQuery = MetaSubjectUpdateQuery
		insertQuery = MetaSubjectInsertQuery

	} else if metadata.CourseId != nil {
		if !*fullUpsert {
			if metadata_id = app.dbHandler.exists(MetaCourseExistQuery, metadata); metadata_id != 0 {
				return
			}
		}
		updateQuery = MetaCourseUpdateQuery
		insertQuery = MetaCourseInsertQuery

	} else if metadata.SectionId != nil {
		if !*fullUpsert {
			if metadata_id = app.dbHandler.exists(MetaSectionExistQuery, metadata); metadata_id != 0 {
				return
			}
		}
		updateQuery = MetaSectionUpdateQuery
		insertQuery = MetaSectionInsertQuery

	} else if metadata.MeetingId != nil {
		if !*fullUpsert {
			if metadata_id = app.dbHandler.exists(MetaMeetingExistQuery, metadata); metadata_id != 0 {
				return
			}
		}
		updateQuery = MetaMeetingUpdateQuery
		insertQuery = MetaMeetingInsertQuery
	}

	metadata_id = app.dbHandler.upsert(insertQuery, updateQuery, metadata)

	return metadata_id
}

type DatabaseHandler interface {
	insert(query string, data interface{}) (id int64)
	update(query string, data interface{}) (id int64)
	upsert(insertQuery, updateQuery string, data interface{}) (id int64)
	exists(query string, data interface{}) (id int64)
}

type DatabaseHandlerImpl struct {
	Database *sqlx.DB
}

func (dbHandler DatabaseHandlerImpl) insert(query string, data interface{}) (id int64) {
	insertionsCh <- 1
	typeName := reflect.TypeOf(data).Name()
	if rows, err := GetCachedStmt(query).Queryx(data); err != nil {
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

func (dbHandler DatabaseHandlerImpl) update(query string, data interface{}) (id int64) {
	updatesCh <- 1
	typeName := reflect.TypeOf(data).Name()

	if rows, err := GetCachedStmt(query).Queryx(data); err != nil {
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

func (dbHandler DatabaseHandlerImpl) upsert(insertQuery, updateQuery string, data interface{}) (id int64) {
	upsertsCh <- 1
	if id = dbHandler.update(updateQuery, data); id != 0 {
	} else if id == 0 {
		id = dbHandler.insert(insertQuery, data)
	}
	return
}

func (dbHandler DatabaseHandlerImpl) exists(query string, data interface{}) (id int64) {
	existentialCh <- 1

	if rows, err := GetCachedStmt(query).Queryx(data); err != nil {
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

func GetCachedStmt(key string) (stmt *sqlx.NamedStmt) {
	return preparedStmts[key]
}

func (dbHandler DatabaseHandlerImpl) prepare(query string) *sqlx.NamedStmt {
	if named, err := dbHandler.Database.PrepareNamed(query); err != nil {
		uct.CheckError(err)
		return nil
	} else {
		return named
	}
}

func (stats AuditStats) Log() {
	uct.Log("Insertions: ", stats.insertions)
	uct.Log("Updates: ", stats.updates)
	uct.Log("Upserts: ", stats.upserts)
	uct.Log("Existential: ", stats.existential)
	uct.Log("Subjects: ", stats.subjectCount)
	uct.Log("Courses: ", stats.courseCount)
	uct.Log("Sections: ", stats.sectionCount)
	uct.Log("Meetings: ", stats.metadataCount)
	uct.Log("Metadata: ", stats.metadataCount)
}

func (stats AuditStats) audit() {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "universityct",
		Precision: "s",
	})

	tags := map[string]string{
		"university_name": stats.uniName,
	}

	fields := map[string]interface{}{
		"insertions":    stats.insertions,
		"updates":       stats.updates,
		"upserts":       stats.upserts,
		"existential":   stats.existential,
		"subjectCount":  stats.subjectCount,
		"courseCount":   stats.courseCount,
		"sectionCount":  stats.sectionCount,
		"meetingCount":  stats.meetingCount,
		"metadataCount": stats.metadataCount,
		"elapsed":       stats.elapsed.Seconds(),
	}

	point, err := client.NewPoint(
		"db_ops",
		tags,
		fields,
		pointTime,
	)

	uct.CheckError(err)

	bp.AddPoint(point)

	err = stats.influxClient.Write(bp)
	uct.CheckError(err)

	log.Println("InfluxDB logging: ", tags, fields)
}

func audit() {
	influxClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     uct.INFLUX_HOST,
		Username: uct.INFLUX_USER,
		Password: uct.INFLUX_PASS,
	})
	uct.CheckError(err)

	var university string
	var insertions int
	var updates int
	var upserts int
	var existential int
	var subjectCount int
	var courseCount int
	var sectionCount int
	var meetingCount int
	var metadataCount int
	var startTime time.Time
	for {
		select {
		case <-startAudit:
			startTime = time.Now()
		case t1 := <-insertionsCh:
			insertions += t1
		case t2 := <-updatesCh:
			updates += t2
		case t3 := <-upsertsCh:
			upserts += t3
		case t4 := <-existentialCh:
			existential += t4
		case t5 := <-subjectCountCh:
			subjectCount += t5
		case t6 := <-courseCountCh:
			courseCount += t6
		case t7 := <-sectionCountCh:
			sectionCount += t7
		case t8 := <-meetingCountCh:
			meetingCount += t8
		case t9 := <-metadataCountCh:
			metadataCount += t9
		case <-endAudit:
			stats := AuditStats{influxClient, university, time.Since(startTime), insertions, updates,
				upserts, existential, subjectCount,
				courseCount, sectionCount, meetingCount, metadataCount}
			insertions, updates, upserts, existential, subjectCount, courseCount, sectionCount, meetingCount, metadataCount = 0, 0, 0, 0, 0, 0, 0, 0, 0

			stats.Log()
		}
	}
}

var preparedStmts = make(map[string]*sqlx.NamedStmt)

func (dbHandler DatabaseHandlerImpl) PrepareAllStmts() {
	queries := []string{UniversityInsertQuery,
		UniversityUpdateQuery,
		SubjectExistQuery,
		SubjectInsertQuery,
		SubjectUpdateQuery,
		CourseUpdateQuery,
		CourseExistQuery,
		CourseInsertQuery,
		SectionInsertQuery,
		SectionUpdateQuery,
		MeetingUpdateQuery,
		MeetingInsertQuery,
		MeetingExistQuery,
		InstructorExistQuery,
		InstructorInsertQuery,
		BookUpdateQuery,
		BookInsertQuery,
		RegistrationUpdateQuery,
		RegistrationInsertQuery,
		MetaUniExistQuery,
		MetaUniUpdateQuery,
		MetaUniInsertQuery,
		MetaSubjectExistQuery,
		MetaSubjectUpdateQuery,
		MetaSubjectInsertQuery,
		MetaCourseExistQuery,
		MetaCourseUpdateQuery,
		MetaCourseInsertQuery,
		MetaSectionExistQuery,
		MetaSectionInsertQuery,
		MetaSectionUpdateQuery,
		MetaSectionExistQuery,
		MetaMeetingInsertQuery,
		MetaMeetingUpdateQuery}

	for _, query := range queries {
		preparedStmts[query] = dbHandler.prepare(query)
	}
}

var (
	UniversityInsertQuery = `INSERT INTO university (name, abbr, home_page, registration_page, main_color, accent_color, topic_name)
                    VALUES (:name, :abbr, :home_page, :registration_page, :main_color, :accent_color, :topic_name)
                    RETURNING university.id`
	UniversityUpdateQuery = `UPDATE university SET (abbr, home_page, registration_page, main_color, accent_color, topic_name) =
	                (:abbr, :home_page, :registration_page, :main_color, :accent_color, :topic_name)
	                WHERE name = :name
	                RETURNING university.id`

	SubjectExistQuery = `SELECT id FROM course
						WHERE hash = :hash`
	SubjectInsertQuery = `INSERT INTO subject (university_id, name, number, season, year, hash, topic_name)
                   	VALUES  (:university_id, :name, :number, :season, :year, :hash, :topic_name)
                   	RETURNING subject.id`
	SubjectUpdateQuery = `UPDATE subject SET (name, season, year, topic_name) = (:name, :season, :year, :topic_name)
					WHERE hash = :hash
                   	RETURNING subject.id`

	CourseExistQuery = `SELECT id FROM course
						WHERE hash = :hash`

	CourseUpdateQuery = `UPDATE course SET (name, synopsis, topic_name) = (:name, :synopsis, :topic_name)
					WHERE hash = :hash
                    RETURNING course.id`

	CourseInsertQuery = `INSERT INTO course (subject_id, name, number, synopsis, hash, topic_name)
                    VALUES  (:subject_id, :name, :number, :synopsis, :hash, :topic_name)
                    RETURNING course.id`

	SectionInsertQuery = `INSERT INTO section (course_id, number, call_number, max, now, status, credits, topic_name)
                    VALUES  (:course_id, :number, :call_number, :max, :now, :status, :credits, :topic_name)
                    RETURNING section.id`

	SectionUpdateQuery = `UPDATE section SET (max, now, status, credits, topic_name) = (:max, :now, :status, :credits, :topic_name)
					WHERE course_id = :course_id AND call_number = :call_number AND number = :number
                    RETURNING section.id`

	MeetingExistQuery = `SELECT id FROM meeting
						WHERE section_id = :section_id AND index = :index`

	MeetingUpdateQuery = `UPDATE meeting SET (room, day, start_time, end_time) = (:room, :day, :start_time, :end_time)
					WHERE section_id = :section_id AND index = :index
                    RETURNING meeting.id`

	MeetingInsertQuery = `INSERT INTO meeting (section_id, room, day, start_time, end_time, index)
                    VALUES  (:section_id, :room, :day, :start_time, :end_time, :index)
                    RETURNING meeting.id`

	InstructorExistQuery = `SELECT id FROM instructor
				WHERE section_id = :section_id AND name = :name`

	InstructorInsertQuery = `INSERT INTO instructor (section_id, name)
				VALUES  (:section_id, :name)
				RETURNING instructor.id`

	BookUpdateQuery = `UPDATE book SET (title, url) = (:title, :url)
					WHERE section_id = :section_id AND title = :title
                    RETURNING book.id`

	BookInsertQuery = `INSERT INTO book (section_id, title, url)
                    VALUES  (:section_id, :title, :url)
                    RETURNING book.id`

	RegistrationUpdateQuery = `UPDATE registration SET (period_date) = (:period_date)
					WHERE university_id = :university_id AND period = :period
	                RETURNING registration.id`

	RegistrationInsertQuery = `INSERT INTO registration (university_id, period, period_date)
                    VALUES (:university_id, :period, :period_date)
                    RETURNING registration.id`

	MetaUniExistQuery = `SELECT id FROM metadata
						WHERE metadata.university_id = :university_id AND metadata.title = :title`
	MetaUniUpdateQuery = `UPDATE metadata SET (title, content) = (:title, :content)
					   WHERE metadata.university_id = :university_id AND metadata.title = :title
		               RETURNING metadata.id`
	MetaUniInsertQuery = `INSERT INTO metadata (university_id, title, content)
                       VALUES (:university_id, :title, :content)
                       RETURNING metadata.id`

	MetaSubjectExistQuery = `SELECT id FROM metadata
						WHERE metadata.subject_id = :subject_id AND metadata.title = :title`
	MetaSubjectUpdateQuery = `UPDATE metadata SET (title, content) = (:title, :content)
					   WHERE metadata.subject_id = :subject_id AND metadata.title = :title
		               RETURNING metadata.id`
	MetaSubjectInsertQuery = `INSERT INTO metadata (subject_id, title, content)
                       VALUES (:subject_id, :title, :content)
                       RETURNING metadata.id`

	MetaCourseExistQuery = `SELECT id FROM metadata
						WHERE metadata.course_id = :course_id AND metadata.title = :title`

	MetaCourseUpdateQuery = `UPDATE metadata SET (title, content) = (:title, :content)
					   WHERE metadata.course_id = :course_id AND metadata.title = :title
		               RETURNING metadata.id`

	MetaCourseInsertQuery = `INSERT INTO metadata (course_id, title, content)
                       VALUES (:course_id, :title, :content)
                       RETURNING metadata.id`

	MetaSectionExistQuery = `SELECT id FROM metadata
						WHERE metadata.section_id = :section_id AND metadata.title = :title`
	MetaSectionInsertQuery = `INSERT INTO metadata (section_id, title, content)
                       VALUES (:section_id, :title, :content)
                       RETURNING metadata.id`
	MetaSectionUpdateQuery = `UPDATE metadata SET (title, content) = (:title, :content)
					   WHERE metadata.section_id = :section_id AND metadata.title = :title
		               RETURNING metadata.id`

	MetaMeetingExistQuery = `SELECT id FROM metadata
						WHERE metadata.meeting_id = :meeting_id AND metadata.title = :title`
	MetaMeetingInsertQuery = `INSERT INTO metadata (meeting_id, title, content)
                       VALUES (:meeting_id, :title, :content)
                       RETURNING metadata.id`
	MetaMeetingUpdateQuery = `UPDATE metadata SET (title, content) = (:title, :content)
					   WHERE metadata.meeting_id = :meeting_id AND metadata.title = :title
		               RETURNING metadata.id`
)

type ChannelSubjects struct {
	subjectId int64
	courses   []uct.Course
}

type AuditStats struct {
	influxClient  client.Client
	uniName       string
	elapsed       time.Duration
	insertions    int
	updates       int
	upserts       int
	existential   int
	subjectCount  int
	courseCount   int
	sectionCount  int
	meetingCount  int
	metadataCount int
}

var (
	insertionsCh    = make(chan int)
	updatesCh       = make(chan int)
	upsertsCh       = make(chan int)
	existentialCh   = make(chan int)
	subjectCountCh  = make(chan int)
	courseCountCh   = make(chan int)
	sectionCountCh  = make(chan int)
	meetingCountCh  = make(chan int)
	metadataCountCh = make(chan int)
	endAudit        = make(chan bool)
	startAudit      = make(chan bool)
	pointTime       = time.Now()
)
