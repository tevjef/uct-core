package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/vlad-doru/influxus"
	"gopkg.in/alecthomas/kingpin.v2"
	"net"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	uct "uct/common"
	"uct/redis"
	"uct/influxdb"
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
	app        = kingpin.New("ein", "A command-line application for inserting and updated university information")
	noDiff = app.Flag("no-diff", "do not diff against last data").Default("false").Bool()
	fullUpsert = app.Flag("insert-all", "full insert/update of all objects.").Default("true").Short('a').Bool()
	format     = app.Flag("format", "choose input format").Short('f').HintOptions(uct.JSON, uct.PROTOBUF).PlaceHolder("[protobuf, json]").Required().String()
	configFile = app.Flag("config", "configuration file for the application").Short('c').File()
	config     = uct.Config{}

	mutiProgramming = 5
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *format != uct.JSON && *format != uct.PROTOBUF {
		log.Fatalln("Invalid format:", *format)
	}

	log.SetLevel(log.InfoLevel)

	// Parse configuration file
	config = uct.NewConfig(*configFile)
	config.AppName = app.Name

	// Start profiling
	go uct.StartPprof(config.GetDebugSever(app.Name))

	// Start redis client
	wrapper := v1.New(config, app.Name)

	// Start influx logging
	initInflux()

	// Initialize database connection
	database, err := uct.InitDB(config.GetDbConfig(app.Name))
	uct.CheckError(err)

	dbHandler := DatabaseHandlerImpl{Database: database}
	dbHandler.PrepareAllStmts()
	app := App{dbHandler: dbHandler}
	database.SetMaxOpenConns(mutiProgramming)

	for {
		log.Info("Waiting on queue...")
		if data, err := wrapper.Client.BRPop(5*time.Minute, v1.BaseNamespace+":queue").Result(); err != nil {
			uct.CheckError(err)
		} else {
			func () {
				defer func() {
					if r := recover(); r != nil {
						 log.WithError(fmt.Errorf("Recovered from error in queue loop: %v", r)).Error()
					}
				}()

				val := data[1]

				latestData := val + ":data:latest"
				oldData := val + ":data:old"

				log.WithFields(log.Fields{"key":val}).Infoln("RPOP")

				if raw, err := wrapper.Client.Get(latestData).Result(); err != nil {
					log.WithError(err).Panic("Error getting latest data")
				} else {
					var university uct.University

					// Try getting older data from redis
					var oldRaw string
					if oldRaw, err = wrapper.Client.Get(oldData).Result(); err != nil {
						log.Warningln("There was no older data, did it expire or is this first run?")
					}

					// Set old data as the new data we just recieved
					if _, err := wrapper.Client.Set(oldData, raw, 0).Result(); err != nil {
						log.WithError(err).Panic("Error updating old data")
					}

					// Decode new data
					newUniversity := new(uct.University)
					if err := uct.UnmarshallMessage(*format, strings.NewReader(raw), newUniversity); err != nil {
						log.WithError(err).Panic("Error while unmarshalling new data")
					}

					// Make sure the data received is primed for the database
					if err := uct.ValidateAll(newUniversity); err != nil {
						log.WithError(err).Panic("Error while validating newUniversity")
					}

					// Decode old data if have some
					if oldRaw != "" && !*noDiff {
						oldUniversity := new(uct.University)
						if err := uct.UnmarshallMessage(*format, strings.NewReader(oldRaw), oldUniversity); err != nil {
							log.WithError(err).Panic("Error while unmarshalling old data")
						}

						if err := uct.ValidateAll(oldUniversity); err != nil {
							log.WithError(err).Panic("Error while validating oldUniversity")
						}

						university = uct.DiffAndFilter(*oldUniversity, *newUniversity)
					} else {
						university = *newUniversity
					}

					// Start logging with influx
					go audit(university.TopicName)

					app.insertUniversity(&university)
					//app.insertUniversity(newUniversity)
					app.updateSerial(*newUniversity)

					// Log bytes received
					auditLogger.WithFields(log.Fields{"bytes": len([]byte(raw)), "university_name":university.TopicName}).Info(latestData)

					doneAudit <- true
					<-doneAudit
					//break
				}

			}()

		}
	}
}

