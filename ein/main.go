package ein

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	cloudStorage "cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/storage"
	"golang.org/x/oauth2/google"
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
	newUniversityData []byte
	ctx               context.Context
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

	firebaseConf := &firebase.Config{
		ProjectID:     econf.firebaseProjectID,
		StorageBucket: "universitycoursetracker.appspot.com",
	}

	credentials, err := google.FindDefaultCredentials(ctx)
	firebaseApp, err := firebase.NewApp(ctx, firebaseConf, option.WithCredentials(credentials))
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

	(&ein{
		app:               app.Model(),
		config:            econf,
		firebaseApp:       firebaseApp,
		storageClient:     storageClient,
		firestoreClient:   firestoreClient,
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

	objName := "scraper-cache/" + newUniversity.TopicName
	bucket, err := ein.storageClient.DefaultBucket()
	if err != nil {
		log.WithError(err).Errorln("failed to get default bucket")
	}
	if oldUniversityReader, err := bucket.Object(objName).NewReader(ein.ctx); err == cloudStorage.ErrObjectNotExist {
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

	w := bucket.Object(objName).NewWriter(ein.ctx)

	_, err = w.Write(ein.newUniversityData)
	err = w.Close()
	if err != nil {
		log.WithError(err).Errorln("failed to result to cloud storage")
	}

	ein.insertUniversity(newUniversity, university)
	ein.updateSerial(newUniversity, university)

	return nil
}

// uses raw because the previously validated university was mutated some where and I couldn't find where
func (ein *ein) updateSerial(newUniversity model.University, diff model.University) {
	defer model.TimeTrack(time.Now(), "updateSerial")

	var allSections []*model.Section

	for subjectIndex := range diff.Subjects {
		subject := diff.Subjects[subjectIndex]
		for courseIndex := range subject.Courses {
			course := subject.Courses[courseIndex]
			for sectionIndex := range course.Sections {
				section := course.Sections[sectionIndex]
				allSections = append(allSections, section)
			}
		}
	}

	if len(allSections) == 0 {
		log.Infoln("updateSerial: no new sections")
		return
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
