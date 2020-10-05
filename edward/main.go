package main

import (
	"context"
	_ "net/http/pprof"
	"os"
	"strconv"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/conf"
	"github.com/tevjef/uct-backend/common/database"
	_ "github.com/tevjef/uct-backend/common/metrics"
	"github.com/tevjef/uct-backend/common/middleware"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/edward/store"
	"google.golang.org/api/option"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
// prometheus
)

type edward struct {
	app       *kingpin.ApplicationModel
	config    *edwardConfig
	firestore *firestore.Client
	postgres  database.Handler
	ctx       context.Context
}

type edwardConfig struct {
	service             conf.Config
	dryRun              bool
	firebaseProjectID   string
	gcpProject          string
	credentialsLocation string
	port                uint16
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)

	prometheus.MustRegister(
	// prometheus
	)
}

func main() {
	hconf := &edwardConfig{}

	app := kingpin.New("edward", "A server that listens to a database for events and publishes notifications to Firebase Cloud Messaging")

	app.Flag("dry-run", "enable dry-run").
		Short('d').
		Default("true").
		Envar("ED_DRY_RUN").
		BoolVar(&hconf.dryRun)

	app.Flag("listen", "port to start server on").
		Short('l').
		Default("2058").
		Envar("ED_LISTEN").
		Uint16Var(&hconf.port)

	app.Flag("firebase-project-id", "Firebase project Id").
		Default("universitycoursetracker").
		Envar("FIREBASE_PROJECT_ID").
		StringVar(&hconf.firebaseProjectID)

	app.Flag("project", "Google Cloud Platform project").
		Short('p').
		Envar("ED_GCP_PROJECT").
		StringVar(&hconf.gcpProject)

	app.Flag("google-credentials", "Google credentials location").
		Default("universitycoursetracker.json").
		Envar("CREDENTIALS_LOCATION").
		StringVar(&hconf.credentialsLocation)

	configFile := app.Flag("config", "configuration file for the application").
		Short('c').
		Envar("ED_CONFIG").
		File()

	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Parse configuration file
	hconf.service = conf.OpenConfigWithName(*configFile, app.Name)

	if hconf.dryRun {
		log.Infoln("Enabling FCM in dry run mode")
	} else {
		log.Infoln("Enabling FCM in production mode")
	}

	// Open database connection
	pgDatabase, err := model.OpenPostgres(hconf.service.DatabaseConfig(app.Name))
	if err != nil {
		log.WithError(err).Fatalln("failed to open database connection")
	}

	ctx := context.Background()

	opt := option.WithCredentialsFile(hconf.credentialsLocation)
	firebaseConf := &firebase.Config{ProjectID: hconf.firebaseProjectID}
	firebaseApp, err := firebase.NewApp(ctx, firebaseConf, opt)
	if err != nil {
		log.WithError(err).Fatalln("failed to crate firebase app")
	}

	fireStoreClient, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.WithError(err).Fatalln("failed to create firestore app")
	}

	// Start profiling
	go model.StartPprof(hconf.service.DebugSever(app.Name))

	(&edward{
		app:       app.Model(),
		config:    hconf,
		postgres:  database.NewHandler(app.Name, pgDatabase, store.Queries),
		firestore: fireStoreClient,
	}).init()
}

func (edward *edward) init() {
	// recovery and logging
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Ginrus())
	r.Use(middleware.Database(edward.postgres))

	if edward.config.gcpProject != "" {
		//traceClient, err := trace.NewClient(edward.ctx, edward.config.gcpProject)
		//if err != nil {
		//	log.Fatalf("failed to create client: %v", err)
		//}
		//r.Use(mtrace.Trace(traceClient))
	}

	v1 := r.Group("/v1")
	{
		v1.Use(middleware.ContentNegotiation(middleware.JsonContentType))
		v1.Use(middleware.Decorator)

		v1.GET("/hotness/:topic", courseHandler)
	}

	err := r.Run(":" + strconv.Itoa(int(edward.config.port)))
	if err != nil {
		log.WithError(err).Fatalln("failed to run server")
	}
}
