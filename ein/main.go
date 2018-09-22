package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"context"

	log "github.com/Sirupsen/logrus"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/tevjef/uct-backend/common/conf"
	"github.com/tevjef/uct-backend/common/database"
	_ "github.com/tevjef/uct-backend/common/metrics"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/common/redis"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type ein struct {
	app      *kingpin.ApplicationModel
	config   *einConfig
	redis    *redis.Helper
	postgres database.Handler
	metrics  metrics
	ctx      context.Context
}

type metrics struct {
	insertions  *prometheus.GaugeVec
	updates     *prometheus.GaugeVec
	upserts     *prometheus.GaugeVec
	existential *prometheus.GaugeVec

	subject  *prometheus.GaugeVec
	course   *prometheus.GaugeVec
	section  *prometheus.GaugeVec
	meeting  *prometheus.GaugeVec
	metadata *prometheus.GaugeVec

	diffSubject  *prometheus.GaugeVec
	diffCourse   *prometheus.GaugeVec
	diffSection  *prometheus.GaugeVec
	diffMeeting  *prometheus.GaugeVec
	diffMetadata *prometheus.GaugeVec

	diffSerialCourse   *prometheus.GaugeVec
	diffSerialSection  *prometheus.GaugeVec
	diffSerialSubject  *prometheus.GaugeVec
	diffSerialMeeting  *prometheus.GaugeVec
	diffSerialMetadata *prometheus.GaugeVec

	elapsed      *prometheus.GaugeVec
	payloadBytes *prometheus.GaugeVec
}

type einConfig struct {
	service     conf.Config
	noDiff      bool
	fullUpsert  bool
	inputFormat string
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
}