func (app App) updateSerial(uni uct.University) {
	defer uct.TimeTrack(time.Now(), "updateSerial")

	sem := make(chan bool, mutiProgramming)
	for subjectIndex := range uni.Subjects {
		subject := uni.Subjects[subjectIndex]

		app.updateSerialSubject(subject)

		cwg := sync.WaitGroup{}
		for courseIndex := range subject.Courses {
			course := subject.Courses[courseIndex]
			cwg.Add(1)

			sem <- true
			go func() {
				app.updateSerialCourse(course)

				for sectionIndex := range course.Sections {
					section := course.Sections[sectionIndex]
					app.updateSerialSection(section)
				}

				<-sem
				cwg.Done()
			}()
		}
		cwg.Wait()
	}

}

func (app App) updateSerialSubject(subject *uct.Subject) {
	serialSubjectCh <- 1
	data, err := subject.Marshal()
	uct.CheckError(err)
	arg := serialSubject{serial{TopicName: subject.TopicName, Data: data}}
	app.dbHandler.update(SerialSubjectUpdateQuery, arg)

	// Sanity Check
	// log.WithFields(log.Fields{"subject": subject.TopicId, "bytes": len(data)}).Debugln("sanity")
}

func (app App) updateSerialCourse(course *uct.Course) {
	serialCourseCh <- 1
	data, err := course.Marshal()
	uct.CheckError(err)
	arg := serialCourse{serial{TopicName: course.TopicName, Data: data}}
	app.dbHandler.update(SerialCourseUpdateQuery, arg)

	// Sanity Check
	// log.WithFields(log.Fields{"course": course.TopicId, "bytes": len(data)}).Debugln("sanity")
}

func (app App) updateSerialSection(section *uct.Section) {
	serialSectionCh <- 1
	data, err := section.Marshal()
	uct.CheckError(err)
	arg := serialSection{serial{TopicName: section.TopicName, Data: data}}
	app.dbHandler.update(SerialSectionUpdateQuery, arg)

	// Sanity Check
	// log.WithFields(log.Fields{"section": section.TopicId, "bytes": len(data)}).Debugln("sanity")
}

func (app App) insertUniversity(uni *uct.University) {
	defer uct.TimeTrack(time.Now(), "insertUniversity")

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
	// uct.TimeTrack(time.Now(), "insert")

	insertionsCh <- 1
	typeName := fmt.Sprintf("%T", data)
	if rows, err := GetCachedStmt(query).Queryx(data); err != nil {
		log.WithFields(log.Fields{"ein_op": "Insert", "type": typeName, "data": data}).Panic(err)
	} else {
		for rows.Next() {
			if err = rows.Scan(&id); err != nil {
				log.WithFields(log.Fields{"ein_op": "Insert", "type": typeName, "data": data}).Panic(err)
			}
			rows.Close()
			log.WithFields(log.Fields{"ein_op": "Insert", "type": typeName, "id": id}).Debug()
		}
	}
	return id
}

func (dbHandler DatabaseHandlerImpl) update(query string, data interface{}) (id int64) {
	// uct.TimeTrack(time.Now(), "update")
	typeName := fmt.Sprintf("%T", data)

	for i := 0; i < 5; i++ {
		if rows, err := GetCachedStmt(query).Queryx(data); err != nil {
			if isConnectionError(err) {
				log.Errorf("Retry %d after error %s", i, err)
				continue
			} else {
				log.Panicln(err)
			}
		} else {
			count := 0
			for rows.Next() {
				count++

				if err = rows.Scan(&id); err != nil {
					log.WithFields(log.Fields{"ein_op": "Update", "type": typeName, "data": data}).Panic(err)
				}
				rows.Close()
				log.WithFields(log.Fields{"ein_op": "Update", "type": typeName, "id": id}).Debug()
			}
			if count > 1 {
				log.WithFields(log.Fields{"ein_op": "Update", "type": typeName, "data": data}).Panic("Multiple rows updated at once")
			}

			break
		}
	}

	updatesCh <- 1
	return id
}

