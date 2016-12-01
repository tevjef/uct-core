package main

import (
	_ "expvar"
	"fmt"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"uct/common/conf"
	"uct/common/model"
	"uct/redis"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"bytes"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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
	noDiff     = app.Flag("no-diff", "do not diff against last data").Default("false").Bool()
	fullUpsert = app.Flag("insert-all", "full insert/update of all objects.").Default("true").Short('a').Bool()
	format     = app.Flag("format", "choose input format").Short('f').HintOptions(model.Json, model.Protobuf).PlaceHolder("[protobuf, json]").Required().String()
	configFile = app.Flag("config", "configuration file for the application").Short('c').File()
	config     = conf.Config{}

	multiProgramming = 5
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *format != model.Json && *format != model.Protobuf {
		log.Fatalln("Invalid format:", *format)
	}

	log.SetLevel(log.InfoLevel)

	// Parse configuration file
	config = conf.OpenConfig(*configFile)
	config.AppName = app.Name

	// Start profiling
	go model.StartPprof(config.DebugSever(app.Name))

	// Start redis client
	wrapper := redis.NewHelper(config, app.Name)

	var database *sqlx.DB
	var err error

	// Initialize database connection
	if database, err = model.InitDB(config.DatabaseConfig(app.Name)); err != nil {
		log.WithError(err).Fatalln()
	}

	dbHandler := DatabaseHandlerImpl{Database: database}
	dbHandler.PrepareAllStmts()
	app := App{dbHandler: dbHandler}
	database.SetMaxOpenConns(multiProgramming)

	for {
		log.Infoln("Waiting on queue...")
		if data, err := wrapper.Client.BRPop(10*time.Minute, redis.ScraperQueue).Result(); err != nil {
			log.WithError(err).Warningln("Queue blocking exceeded timeout")
			continue
		} else {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.WithError(fmt.Errorf("Recovered from error in queue loop: %v", r)).Errorln()
					}
				}()

				val := data[1]

				latestData := val + ":data:latest"
				oldData := val + ":data:old"

				log.WithFields(log.Fields{"key": val}).Debugln("RPOP")

				if raw, err := wrapper.Client.Get(latestData).Bytes(); err != nil {
					log.WithError(err).Panic("Error getting latest data")
				} else {
					var university model.University

					// Try getting older data from redis
					var oldRaw string
					if oldRaw, err = wrapper.Client.Get(oldData).Result(); err != nil {
						log.Warningln("There was no older data, did it expire or is this first run?")
					}

					// Decode new data
					var newUniversity model.University
					if err := model.UnmarshalMessage(*format, bytes.NewReader(raw), &newUniversity); err != nil {
						log.WithError(err).Panic("Error while unmarshalling new data")
					}

					// Make sure the data received is primed for the database
					if err := model.ValidateAll(&newUniversity); err != nil {
						log.WithError(err).Panic("Error while validating newUniversity")
					}

					// Decode old data if have some
					if oldRaw != "" && !*noDiff {
						var oldUniversity model.University
						if err := model.UnmarshalMessage(*format, strings.NewReader(oldRaw), &oldUniversity); err != nil {
							log.WithError(err).Panic("Error while unmarshalling old data")
						}

						if err := model.ValidateAll(&oldUniversity); err != nil {
							log.WithError(err).Panic("Error while validating oldUniversity")
						}

						university = model.DiffAndFilter(oldUniversity, newUniversity)

					} else {
						university = newUniversity
					}

					// Set old data as the new data we just received. Important that this is after validating the new raw data
					if _, err := wrapper.Client.Set(oldData, raw, 0).Result(); err != nil {
						log.WithError(err).Panic("Error updating old data")
					}

					go audit(university.TopicName)

					app.insertUniversity(university)
					app.updateSerial(raw, university)
					// Log bytes received
					log.WithFields(log.Fields{"bytes": len([]byte(raw)), "university_name": university.TopicName}).Infoln(latestData)

					doneAudit <- true
					<-doneAudit
					//break
				}

			}()

		}
	}
}

