package main

import (
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

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
}

func main() {
	app := kingpin.New("ein", "A command-line application for inserting and updated university information")
	noDiff := app.Flag("no-diff", "do not diff against last data").Default("false").Bool()
	fullUpsert := app.Flag("insert-all", "full insert/update of all objects.").Default("true").Short('a').Bool()
	format := app.Flag("format", "choose input format").Short('f').HintOptions(model.Json, model.Protobuf).PlaceHolder("[protobuf, json]").Required().String()
	configFile := app.Flag("config", "configuration file for the application").Short('c').File()
	config := conf.Config{}

	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *format != model.Json && *format != model.Protobuf {
		log.Fatalln("Invalid format:", *format)
	}

	// Parse configuration file
	config = conf.OpenConfig(*configFile)
	config.AppName = app.Name

	// Start profiling
	go model.StartPprof(config.DebugSever(app.Name))

	var database *sqlx.DB
	var err error

	// Initialize database connection
	if database, err = model.OpenPostgres(config.DatabaseConfig(app.Name)); err != nil {
		log.WithError(err).Fatalln()
	} else {
		database.SetMaxOpenConns(config.Postgres.ConnMax)
	}

	ein := &ein{
		app: app.Model(),
		config: &einConfig{
			service:     config,
			noDiff:      *noDiff,
			fullUpsert:  *fullUpsert,
			inputFormat: *format,
		},
		redis:     redis.NewHelper(config, app.Name),
		postgres: DatabaseHandlerImpl{
			database: database,
			statements: make(map[string]*sqlx.NamedStmt),
		},
	}

	ein.init()
}

func (ein *ein) init() {
	ein.postgres.prepareStatements()

	for {
		log.Infoln("Waiting on queue...")
		if data, err := ein.redis.Client.BRPop(0, redis.ScraperQueue).Result(); err != nil {
			log.WithError(err).Fatalln()
		} else {
			ein.process(data)
		}
	}
}

