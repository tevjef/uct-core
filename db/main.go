package main

import (
	"bufio"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gopkg.in/alecthomas/kingpin.v2"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"
	uct "uct/common"
)

type App struct {
	dbHandler DatabaseHandler
}

type serial struct {
	TopicName string `db:"topic_name"`
	Data      []byte `db:"data"`
}

type serialSubject struct {
	serial
}

type serialCourse struct {
	serial
}

type serialSection struct {
	serial
}

var (
	app        = kingpin.New("uct-db", "A command-line application for inserting and updated university information")
	fullUpsert = app.Flag("insert-all", "full insert/update of all objects.").Default("true").Short('a').Bool()
	oldFile    = app.Flag("diff", "file to read university data from.").Short('d').File()
	newFile    = app.Flag("input", "file to read university data from.").Short('i').File()
	format     = app.Flag("format", "choose input format").Short('f').HintOptions(uct.JSON, uct.PROTOBUF).PlaceHolder("[protobuf, json]").Required().String()
	server     = app.Flag("pprof", "host:port to start profiling on").Short('p').Default(uct.DB_DEBUG_SERVER).TCP()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *format != uct.JSON && *format != uct.PROTOBUF {
		log.Fatalln("Invalid format:", *format)
	}

	go uct.StartPprof(*server)

	var input *bufio.Reader
	if *newFile != nil {
		input = bufio.NewReader(*newFile)
	} else {
		input = bufio.NewReader(os.Stdin)
	}

	var university uct.University
	newUniversity := new(uct.University)
	uct.UnmarshallMessage(*format, input, newUniversity)

	// Make sure the data received is primed for the database
	uct.ValidateAll(newUniversity)

	oldUniversity := new(uct.University)

	// If an old version was supplied diff the old and new to create a new university
	if *oldFile != nil {
		old := bufio.NewReader(*oldFile)
		if err := uct.UnmarshallMessage(*format, old, oldUniversity); err != nil {
			log.Debug(err)
		}
		uct.ValidateAll(oldUniversity)
		university = uct.DiffAndFilter(*oldUniversity, *newUniversity)
	} else {
		university = *newUniversity
	}

	// Initialize database connection
	database, err := uct.InitDB(uct.GetUniversityDB())
	uct.CheckError(err)

	dbHandler := DatabaseHandlerImpl{Database: database}
	dbHandler.PrepareAllStmts()
	app := App{dbHandler: dbHandler}

	// Start logging with influx
	go audit(university.TopicName)

	app.insertUniversity(&university)
	app.updateSerial(*newUniversity)

	endAudit <- true
	<- doneAudit
	// Before main ends, close the database and stop writing profile
	defer func() {
		database.Close()
	}()

}

func (app App) updateSerial(uni uct.University) {
	for subjectIndex := range uni.Subjects {
		subject := uni.Subjects[subjectIndex]

		app.updateSerialSubject(subject)

		for courseIndex := range subject.Courses {
			course := subject.Courses[courseIndex]

			app.updateSerialCourse(course)

			for sectionIndex := range course.Sections {
				section := course.Sections[sectionIndex]

				app.updateSerialSection(section)
			}
		}
	}
}

func (app App) updateSerialSubject(subject *uct.Subject) {
	serialSubjectCh <- 1
	data, err := subject.Marshal()
	uct.CheckError(err)
	arg := serialSubject{serial{TopicName: subject.TopicName, Data: data}}
	app.dbHandler.update(SerialSubjectUpdateQuery, arg)

	// Sanity Check
	log.WithFields(log.Fields{"subject": subject.TopicId, "bytes": len(data)}).Debugln("sanity")
}

func (app App) updateSerialCourse(course *uct.Course) {
	serialCourseCh <- 1
	data, err := course.Marshal()
	uct.CheckError(err)
	arg := serialCourse{serial{TopicName: course.TopicName, Data: data}}
	app.dbHandler.update(SerialCourseUpdateQuery, arg)

	// Sanity Check
	log.WithFields(log.Fields{"course": course.TopicId, "bytes": len(data)}).Debugln("sanity")
}