func main() {
	econf := &einConfig{}

	app := kingpin.New("ein", "A command-line application for inserting and updated university information")

	app.Flag("no-diff", "do not diff against last data").
		Default("false").
		Envar("EIN_NO_DIFF").
		BoolVar(&econf.noDiff)

	app.Flag("insert-all", "full insert/update of all objects.").
		Default("true").
		Short('a').
		Envar("EIN_INSERT_ALL").
		BoolVar(&econf.fullUpsert)

	app.Flag("format", "choose input format").
		Short('f').
		HintOptions(model.Json, model.Protobuf).
		PlaceHolder("[protobuf, json]").
		Required().
		Envar("EIN_INPUT_FORMAT").
		EnumVar(&econf.inputFormat, "protobuf", "json")

	configFile := app.Flag("config", "configuration file for the application").
		Short('c').
		Envar("EIN_CONFIG").
		File()

	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Parse configuration file
	econf.service = conf.OpenConfigWithName(*configFile, app.Name)

	// Start profiling
	go model.StartPprof(econf.service.DebugSever(app.Name))

	pgDatabase, err := model.OpenPostgres(econf.service.DatabaseConfig(app.Name))
	if err != nil {
		log.WithError(err).Fatalln()
	}

	pgDatabase.SetMaxOpenConns(econf.service.Postgres.ConnMax)

	appMetrics := metrics{
		insertions: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_database_insertions",
			Help: "Number of records inserted into the database",
		}, []string{"university_name"}),

		updates: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_database_updates",
			Help: "Number of records updated in the database",
		}, []string{"university_name"}),

		upserts: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_database_upserts",
			Help: "Number of records upserted in the database",
		}, []string{"university_name"}),

		existential: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_database_existential",
			Help: "Number of existential queries performed on the database",
		}, []string{"university_name"}),

		subject: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_uct_subject",
			Help: "Number of subject objects",
		}, []string{"university_name"}),

		course: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_uct_course",
			Help: "Number of course objects",
		}, []string{"university_name"}),

		section: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_uct_section",
			Help: "Number of section objects",
		}, []string{"university_name"}),

		meeting: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_uct_meeting",
			Help: "Number of meeting objects",
		}, []string{"university_name"}),

		metadata: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_uct_metadata",
			Help: "Number of metadata objects",
		}, []string{"university_name"}),

		diffSubject: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_uct_subject_diff",
			Help: "Number of diff subject objects",
		}, []string{"university_name"}),

		diffCourse: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_uct_course_diff",
			Help: "Number of diff course objects",
		}, []string{"university_name"}),

		diffSection: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_uct_section_diff",
			Help: "Number of diff section objects",
		}, []string{"university_name"}),

		diffMeeting: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_uct_meeting_diff",
			Help: "Number of diff meeting objects",
		}, []string{"university_name"}),

		diffMetadata: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_uct_metadata_diff",
			Help: "Number of diff metadata objects",
		}, []string{"university_name"}),

		diffSerialSubject: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_uct_subject_serial_diff",
			Help: "Diff of serialized subject objects",
		}, []string{"university_name"}),

		diffSerialCourse: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_uct_course_serial_diff",
			Help: "Diff of serialized course objects",
		}, []string{"university_name"}),

		diffSerialSection: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_uct_section_serial_diff",
			Help: "Diff of serialized section objects",
		}, []string{"university_name"}),

		diffSerialMeeting: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_uct_meeting_serial_diff",
			Help: "Diff of serialized meeting objects",
		}, []string{"university_name"}),

		diffSerialMetadata: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_uct_metadata_serial_diff",
			Help: "Diff of serialized metadata objects",
		}, []string{"university_name"}),

		elapsed: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_process_elapsed_seconds",
			Help: "Time taken to process all objects",
		}, []string{"university_name"}),

		payloadBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ein_payload_bytes",
			Help: "Size of the the data to process",
		}, []string{"university_name"}),
	}

	prometheus.MustRegister(
		appMetrics.insertions,
		appMetrics.updates,
		appMetrics.upserts,
		appMetrics.existential,
		appMetrics.subject,
		appMetrics.course,
		appMetrics.section,
		appMetrics.meeting,
		appMetrics.metadata,
		appMetrics.diffSubject,
		appMetrics.diffCourse,
		appMetrics.diffSection,
		appMetrics.diffMeeting,
		appMetrics.diffMetadata,
		appMetrics.diffSerialSubject,
		appMetrics.diffSerialCourse,
		appMetrics.diffSerialSection,
		appMetrics.diffSerialMeeting,
		appMetrics.diffSerialMetadata,
		appMetrics.elapsed,
		appMetrics.payloadBytes,
	)

	(&ein{
		app:      app.Model(),
		config:   econf,
		redis:    redis.NewHelper(econf.service, app.Name),
		postgres: database.NewHandler(app.Name, pgDatabase, queries),
		metrics:  appMetrics,
	}).init()
}

func (ein *ein) init() {
	for {
		log.Infoln("waiting on queue:", redis.ScraperQueue)
		if data, err := ein.redis.Client.BRPop(0, redis.ScraperQueue).Result(); err != nil {
			log.WithError(err).Fatalln()
		} else {
			if err := ein.process(data); err != nil {
				log.WithError(err).Errorln("failure while processing data")
				continue
			}
		}
	}
}