func (ein *ein) process(data []string) {
	defer func() {
		if r := recover(); r != nil {
			log.WithError(fmt.Errorf("Recovered from error in queue loop: %v", r)).Errorln()
		}
	}()

	val := data[1]

	latestData := val + ":data:latest"
	oldData := val + ":data:old"

	log.WithFields(log.Fields{"key": val}).Debugln("RPOP")

	if raw, err := ein.redis.Client.Get(latestData).Bytes(); err != nil {
		log.WithError(err).Panic("Error getting latest data")
	} else {
		var university model.University

		// Try getting older data from redis
		var oldRaw string
		if oldRaw, err = ein.redis.Client.Get(oldData).Result(); err != nil {
			log.Warningln("There was no older data, did it expire or is this first run?")
		}

		// Decode new data
		var newUniversity model.University
		if err := model.UnmarshalMessage(ein.config.inputFormat, bytes.NewReader(raw), &newUniversity); err != nil {
			log.WithError(err).Panic("Error while unmarshalling new data")
		}

		// Make sure the data received is primed for the database
		if err := model.ValidateAll(&newUniversity); err != nil {
			log.WithError(err).Panic("Error while validating newUniversity")
		}

		// Decode old data if have some
		if oldRaw != "" && !ein.config.noDiff {
			var oldUniversity model.University
			if err := model.UnmarshalMessage(ein.config.inputFormat, strings.NewReader(oldRaw), &oldUniversity); err != nil {
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
		if _, err := ein.redis.Client.Set(oldData, raw, 0).Result(); err != nil {
			log.WithError(err).Panic("Error updating old data")
		}

		go audit(university.TopicName)

		ein.insertUniversity(university)
		ein.updateSerial(raw, university)
		// Log bytes received
		log.WithFields(log.Fields{"bytes": len([]byte(raw)), "university_name": university.TopicName}).Infoln(latestData)

		doneAudit <- true
		<-doneAudit
		//break
	}
}

// uses raw because the previously validated university was mutated some where and I couldn't find where
func (ein *ein) updateSerial(raw []byte, diff model.University) {
	defer model.TimeTrack(time.Now(), "updateSerial")

	// Decode new data
	var newUniversity model.University
	if err := model.UnmarshalMessage(ein.config.inputFormat, bytes.NewReader(raw), &newUniversity); err != nil {
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

	sem := make(chan bool, ein.config.service.Postgres.ConnMax)
	for subjectIndex := range newUniversity.Subjects {
		subject := newUniversity.Subjects[subjectIndex]
		ein.updateSerialSubject(subject)
	}

	cwg := sync.WaitGroup{}
	for courseIndex := range diffCourses {
		course := diffCourses[courseIndex]

		cwg.Add(1)

		sem <- true
		go func() {
			ein.updateSerialCourse(course)

			for sectionIndex := range course.Sections {
				section := course.Sections[sectionIndex]
				ein.updateSerialSection(section)
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

func (ein *ein) updateSerialSubject(subject *model.Subject) {
	data, err := subject.Marshal()
	if err != nil {
		log.WithError(err).Fatalln("failed to marshal subject")
	}
	arg := serialSubject{serial{TopicName: subject.TopicName, Data: data}}
	ein.postgres.update(SerialSubjectUpdateQuery, arg)

	// Sanity Check
	log.WithFields(log.Fields{"subject": subject.TopicId, "bytes": len(data)}).Debugln("sanity")
}

func (ein *ein) updateSerialCourse(course *model.Course) {
	data, err := course.Marshal()
	if err != nil {
		log.WithError(err).Fatalln("failed to marshal course")
	}
	arg := serialCourse{serial{TopicName: course.TopicName, Data: data}}
	ein.postgres.update(SerialCourseUpdateQuery, arg)

	// Sanity Check
	log.WithFields(log.Fields{"course": course.TopicId, "bytes": len(data)}).Debugln("sanity")
}

func (ein *ein) updateSerialSection(section *model.Section) {
	data, err := section.Marshal()
	if err != nil {
		log.WithError(err).Fatalln("failed to marshal section")
	}
	arg := serialSection{serial{TopicName: section.TopicName, Data: data}}
	ein.postgres.update(SerialSectionUpdateQuery, arg)

	// Sanity Check
	log.WithFields(log.Fields{"section": section.TopicId, "bytes": len(data)}).Debugln("sanity")
}

func (ein *ein) insertUniversity(university model.University) {
	defer model.TimeTrack(time.Now(), "insertUniversity")

	university.Id = ein.postgres.upsert(UniversityInsertQuery, UniversityUpdateQuery, university)

	ein.insertSubjects(&university)

	// ResolvedSemesters
	ein.insertSemester(&university)

	// Registrations
	for _, registrations := range university.Registrations {
		registrations.UniversityId = university.Id
		ein.insertRegistration(registrations)
	}

	// university []Metadata
	metadata := university.Metadata
	for metadataIndex := range metadata {
		metadata := metadata[metadataIndex]

		metadata.UniversityId = &university.Id
		ein.insertMetadata(metadata)
	}
}

func (ein *ein) insertSubjects(university *model.University) {

	for subjectIndex := range university.Subjects {
		subject := university.Subjects[subjectIndex]
		subject.UniversityId = university.Id

		subject.Id = ein.insertSubject(subject)

		ein.insertCourses(subject)

	}
}

func (ein *ein) insertCourses(subject *model.Subject) {
	courses := subject.Courses
	for courseIndex := range courses {
		course := courses[courseIndex]

		course.SubjectId = subject.Id
		course.Id = ein.insertCourse(course)

		ein.insertSections(course)

		// Course []Metadata
		metadatas := course.Metadata
		for metadataIndex := range metadatas {
			metadata := metadatas[metadataIndex]

			metadata.CourseId = &course.Id
			ein.insertMetadata(metadata)
		}
	}
}

func (ein *ein) insertSections(course *model.Course) {
	sections := course.Sections

	for sectionIndex := range sections {
		section := sections[sectionIndex]

		section.CourseId = course.Id
		sectionId := ein.insertSection(section)
		// Make section data available as soon as possible
		ein.updateSerialSection(section)

		//[]Instructors
		instructors := section.Instructors
		for instructorIndex := range instructors {
			instructor := instructors[instructorIndex]
			instructor.SectionId = sectionId
			ein.insertInstructor(instructor)
		}

		//[]Meeting
		meetings := section.Meetings
		for meetingIndex := range meetings {
			meeting := meetings[meetingIndex]

			meeting.SectionId = sectionId
			meetingId := ein.insertMeeting(meeting)

			// Meeting []Metadata
			metadata := meeting.Metadata
			for metadataIndex := range metadata {
				metadata := metadata[metadataIndex]

				metadata.MeetingId = &meetingId
				ein.insertMetadata(metadata)
			}
		}

		//[]Books
		books := section.Books
		for bookIndex := range books {
			book := books[bookIndex]

			book.SectionId = sectionId
			ein.insertBook(book)
		}

		// Section []Metadata
		metadata := section.Metadata
		for metadataIndex := range metadata {
			metadata := metadata[metadataIndex]

			metadata.SectionId = &sectionId
			ein.insertMetadata(metadata)
		}
	}

}

func (ein *ein) insertSubject(sub *model.Subject) (subjectId int64) {
	if !ein.config.fullUpsert {
		if subjectId = ein.postgres.exists(SubjectExistQuery, sub); subjectId != 0 {
			return
		}
	}
	subjectId = ein.postgres.upsert(SubjectInsertQuery, SubjectUpdateQuery, sub)

	// Subject []Metadata
	metadatas := sub.Metadata
	for metadataIndex := range metadatas {
		metadata := metadatas[metadataIndex]

		metadata.SubjectId = &subjectId
		ein.insertMetadata(metadata)
	}
	return subjectId
}

func (ein *ein) insertCourse(course *model.Course) (courseId int64) {
	if !ein.config.fullUpsert {
		if courseId = ein.postgres.exists(CourseExistQuery, course); courseId != 0 {
			return
		}
	}
	courseId = ein.postgres.upsert(CourseInsertQuery, CourseUpdateQuery, course)

	return courseId
}

func (ein *ein) insertSemester(university *model.University) int64 {
	return ein.postgres.upsert(SemesterInsertQuery, SemesterUpdateQuery, &model.DBResolvedSemester{
		UniversityId: university.Id,
		CurrentSeason: university.ResolvedSemesters.Current.Season,
		CurrentYear: strconv.Itoa(int(university.ResolvedSemesters.Current.Year)),
		LastSeason: university.ResolvedSemesters.Last.Season,
		LastYear: strconv.Itoa(int(university.ResolvedSemesters.Last.Year)),
		NextSeason: university.ResolvedSemesters.Next.Season,
		NextYear: strconv.Itoa(int(university.ResolvedSemesters.Next.Year)),
	})
}

func (ein *ein) insertSection(section *model.Section) int64 {
	return ein.postgres.upsert(SectionInsertQuery, SectionUpdateQuery, section)
}

func (ein *ein) insertMeeting(meeting *model.Meeting) (meetingId int64) {
	if !!ein.config.fullUpsert {
		if meetingId = ein.postgres.exists(MeetingExistQuery, meeting); meetingId != 0 {
			return
		}
	}
	return ein.postgres.upsert(MeetingInsertQuery, MeetingUpdateQuery, meeting)
}

func (ein *ein) insertInstructor(instructor *model.Instructor) (instructorId int64) {
	if instructorId = ein.postgres.exists(InstructorExistQuery, instructor); instructorId != 0 {
		return
	}
	return ein.postgres.upsert(InstructorInsertQuery, InstructorUpdateQuery, instructor)
}

func (ein *ein) insertBook(book *model.Book) (bookId int64) {
	bookId = ein.postgres.upsert(BookInsertQuery, BookUpdateQuery, book)

	return bookId
}

func (ein *ein) insertRegistration(registration *model.Registration) int64 {
	return ein.postgres.upsert(RegistrationInsertQuery, RegistrationUpdateQuery, registration)
}

func (ein *ein) insertMetadata(metadata *model.Metadata) (metadataId int64) {
	var insertQuery string
	var updateQuery string

	if metadata.UniversityId != nil {
		if !ein.config.fullUpsert {
			if metadataId = ein.postgres.exists(MetaUniExistQuery, metadata); metadataId != 0 {
				return
			}
		}
		updateQuery = MetaUniUpdateQuery
		insertQuery = MetaUniInsertQuery

	} else if metadata.SubjectId != nil {
		if !ein.config.fullUpsert {
			if metadataId = ein.postgres.exists(MetaSubjectExistQuery, metadata); metadataId != 0 {
				return
			}
		}
		updateQuery = MetaSubjectUpdateQuery
		insertQuery = MetaSubjectInsertQuery

	} else if metadata.CourseId != nil {
		if !ein.config.fullUpsert {
			if metadataId = ein.postgres.exists(MetaCourseExistQuery, metadata); metadataId != 0 {
				return
			}
		}
		updateQuery = MetaCourseUpdateQuery
		insertQuery = MetaCourseInsertQuery

	} else if metadata.SectionId != nil {
		if !ein.config.fullUpsert {
			if metadataId = ein.postgres.exists(MetaSectionExistQuery, metadata); metadataId != 0 {
				return
			}
		}
		updateQuery = MetaSectionUpdateQuery
		insertQuery = MetaSectionInsertQuery

	} else if metadata.MeetingId != nil {
		if !ein.config.fullUpsert {
			if metadataId = ein.postgres.exists(MetaMeetingExistQuery, metadata); metadataId != 0 {
				return
			}
		}
		updateQuery = MetaMeetingUpdateQuery
		insertQuery = MetaMeetingInsertQuery
	}
	return ein.postgres.upsert(insertQuery, updateQuery, metadata)
}
