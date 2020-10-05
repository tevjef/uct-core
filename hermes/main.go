package hermes

import (
	"context"
	"fmt"
	_ "net/http/pprof"
	"os"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"firebase.google.com/go/storage"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/conf"
	uctfirestore "github.com/tevjef/uct-backend/common/firestore"
	_ "github.com/tevjef/uct-backend/common/trace"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"gopkg.in/alecthomas/kingpin.v2"
)

type hermes struct {
	app             *kingpin.ApplicationModel
	config          *hermesConfig
	fcmClient       *messaging.Client
	firebaseApp     *firebase.App
	storageClient   *storage.Client
	firestoreClient *firestore.Client
	uctFSClient     *uctfirestore.Client
	event           uctfirestore.FirestoreEvent
	logger          *log.Entry
	ctx             context.Context
}

type hermesConfig struct {
	service           conf.Config
	dryRun            bool
	firebaseProjectID string
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyLevel: "severity",
			log.FieldKeyMsg:   "message",
		},
	})
	log.SetLevel(log.DebugLevel)
}

func Hermes(context context.Context, event uctfirestore.FirestoreEvent) error {
	meta, err := metadata.FromContext(context)
	if err != nil {
		return fmt.Errorf("metadata.FromContext: %v", err)
	}
	log.Debugf("function triggered by change to: %v on event %v", meta, event)
	return MainFunc(event)
}

func MainFunc(firebaseEvent uctfirestore.FirestoreEvent) error {
	hconf := &hermesConfig{}

	app := kingpin.New("hermes", "A server that listens to a database for events and publishes notifications to Firebase Cloud Messaging")

	app.Flag("dry-run", "enable dry-run").
		Short('d').
		Default("true").
		Envar("HERMES_DRY_RUN").
		BoolVar(&hconf.dryRun)

	kingpin.MustParse(app.Parse([]string{}))

	ctx := context.Background()

	if hconf.dryRun {
		log.Infoln("enabling hermes in dry run mode")
	} else {
		log.Infoln("enabling hermes in production mode")
	}

	logger := log.WithFields(log.Fields{})

	firebaseConf := &firebase.Config{
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

	fcmClient, err := firebaseApp.Messaging(ctx)
	if err != nil {
		log.WithError(err).Errorln("failed to create cloud messaging client")
	}

	uctFSClient := uctfirestore.NewClient(ctx, firestoreClient, logger)

	return (&hermes{
		app:             app.Model(),
		config:          hconf,
		fcmClient:       fcmClient,
		firebaseApp:     firebaseApp,
		storageClient:   storageClient,
		firestoreClient: firestoreClient,
		event:           firebaseEvent,
		uctFSClient:     uctFSClient,
		logger:          logger,
		ctx:             context.Background(),
	}).init()
}

func (hermes *hermes) init() error {
	oldValue, err := uctfirestore.FromFirestoreValue(hermes.event.OldValue)
	if err != nil {
		log.WithError(err).Errorln("failed to parse old FirestoreValue from event")
		return err
	}

	oldSection, err := uctfirestore.SectionFromBytes(oldValue.Data)
	if err != nil {
		hermes.logger.WithError(err).Fatalf("firestore: failed to unmarshal model.Section")
		return err
	}

	value, err := uctfirestore.FromFirestoreValue(hermes.event.Value)
	if err != nil {
		log.WithError(err).Errorln("failed to parse new FirestoreValue from event")
		return err
	}

	newSection, err := uctfirestore.SectionFromBytes(value.Data)
	if err != nil {
		hermes.logger.WithError(err).Fatalf("failed to unmarshal model.Section")
		return err
	}

	if oldSection.Status == newSection.Status {
		log.WithError(err).
			WithField("old_section", oldSection.String()).
			WithField("new_section", newSection.String()).
			WithField("section", newSection.TopicName).
			WithField("event", fmt.Sprintf("%+v", hermes.event)).
			Warningln("section status did not update")
		return nil
	} else {
		log.WithField("old_status", oldSection.Status).
			WithField("new_status", newSection.Status).
			WithField("event", fmt.Sprintf("%+v", hermes.event)).
			Debugf("firestore: document updated: %v", hermes.event.UpdateMask.FieldPaths)
	}

	sectionNotification, err := hermes.uctFSClient.GetSectionNotification(newSection.TopicName)
	if err != nil {
		hermes.logger.WithError(err).Fatalf("failed to get additional section data")
		return err
	}

	err = hermes.sendNotification(sectionNotification)
	if err != nil {
		hermes.logger.WithError(err).Fatalf("failed to send notification")
		return err
	}

	err = hermes.uctFSClient.InsertSectionNotification(sectionNotification)
	if err != nil {
		hermes.logger.WithError(err).Fatalf("failed to insert notification")
		return err
	}

	return err
}