func (app App) updateSerialSection(section *uct.Section) {
	serialSectionCh <- 1
	data, err := section.Marshal()
	uct.CheckError(err)
	arg := serialSection{serial{TopicName: section.TopicName, Data: data}}
	app.dbHandler.update(SerialSectionUpdateQuery, arg)

	// Sanity Check
	log.WithFields(log.Fields{"section": section.TopicId, "bytes": len(data)}).Debugln("sanity")
}

func (app App) insertUniversity(uni *uct.University) {
	universityId := app.dbHandler.upsert(UniversityInsertQuery, UniversityUpdateQuery, uni)

	subjectCountCh <- len(uni.Subjects)
	for subjectIndex := range uni.Subjects {
		subject := uni.Subjects[subjectIndex]
		subject.UniversityId = universityId

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
				// Make section data available as soon as possible
				app.updateSerialSection(section)

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

	// ResolvedSemesters
	app.insertSemester(universityId, uni.ResolvedSemesters)

	// Registrations
	for _, registrations := range uni.Registrations {
		registrations.UniversityId = universityId
		app.insertRegistration(registrations)
	}

	// university []Metadata
	metadatas := uni.Metadata
	for metadataIndex := range metadatas {
		metadata := metadatas[metadataIndex]

		metadata.UniversityId = &universityId
		app.insertMetadata(metadata)
	}
}

func (app App) insertSubject(sub *uct.Subject) (subjectId int64) {
	if !*fullUpsert {

		if subjectId = app.dbHandler.exists(SubjectExistQuery, sub); subjectId != 0 {
			return
		}
	}
	subjectId = app.dbHandler.upsert(SubjectInsertQuery, SubjectUpdateQuery, sub)

	// Subject []Metadata
	metadatas := sub.Metadata
	for metadataIndex := range metadatas {
		metadata := metadatas[metadataIndex]

		metadata.SubjectId = &subjectId
		app.insertMetadata(metadata)
	}
	return subjectId
}

func (app App) insertCourse(course *uct.Course) (courseId int64) {
	if !*fullUpsert {

		if courseId = app.dbHandler.exists(CourseExistQuery, course); courseId != 0 {
			return
		}
	}
	courseId = app.dbHandler.upsert(CourseInsertQuery, CourseUpdateQuery, course)

	return courseId
}

func (app App) insertSemester(universityId int64, resolvedSemesters *uct.ResolvedSemester) int64 {
	rs := &uct.DBResolvedSemester{}
	rs.UniversityId = universityId
	rs.CurrentSeason = resolvedSemesters.Current.Season
	rs.CurrentYear = strconv.Itoa(int(resolvedSemesters.Current.Year))
	rs.LastSeason = resolvedSemesters.Last.Season
	rs.LastYear = strconv.Itoa(int(resolvedSemesters.Last.Year))
	rs.NextSeason = resolvedSemesters.Next.Season
	rs.NextYear = strconv.Itoa(int(resolvedSemesters.Next.Year))
	return app.dbHandler.upsert(SemesterInsertQuery, SemesterUpdateQuery, rs)
}

func (app App) insertSection(section *uct.Section) int64 {
	return app.dbHandler.upsert(SectionInsertQuery, SectionUpdateQuery, section)
}

func (app App) insertMeeting(meeting *uct.Meeting) (meetingId int64) {
	if !*fullUpsert {
		if meetingId = app.dbHandler.exists(MeetingExistQuery, meeting); meetingId != 0 {
			return
		}
	}
	return app.dbHandler.upsert(MeetingInsertQuery, MeetingUpdateQuery, meeting)
}

func (app App) insertInstructor(instructor *uct.Instructor) (instructorId int64) {
	if instructorId = app.dbHandler.exists(InstructorExistQuery, instructor); instructorId != 0 {
		return
	}
	return app.dbHandler.upsert(InstructorInsertQuery, InstructorUpdateQuery, instructor)
}

func (app App) insertBook(book *uct.Book) (bookId int64) {
	bookId = app.dbHandler.upsert(BookInsertQuery, BookUpdateQuery, book)

	return bookId
}

func (app App) insertRegistration(registration *uct.Registration) int64 {
	return app.dbHandler.upsert(RegistrationInsertQuery, RegistrationUpdateQuery, registration)
}

func (app App) insertMetadata(metadata *uct.Metadata) (metadataId int64) {
	var insertQuery string
	var updateQuery string

	if metadata.UniversityId != nil {
		if !*fullUpsert {
			if metadataId = app.dbHandler.exists(MetaUniExistQuery, metadata); metadataId != 0 {
				return
			}
		}
		updateQuery = MetaUniUpdateQuery
		insertQuery = MetaUniInsertQuery

	} else if metadata.SubjectId != nil {
		if !*fullUpsert {
			if metadataId = app.dbHandler.exists(MetaSubjectExistQuery, metadata); metadataId != 0 {
				return
			}
		}
		updateQuery = MetaSubjectUpdateQuery
		insertQuery = MetaSubjectInsertQuery

	} else if metadata.CourseId != nil {
		if !*fullUpsert {
			if metadataId = app.dbHandler.exists(MetaCourseExistQuery, metadata); metadataId != 0 {
				return
			}
		}
		updateQuery = MetaCourseUpdateQuery
		insertQuery = MetaCourseInsertQuery

	} else if metadata.SectionId != nil {
		if !*fullUpsert {
			if metadataId = app.dbHandler.exists(MetaSectionExistQuery, metadata); metadataId != 0 {
				return
			}
		}
		updateQuery = MetaSectionUpdateQuery
		insertQuery = MetaSectionInsertQuery

	} else if metadata.MeetingId != nil {
		if !*fullUpsert {
			if metadataId = app.dbHandler.exists(MetaMeetingExistQuery, metadata); metadataId != 0 {
				return
			}
		}
		updateQuery = MetaMeetingUpdateQuery
		insertQuery = MetaMeetingInsertQuery
	}
	return app.dbHandler.upsert(insertQuery, updateQuery, metadata)
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
	typeName := fmt.Sprintf("%T", data)
	if rows, err := GetCachedStmt(query).Queryx(data); err != nil {
		log.WithFields(log.Fields{"db_op": "Insert", "type": typeName, "data": data}).Panic(err)
	} else {
		for rows.Next() {
			if err = rows.Scan(&id); err != nil {
				log.WithFields(log.Fields{"db_op": "Insert", "type": typeName, "data": data}).Panic(err)
			}
			rows.Close()
			log.WithFields(log.Fields{"db_op": "Insert", "type": typeName, "id": id}).Info()
		}
	}
	return id
}

func (dbHandler DatabaseHandlerImpl) update(query string, data interface{}) (id int64) {
	typeName := fmt.Sprintf("%T", data)
	updatesCh <- 1

	if rows, err := GetCachedStmt(query).Queryx(data); err != nil {
		log.WithFields(log.Fields{"db_op": "Update", "type": typeName, "data": data}).Panic(err)
	} else {
		count := 0
		for rows.Next() {
			count++
			if err = rows.Scan(&id); err != nil {
				log.WithFields(log.Fields{"db_op": "Update", "type": typeName, "data": data}).Panic(err)
			}
			rows.Close()
			log.WithFields(log.Fields{"db_op": "Update", "type": typeName, "id": id}).Info()
		}
		if count > 1 {
			log.WithFields(log.Fields{"db_op": "Update", "type": typeName, "data": data}).Panic("Multiple rows updated at once")
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
	typeName := fmt.Sprintf("%T", data)
	existentialCh <- 1

	if rows, err := GetCachedStmt(query).Queryx(data); err != nil {
		log.WithFields(log.Fields{"db_op": "Exists", "type": typeName, "data": data}).Panic(err)
	} else {
		count := 0
		for rows.Next() {
			count++
			if err = rows.Scan(&id); err != nil {
				log.WithFields(log.Fields{"db_op": "Exists", "type": typeName, "data": data}).Panic(err)
			}
			log.WithFields(log.Fields{"db_op": "Exists", "type": typeName, "id": id}).Info()
		}
		if count > 1 {
			log.WithFields(log.Fields{"db_op": "Exists", "type": typeName, "data": data}).Panic("Multple rows exists")
		}
	}

	return
}

func GetCachedStmt(key string) *sqlx.NamedStmt {
	return preparedStmts[key]
}

func (dbHandler DatabaseHandlerImpl) prepare(query string) *sqlx.NamedStmt {
	if named, err := dbHandler.Database.PrepareNamed(query); err != nil {
		log.Debugln(query)
		uct.CheckError(err)
		return nil
	} else {
		return named
	}
}

func (stats AuditStats) log() {
	log.WithFields(log.Fields{
		"Insertions":       stats.insertions,
		"Updates":          stats.updates,
		"Upserts":          stats.upserts,
		"Existential":      stats.existential,
		"Subjects":         stats.subjectCount,
		"Courses":          stats.courseCount,
		"Sections":         stats.sectionCount,
		"Meetings":         stats.meetingCount,
		"Metadata":         stats.metadataCount,
		"SerialSubject":    stats.serialSubject,
		"SerialCourse":     stats.serialCourse,
		"SerialSection":    stats.serialSection,
	}).Info("DB operations complete")
}

func (stats AuditStats) audit() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from influx error", r)
		}
	}()

	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "universityct",
		Precision: "s",
	})

	tags := map[string]string{
		"university_name": stats.uniName,
	}

	fields := map[string]interface{}{
		"insertions":       stats.insertions,
		"updates":          stats.updates,
		"upserts":          stats.upserts,
		"existential":      stats.existential,
		"subjectCount":     stats.subjectCount,
		"courseCount":      stats.courseCount,
		"sectionCount":     stats.sectionCount,
		"meetingCount":     stats.meetingCount,
		"metadataCount":    stats.metadataCount,
		"serialSubject":    stats.serialSubject,
		"serialCourse":     stats.serialCourse,
		"serialSection":    stats.serialSection,
		"elapsed":          stats.elapsed.Seconds(),
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
}