func (dbHandler DatabaseHandlerImpl) upsert(insertQuery, updateQuery string, data interface{}) (id int64) {
	// uct.TimeTrack(time.Now(), "upsert")
	upsertsCh <- 1
	if id = dbHandler.update(updateQuery, data); id != 0 {
	} else if id == 0 {
		id = dbHandler.insert(insertQuery, data)
	}
	return
}

func isConnectionError(err error) bool {
	if operr, ok := err.(*net.OpError); ok {
		if syserr, ok := operr.Err.(*os.SyscallError); ok {
			if errNo := syserr.Err; errNo == syscall.ECONNRESET || errNo == syscall.EPIPE || errNo == syscall.EPROTOTYPE {
				return true
			}
		}
	}
	return false
}

func (dbHandler DatabaseHandlerImpl) exists(query string, data interface{}) (id int64) {
	typeName := fmt.Sprintf("%T", data)
	existentialCh <- 1

	if rows, err := GetCachedStmt(query).Queryx(data); err != nil {
		log.WithFields(log.Fields{"ein_op": "Exists", "type": typeName, "data": data}).Panic(err)
	} else {
		count := 0
		for rows.Next() {
			count++
			if err = rows.Scan(&id); err != nil {
				log.WithFields(log.Fields{"ein_op": "Exists", "type": typeName, "data": data}).Panic(err)
			}
			log.WithFields(log.Fields{"ein_op": "Exists", "type": typeName, "id": id}).Debug()
		}
		if count > 1 {
			log.WithFields(log.Fields{"ein_op": "Exists", "type": typeName, "data": data}).Panic("Multple rows exists")
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

func audit(university string) {
	var err error

	start := time.Now()

	if err != nil {
		log.Fatalf("Error while creating the hook: %v", err)
	}

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

	Outerloop:
	for {
		select {
		case count := <-insertionsCh:
			insertions += count
		case count := <-updatesCh:
			updates += count
		case count := <-upsertsCh:
			upserts += count
		case count := <-existentialCh:
			existential += count
		case count := <-subjectCountCh:
			subjectCount += count
		case count := <-courseCountCh:
			courseCount += count
		case count := <-sectionCountCh:
			sectionCount += count
		case count := <-meetingCountCh:
			meetingCount += count
		case count := <-metadataCountCh:
			metadataCount += count
		case count := <-serialSubjectCh:
			serialSubject += count
		case count := <-serialCourseCh:
			serialCourse += count
		case count := <-serialSectionCh:
			serialSection += count
		case <-doneAudit:

			auditLogger.WithFields(log.Fields{
				"university_name": university,

				"insertions":       insertions,
				"updates":          updates,
				"upserts":          upserts,
				"existential":      existential,
				"subjectCount":     subjectCount,
				"courseCount":      courseCount,
				"sectionCount":     sectionCount,
				"meetingCount":     meetingCount,
				"metadataCount":    metadataCount,
				"serialSubject":    serialSubject,
				"serialCourse":     serialCourse,
				"serialSection":    serialSection,
				"elapsed":          time.Since(start).Seconds(),
			}).Info("done!")

			doneAudit <- true
			break Outerloop // Break out of loop to end goroutine
		}
	}
}

var (
	influxClient client.Client
	auditLogger *log.Logger
)

func initInflux() {
	var err error
	// Create the InfluxDB client.
	influxClient, err = influxdbhelper.GetClient(config)

	if err != nil {
		log.Fatalf("Error while creating the client: %v", err)
	}

	// Create and add the hook.
	auditHook, err := influxus.NewHook(
		&influxus.Config{
			Client:             influxClient,
			Database:           "universityct", // DATABASE MUST BE CREATED
			DefaultMeasurement: "ein_ops",
			BatchSize:          1, // default is 100
			BatchInterval:      1, // default is 5 seconds
			Tags:               []string{"university_name"},
			Precision: "s",
		})

	uct.CheckError(err)

	// Add the hook to the standard logger.
	auditLogger = log.New()
	auditLogger.Hooks.Add(auditHook)
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

	serialCourseCh  = make(chan int)
	serialSectionCh = make(chan int)
	serialSubjectCh = make(chan int)

	doneAudit = make(chan bool)
)

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