func (ein *ein) process(data []string) error {
	defer func() {
		if r := recover(); r != nil {
			log.WithError(fmt.Errorf("recovered from error in queue loop: %v", r)).Errorln()
		}
	}()

	val := data[1]
	latestData := val + ":data:latest"
	oldData := val + ":data:old"

	log.WithFields(log.Fields{"key": val}).Debugln("RPOP")

	raw, err := ein.redis.Client.Get(latestData).Bytes()
	if err != nil {
		return errors.New("error while getting latest data")
	}

	var university model.University

	// Try getting older data from redis
	var oldRaw string
	if oldRaw, err = ein.redis.Client.Get(oldData).Result(); err != nil {
		log.Warningln("there was no older data, did it expire or is this first run?")
	}

	// Decode new data
	var newUniversity model.University
	if err := model.UnmarshalMessage(ein.config.inputFormat, bytes.NewReader(raw), &newUniversity); err != nil {
		return errors.Wrap(err, "error while unmarshalling new data")
	}

	// Make sure the data received is primed for the database
	if err := model.ValidateAll(&newUniversity); err != nil {
		return errors.Wrap(err, "error while validating newUniversity")
	}

	// Decode old data if have some
	if oldRaw != "" && !ein.config.noDiff {
		var oldUniversity model.University
		if err := model.UnmarshalMessage(ein.config.inputFormat, strings.NewReader(oldRaw), &oldUniversity); err != nil {
			return errors.Wrap(err, "error while unmarshalling old data")
		}

		if err := model.ValidateAll(&oldUniversity); err != nil {
			return errors.Wrap(err, "error while validating oldUniversity")
		}

		university = model.DiffAndFilter(oldUniversity, newUniversity)

	} else {
		university = newUniversity
	}

	// Replace old data with new data we just received.
	if _, err := ein.redis.Client.Set(oldData, raw, 0).Result(); err != nil {
		return errors.Wrap(err, "error updating old data")
	}

	ein.metrics.payloadBytes.With(prometheus.Labels{"university_name": university.TopicName}).Set(float64(len([]byte(raw))))
	// Log bytes received
	log.WithFields(log.Fields{"bytes": len([]byte(raw)), "university_name": university.TopicName}).Infoln(latestData)

	go statsCollector(ein, university.TopicName)

	ein.insertUniversity(university)
	ein.updateSerial(raw, university)

	collectDatabaseStats(ein.postgres)
	doneAudit <- true
	<-doneAudit
	//break

	return nil
}

// uses raw because the previously validated university was mutated some where and I couldn't find where
func (ein *ein) updateSerial(raw []byte, diff model.University) {
	defer model.TimeTrack(time.Now(), "updateSerial")

	// Decode new data
	var newUniversity model.University
	if err := model.UnmarshalMessage(ein.config.inputFormat, bytes.NewReader(raw), &newUniversity); err != nil {
		log.WithError(err).Fatalln("error while unmarshalling new data")
	}

	// Make sure the data received is primed for the database
	if err := model.ValidateAll(&newUniversity); err != nil {
		log.WithError(err).Fatalln("error while validating newUniversity")
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

// For ever course that's in the diff return the course that has full data.
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

type serial struct {
	TopicName string `db:"topic_name"`
	Data      []byte `db:"data"`
}

func (ein *ein) updateSerialSubject(subject *model.Subject) {
	data, err := subject.Marshal()
	if err != nil {
		log.WithError(err).Fatalln("failed to marshal subject")
	}
	arg := serial{TopicName: subject.TopicName, Data: data}
	ein.postgres.Update(SerialSubjectUpdateQuery, arg)

	// Sanity Check
	log.WithFields(log.Fields{"subject": subject.TopicId, "bytes": len(data)}).Debugln("sanity")
}

func (ein *ein) updateSerialCourse(course *model.Course) {
	data, err := course.Marshal()
	if err != nil {
		log.WithError(err).Fatalln("failed to marshal course")
	}
	arg := serial{TopicName: course.TopicName, Data: data}
	ein.postgres.Update(SerialCourseUpdateQuery, arg)

	// Sanity Check
	log.WithFields(log.Fields{"course": course.TopicId, "bytes": len(data)}).Debugln("sanity")
}

func (ein *ein) updateSerialSection(section *model.Section) {
	data, err := section.Marshal()
	if err != nil {
		log.WithError(err).Fatalln("failed to marshal section")
	}
	arg := serial{TopicName: section.TopicName, Data: data}
	ein.postgres.Update(SerialSectionUpdateQuery, arg)

	// Sanity Check
	log.WithFields(log.Fields{"section": section.TopicId, "bytes": len(data)}).Debugln("sanity")
}

func (ein *ein) insertUniversity(university model.University) {
	defer model.TimeTrack(time.Now(), "insertUniversity")

	university.Id = ein.postgres.Upsert(UniversityInsertQuery, UniversityUpdateQuery, university)

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
		if subjectId = ein.postgres.Exists(SubjectExistQuery, sub); subjectId != 0 {
			return
		}
	}
	subjectId = ein.postgres.Upsert(SubjectInsertQuery, SubjectUpdateQuery, sub)

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
		if courseId = ein.postgres.Exists(CourseExistQuery, course); courseId != 0 {
			return
		}
	}
	courseId = ein.postgres.Upsert(CourseInsertQuery, CourseUpdateQuery, course)

	return courseId
}