// uses raw because the previously validated university was mutated some where and I couldn't find where
func (app App) updateSerial(raw []byte, diff model.University) {
	defer model.TimeTrack(time.Now(), "updateSerial")

	// Decode new data
	var newUniversity model.University
	if err := model.UnmarshalMessage(*format, bytes.NewReader(raw), &newUniversity); err != nil {
		log.WithError(err).Panic("Error while unmarshalling new data")
	}

	// Make sure the data received is primed for the database
	if err := model.ValidateAll(&newUniversity); err != nil {
		log.WithError(err).Panic("Error while validating newUniversity")
	}

	diffCourses := diffAndMergeCourses(newUniversity, diff)

	// log number of subjects, courses, and sections found
	countUniversity(newUniversity, subjectCh, courseCh, sectionCh, meetingCh, metadataCh)
	countUniversity(diff, diffSubjectCh, diffCourseCh, diffSectionCh, diffMeetingCh, diffMetadataCh)
	countSubjects(newUniversity.Subjects, diffCourses, diffSerialSubjectCh, diffSerialCourseCh, diffSerialSectionCh, diffSerialMeetingCountCh, diffSerialMetadataCountCh)

	sem := make(chan bool, multiProgramming)
	for subjectIndex := range newUniversity.Subjects {
		subject := newUniversity.Subjects[subjectIndex]
		app.updateSerialSubject(subject)
	}

	cwg := sync.WaitGroup{}
	for courseIndex := range diffCourses {
		course := diffCourses[courseIndex]

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

func diffAndMergeCourses(full, diff model.University) (coursesToUpdate []*model.Course) {
	allCourses := []*model.Course{}
	diffCourses := []*model.Course{}

	for i := range full.Subjects {
		allCourses = append(allCourses, full.Subjects[i].Courses...)
	}

	for i := range diff.Subjects {
		diffCourses = append(diffCourses, diff.Subjects[i].Courses...)
	}

	for i := range diffCourses {
		course := diffCourses[i]
		for j := range allCourses {
			fullCourse := allCourses[j]
			if course.TopicName == fullCourse.TopicName {
				coursesToUpdate = append(coursesToUpdate, fullCourse)
			}
		}
	}

	return coursesToUpdate
}

func (app App) updateSerialSubject(subject *model.Subject) {
	data, err := subject.Marshal()
	if err != nil {
		log.WithError(err).Fatalln("failed to marshal subject")
	}
	arg := serialSubject{serial{TopicName: subject.TopicName, Data: data}}
	app.dbHandler.update(SerialSubjectUpdateQuery, arg)

	// Sanity Check
	log.WithFields(log.Fields{"subject": subject.TopicId, "bytes": len(data)}).Debugln("sanity")
}

func (app App) updateSerialCourse(course *model.Course) {
	data, err := course.Marshal()
	if err != nil {
		log.WithError(err).Fatalln("failed to marshal course")
	}
	arg := serialCourse{serial{TopicName: course.TopicName, Data: data}}
	app.dbHandler.update(SerialCourseUpdateQuery, arg)

	// Sanity Check
	log.WithFields(log.Fields{"course": course.TopicId, "bytes": len(data)}).Debugln("sanity")
}

func (app App) updateSerialSection(section *model.Section) {
	data, err := section.Marshal()
	if err != nil {
		log.WithError(err).Fatalln("failed to marshal section")
	}
	arg := serialSection{serial{TopicName: section.TopicName, Data: data}}
	app.dbHandler.update(SerialSectionUpdateQuery, arg)

	// Sanity Check
	log.WithFields(log.Fields{"section": section.TopicId, "bytes": len(data)}).Debugln("sanity")
}

func (app App) insertUniversity(uni model.University) {
	defer model.TimeTrack(time.Now(), "insertUniversity")

	universityId := app.dbHandler.upsert(UniversityInsertQuery, UniversityUpdateQuery, uni)

	for subjectIndex := range uni.Subjects {
		subject := uni.Subjects[subjectIndex]
		subject.UniversityId = universityId

		subjectId := app.insertSubject(subject)

		courses := subject.Courses
		for courseIndex := range courses {
			course := courses[courseIndex]

			course.SubjectId = subjectId
			courseId := app.insertCourse(course)

			sections := course.Sections
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
				for meetingIndex := range meetings {
					meeting := meetings[meetingIndex]

					meeting.SectionId = sectionId
					meetingId := app.insertMeeting(meeting)

					// Meeting []Metadata
					metadatas := meeting.Metadata
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
				for metadataIndex := range metadatas {
					metadata := metadatas[metadataIndex]

					metadata.SectionId = &sectionId
					app.insertMetadata(metadata)
				}
			}

			// Course []Metadata
			metadatas := course.Metadata
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

func (app App) insertSubject(sub *model.Subject) (subjectId int64) {
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

func (app App) insertCourse(course *model.Course) (courseId int64) {
	if !*fullUpsert {
		if courseId = app.dbHandler.exists(CourseExistQuery, course); courseId != 0 {
			return
		}
	}
	courseId = app.dbHandler.upsert(CourseInsertQuery, CourseUpdateQuery, course)

	return courseId
}

func (app App) insertSemester(universityId int64, resolvedSemesters *model.ResolvedSemester) int64 {
	rs := &model.DBResolvedSemester{}
	rs.UniversityId = universityId
	rs.CurrentSeason = resolvedSemesters.Current.Season
	rs.CurrentYear = strconv.Itoa(int(resolvedSemesters.Current.Year))
	rs.LastSeason = resolvedSemesters.Last.Season
	rs.LastYear = strconv.Itoa(int(resolvedSemesters.Last.Year))
	rs.NextSeason = resolvedSemesters.Next.Season
	rs.NextYear = strconv.Itoa(int(resolvedSemesters.Next.Year))
	return app.dbHandler.upsert(SemesterInsertQuery, SemesterUpdateQuery, rs)
}

func (app App) insertSection(section *model.Section) int64 {
	return app.dbHandler.upsert(SectionInsertQuery, SectionUpdateQuery, section)
}

func (app App) insertMeeting(meeting *model.Meeting) (meetingId int64) {
	if !*fullUpsert {
		if meetingId = app.dbHandler.exists(MeetingExistQuery, meeting); meetingId != 0 {
			return
		}
	}
	return app.dbHandler.upsert(MeetingInsertQuery, MeetingUpdateQuery, meeting)
}

func (app App) insertInstructor(instructor *model.Instructor) (instructorId int64) {
	if instructorId = app.dbHandler.exists(InstructorExistQuery, instructor); instructorId != 0 {
		return
	}
	return app.dbHandler.upsert(InstructorInsertQuery, InstructorUpdateQuery, instructor)
}

func (app App) insertBook(book *model.Book) (bookId int64) {
	bookId = app.dbHandler.upsert(BookInsertQuery, BookUpdateQuery, book)

	return bookId
}

func (app App) insertRegistration(registration *model.Registration) int64 {
	return app.dbHandler.upsert(RegistrationInsertQuery, RegistrationUpdateQuery, registration)
}

func (app App) insertMetadata(metadata *model.Metadata) (metadataId int64) {
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