func audit(university string) {
	influxClient, _ := client.NewHTTPClient(client.HTTPConfig{
		Addr:     uct.INFLUX_HOST,
		Username: uct.INFLUX_USER,
		Password: uct.INFLUX_PASS,
	})

	var insertions int
	var updates int
	var upserts int
	var existential int
	var subjectCount int
	var courseCount int
	var sectionCount int
	var meetingCount int
	var metadataCount int
	var serialCourse int
	var serialSection int
	var serialSubject int

	for {
		select {
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
		case t11 := <-serialSubjectCh:
			serialSubject += t11
		case t12 := <-serialCourseCh:
			serialCourse += t12
		case t13 := <-serialSectionCh:
			serialSection += t13
		case <-endAudit:
			stats := AuditStats{
				influxClient:     influxClient,
				uniName:          university,
				elapsed:          time.Since(pointTime),
				insertions:       insertions,
				updates:          updates,
				upserts:          upserts,
				existential:      existential,
				subjectCount:     subjectCount,
				courseCount:      courseCount,
				sectionCount:     sectionCount,
				meetingCount:     meetingCount,
				metadataCount:    metadataCount,
				serialSubject:    serialSubject,
				serialCourse:     serialCourse,
				serialSection:    serialSection}

			stats.log()
			stats.audit()
			doneAudit <- true
		}
	}
}