func (ein *ein) insertSemester(university *model.University) int64 {
	return ein.postgres.Upsert(SemesterInsertQuery, SemesterUpdateQuery, &model.DBResolvedSemester{
		UniversityId:  university.Id,
		CurrentSeason: university.ResolvedSemesters.Current.Season,
		CurrentYear:   strconv.Itoa(int(university.ResolvedSemesters.Current.Year)),
		LastSeason:    university.ResolvedSemesters.Last.Season,
		LastYear:      strconv.Itoa(int(university.ResolvedSemesters.Last.Year)),
		NextSeason:    university.ResolvedSemesters.Next.Season,
		NextYear:      strconv.Itoa(int(university.ResolvedSemesters.Next.Year)),
	})
}

func (ein *ein) insertSection(section *model.Section) int64 {
	return ein.postgres.Upsert(SectionInsertQuery, SectionUpdateQuery, section)
}

func (ein *ein) insertMeeting(meeting *model.Meeting) (meetingId int64) {
	if !!ein.config.fullUpsert {
		if meetingId = ein.postgres.Exists(MeetingExistQuery, meeting); meetingId != 0 {
			return
		}
	}
	return ein.postgres.Upsert(MeetingInsertQuery, MeetingUpdateQuery, meeting)
}

func (ein *ein) insertInstructor(instructor *model.Instructor) (instructorId int64) {
	if instructorId = ein.postgres.Exists(InstructorExistQuery, instructor); instructorId != 0 {
		return
	}
	return ein.postgres.Upsert(InstructorInsertQuery, InstructorUpdateQuery, instructor)
}

func (ein *ein) insertBook(book *model.Book) (bookId int64) {
	bookId = ein.postgres.Upsert(BookInsertQuery, BookUpdateQuery, book)

	return bookId
}

func (ein *ein) insertRegistration(registration *model.Registration) int64 {
	return ein.postgres.Upsert(RegistrationInsertQuery, RegistrationUpdateQuery, registration)
}

func (ein *ein) insertMetadata(metadata *model.Metadata) (metadataId int64) {
	var insertQuery string
	var updateQuery string

	if metadata.UniversityId != nil {
		if !ein.config.fullUpsert {
			if metadataId = ein.postgres.Exists(MetaUniExistQuery, metadata); metadataId != 0 {
				return
			}
		}
		updateQuery = MetaUniUpdateQuery
		insertQuery = MetaUniInsertQuery

	} else if metadata.SubjectId != nil {
		if !ein.config.fullUpsert {
			if metadataId = ein.postgres.Exists(MetaSubjectExistQuery, metadata); metadataId != 0 {
				return
			}
		}
		updateQuery = MetaSubjectUpdateQuery
		insertQuery = MetaSubjectInsertQuery

	} else if metadata.CourseId != nil {
		if !ein.config.fullUpsert {
			if metadataId = ein.postgres.Exists(MetaCourseExistQuery, metadata); metadataId != 0 {
				return
			}
		}
		updateQuery = MetaCourseUpdateQuery
		insertQuery = MetaCourseInsertQuery

	} else if metadata.SectionId != nil {
		if !ein.config.fullUpsert {
			if metadataId = ein.postgres.Exists(MetaSectionExistQuery, metadata); metadataId != 0 {
				return
			}
		}
		updateQuery = MetaSectionUpdateQuery
		insertQuery = MetaSectionInsertQuery

	} else if metadata.MeetingId != nil {
		if !ein.config.fullUpsert {
			if metadataId = ein.postgres.Exists(MetaMeetingExistQuery, metadata); metadataId != 0 {
				return
			}
		}
		updateQuery = MetaMeetingUpdateQuery
		insertQuery = MetaMeetingInsertQuery
	}
	return ein.postgres.Upsert(insertQuery, updateQuery, metadata)
}
