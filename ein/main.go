package ein

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	cloudStorage "cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/storage"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/api/option"

	log "github.com/Sirupsen/logrus"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/tevjef/uct-backend/common/conf"
	_ "github.com/tevjef/uct-backend/common/metrics"
	"github.com/tevjef/uct-backend/common/model"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type ein struct {
	app               *kingpin.ApplicationModel
	config            *einConfig
	firebaseApp       *firebase.App
	storageClient     *storage.Client
	firestoreClient   *firestore.Client
	metrics           metrics
	newUniversityData []byte
	ctx               context.Context
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
	service           conf.Config
	noDiff            bool
	fullUpsert        bool
	inputFormat       string
	firebaseProjectID string
}

func Ein(w http.ResponseWriter, r *http.Request) {
	newUniversityData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithError(err).Errorln("failed to read request body")
	}

	MainFunc(newUniversityData)

	fmt.Fprint(w, "Complete")
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.InfoLevel)
}

func MainFunc(newUniversityData []byte) {
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

	app.Flag("firebase-project-id", "Firebase project Id").
		Default("universitycoursetracker").
		Envar("FIREBASE_PROJECT_ID").
		StringVar(&econf.firebaseProjectID)

	kingpin.MustParse(app.Parse([]string{}))

	ctx := context.Background()

	//credentials, err := google.FindDefaultCredentials(ctx)
	credOption := option.WithCredentialsFile("/Users/tevjef/Desktop/universitycoursetracker-1a1af4ac7a86.json")
	firebaseConf := &firebase.Config{
		ProjectID:     econf.firebaseProjectID,
		StorageBucket: "universitycoursetracker.appspot.com",
	}

	firebaseApp, err := firebase.NewApp(ctx, firebaseConf, credOption)
	if err != nil {
		log.WithError(err).Errorln("failed to crate firebase app")
	}

	storageClient, err := firebaseApp.Storage(ctx)
	if err != nil {
		log.WithError(err).Errorln("failed to crate firebase storage client")
	}

	firestoreClient, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.WithError(err).Errorln("failed to create firestore client")
	}

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
		app:               app.Model(),
		config:            econf,
		firebaseApp:       firebaseApp,
		storageClient:     storageClient,
		firestoreClient:   firestoreClient,
		metrics:           appMetrics,
		newUniversityData: newUniversityData,
		ctx:               context.Background(),
	}).init()
}

func (ein *ein) init() {
	if err := ein.process(); err != nil {
		log.WithError(err).Errorln("failure while processing data")
		return
	}
}

func (ein *ein) process() error {

	// Decode new data
	var newUniversity model.University
	if err := model.UnmarshalMessage(ein.config.inputFormat, bytes.NewReader(ein.newUniversityData), &newUniversity); err != nil {
		return errors.Wrap(err, "error while unmarshalling new data")
	}

	// Make sure the data received is primed for the database
	if err := model.ValidateAll(&newUniversity); err != nil {
		return errors.Wrap(err, "error while validating newUniversity")
	}

	var oldRaw []byte

	bucket, err := ein.storageClient.DefaultBucket()
	if err != nil {
		log.WithError(err).Errorln("failed to get default bucket")
	}
	if oldUniversityReader, err := bucket.Object(newUniversity.TopicName).NewReader(ein.ctx); err == cloudStorage.ErrObjectNotExist {
		log.Warningln("there was no older data, did it expire or is this first run?")
	} else if err != nil {
		log.WithError(err).Warningln("failed to get data from storage bucket")
	} else if oldUniversityReader != nil {
		if oldRaw, err = ioutil.ReadAll(oldUniversityReader); err != nil {
			log.Errorln("failed to read university data")
		}
	}

	var university model.University

	// Decode old data if have some
	if len(oldRaw) != 0 && !ein.config.noDiff {
		var oldUniversity model.University
		if err := model.UnmarshalMessage(ein.config.inputFormat, bytes.NewReader(oldRaw), &oldUniversity); err != nil {
			return errors.Wrap(err, "error while unmarshalling old data")
		}

		if err := model.ValidateAll(&oldUniversity); err != nil {
			return errors.Wrap(err, "error while validating oldUniversity")
		}

		university = model.DiffAndFilter(oldUniversity, newUniversity)

	} else {
		university = newUniversity
	}

	go statsCollector(ein, university.TopicName)

	ein.insertUniversity(newUniversity, university)
	ein.updateSerial(ein.newUniversityData, university)

	//doneAudit <- true
	//<-doneAudit
	//break

	w := bucket.Object(newUniversity.TopicName + ".old").NewWriter(ein.ctx)

	_, err = io.Copy(w, bytes.NewReader(ein.newUniversityData))
	err = w.Close()
	if err != nil {
		log.WithError(err).Errorln("failed to result to cloud storage")
	}

	return nil
}

// uses raw because the previously validated university was mutated some where and I couldn't find where
func (ein *ein) updateSerial(raw []byte, diff model.University) {
	defer model.TimeTrack(time.Now(), "updateSerial")

	// Decode new data
	var newUniversity model.University
	if err := model.UnmarshalMessage(ein.config.inputFormat, bytes.NewReader(raw), &newUniversity); err != nil {
		log.WithError(err).Errorln("error while unmarshalling new data")
	}

	// Make sure the data received is primed for the database
	if err := model.ValidateAll(&newUniversity); err != nil {
		log.WithError(err).Errorln("error while validating newUniversity")
	}

	diffCourses := diffAndMergeCourses(newUniversity, diff)

	// log number of subjects, courses, and sections found
	countUniversity(newUniversity, subjectCh, courseCh, sectionCh, meetingCh, metadataCh)
	countUniversity(diff, diffSubjectCh, diffCourseCh, diffSectionCh, diffMeetingCh, diffMetadataCh)
	countSubjects(newUniversity.Subjects, diffCourses, diffSerialSubjectCh, diffSerialCourseCh, diffSerialSectionCh, diffSerialMeetingCountCh, diffSerialMetadataCountCh)

	var allSections []*model.Section
	for courseIndex := range diffCourses {
		course := diffCourses[courseIndex]
		for sectionIndex := range course.Sections {
			section := course.Sections[sectionIndex]
			allSections = append(allSections, section)
		}
	}

	ein.updateSerialSection(allSections)
}

// For ever course that's in the diff return the course that has full data.
func diffAndMergeCourses(full, diff model.University) (coursesToUpdate []*model.Course) {
	var allCourses []*model.Course
	var diffCourses []*model.Course

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