var preparedStmts = make(map[string]*sqlx.NamedStmt)

func (dbHandler DatabaseHandlerImpl) PrepareAllStmts() {
	queries := []string{UniversityInsertQuery,
		UniversityUpdateQuery,
		SemesterInsertQuery,
		SemesterUpdateQuery,
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
		InstructorUpdateQuery,
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
		MetaMeetingUpdateQuery,
		SerialSubjectUpdateQuery,
		SerialCourseUpdateQuery,
		SerialSectionUpdateQuery}

	for _, query := range queries {
		preparedStmts[query] = dbHandler.prepare(query)
	}
}

var (
	UniversityInsertQuery = `INSERT INTO university (name, abbr, home_page, registration_page, main_color, accent_color, topic_name, topic_id)
                    VALUES (:name, :abbr, :home_page, :registration_page, :main_color, :accent_color, :topic_name, :topic_id)
                    RETURNING university.id`
	UniversityUpdateQuery = `UPDATE university SET (abbr, home_page, registration_page, main_color, accent_color, topic_name, topic_id) =
	                (:abbr, :home_page, :registration_page, :main_color, :accent_color, :topic_name, :topic_id)
	                WHERE name = :name
	                RETURNING university.id`

	SemesterInsertQuery = `INSERT INTO semester (university_id, current_season, current_year, last_season, last_year, next_season, next_year)
							VALUES (:university_id, :current_season, :current_year, :last_season, :last_year, :next_season, :next_year) RETURNING semester.id`
	SemesterUpdateQuery = `UPDATE semester SET (current_season, current_year, last_season, last_year, next_season, next_year) =
						(:current_season, :current_year, :last_season, :last_year, :next_season, :next_year) WHERE university_id = :university_id RETURNING semester.id`

	SubjectExistQuery = `SELECT subject.id FROM subject WHERE topic_name = :topic_name`

	SubjectInsertQuery = `INSERT INTO subject (university_id, name, number, season, year, topic_name, topic_id)
                   	VALUES  (:university_id, :name, :number, :season, :year, :topic_name, :topic_id)
                   	RETURNING subject.id`

	SubjectUpdateQuery = SubjectExistQuery

	CourseInsertQuery = `INSERT INTO course (subject_id, name, number, synopsis, topic_name, topic_id) VALUES  (:subject_id, :name, :number, :synopsis, :topic_name, :topic_id) RETURNING course.id`
	CourseExistQuery  = `SELECT course.id FROM course WHERE topic_name = :topic_name`
	CourseUpdateQuery = `UPDATE course SET (synopsis) = (:synopsis) WHERE topic_name = :topic_name RETURNING course.id`

	SectionInsertQuery = `INSERT INTO section (course_id, number, call_number, max, now, status, credits, topic_name, topic_id)
                    VALUES (:course_id, :number, :call_number, :max, :now, :status, :credits, :topic_name, :topic_id)
                    RETURNING section.id`

	SectionUpdateQuery = `UPDATE section SET (max, now, status, credits) = (:max, :now, :status, :credits) WHERE topic_name = :topic_name RETURNING section.id`

	MeetingExistQuery = `SELECT id FROM meeting WHERE section_id = :section_id AND index = :index`

	MeetingUpdateQuery = `UPDATE meeting SET (room, day, start_time, end_time, class_type) = (:room, :day, :start_time, :end_time, :class_type)
					WHERE section_id = :section_id AND index = :index
                    RETURNING meeting.id`

	MeetingInsertQuery = `INSERT INTO meeting (section_id, room, day, start_time, end_time, class_type, index)
                    VALUES  (:section_id, :room, :day, :start_time, :end_time, :class_type, :index)
                    RETURNING meeting.id`

	InstructorExistQuery = `SELECT id FROM instructor
				WHERE section_id = :section_id AND index = :index`

	InstructorUpdateQuery = `UPDATE instructor SET name = :name WHERE section_id = :section_id AND index = :index
				RETURNING instructor.id`

	InstructorInsertQuery = `INSERT INTO instructor (section_id, name, index)
				VALUES  (:section_id, :name, :index)
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

	SerialSubjectUpdateQuery = `UPDATE subject SET data = :data WHERE topic_name = :topic_name RETURNING subject.id`
	SerialCourseUpdateQuery  = `UPDATE course SET data = :data WHERE topic_name = :topic_name RETURNING course.id`
	SerialSectionUpdateQuery = `UPDATE section SET data = :data WHERE topic_name = :topic_name RETURNING section.id`
)

type ChannelSubjects struct {
	subjectId int64
	courses   []uct.Course
}

type AuditStats struct {
	influxClient     client.Client
	uniName          string
	elapsed          time.Duration
	insertions       int
	updates          int
	upserts          int
	existential      int
	subjectCount     int
	courseCount      int
	sectionCount     int
	meetingCount     int
	metadataCount    int
	serialCourse     int
	serialSection    int
	serialSubject    int
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

	serialCourseCh     = make(chan int)
	serialSectionCh    = make(chan int)
	serialSubjectCh    = make(chan int)

	endAudit  = make(chan bool)
	doneAudit  = make(chan bool)

	pointTime = time.Now()
)
