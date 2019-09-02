package main

import (
	"context"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/trace"
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/tevjef/uct-backend/common/conf"
	"github.com/tevjef/uct-backend/common/database"
	_ "github.com/tevjef/uct-backend/common/metrics"
	"github.com/tevjef/uct-backend/common/middleware"
	"github.com/tevjef/uct-backend/common/middleware/cache"
	mtrace "github.com/tevjef/uct-backend/common/middleware/trace"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/common/redis"
	"github.com/tevjef/uct-backend/spike/store"
	"gopkg.in/alecthomas/kingpin.v2"
)

type spike struct {
	app      *kingpin.ApplicationModel
	config   *spikeConfig
	postgres database.Handler
	redis    *redis.Helper
	ctx      context.Context
}

type spikeConfig struct {
	service    conf.Config
	gcpProject string
	port       uint16
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
}

func main() {
	sconf := &spikeConfig{}

	app := kingpin.New("spike", "A command-line application to serve university course information")

	app.Flag("listen", "port to start server on").
		Short('l').
		Default("9876").
		Envar("SPIKE_LISTEN").
		Uint16Var(&sconf.port)

	app.Flag("project", "Google Cloud Platform project for tracing").
		Short('p').
		Envar("SPIKE_GCP_PROJECT").
		StringVar(&sconf.gcpProject)

	configFile := app.Flag("config", "configuration file for the application").
		Short('c').
		Envar("SPIKE_CONFIG").
		File()

	kingpin.MustParse(app.Parse(os.Args[1:]))

	sconf.service = conf.OpenConfigWithName(*configFile, app.Name)

	// Start profiling
	go model.StartPprof(sconf.service.DebugSever(app.Name))

	// Open database connection
	pgdb, err := model.OpenPostgres(sconf.service.DatabaseConfig(app.Name))
	if err != nil {
		log.WithError(err).Fatalln("failed to open connection to database")
	}
	pgdb.SetMaxOpenConns(sconf.service.Postgres.ConnMax)
	pgdb.SetMaxIdleConns(sconf.service.Postgres.ConnMax)

	(&spike{
		app:      app.Model(),
		config:   sconf,
		redis:    redis.NewHelper(sconf.service, app.Name),
		postgres: database.NewHandler(app.Name, pgdb, store.Queries),
		ctx:      context.Background(),
	}).init()
}

func (spike *spike) init() {
	// recovery and logging
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Ginrus())
	r.Use(middleware.Database(spike.postgres))
	r.Use(cache.Cache(cache.NewRedisCache(
		spike.config.service.RedisAddr(),
		spike.config.service.Redis.Password,
		spike.config.service.Spike.RedisDb,
		10*time.Second)))

	if spike.config.gcpProject != "" {
		traceClient, err := trace.NewClient(spike.ctx, spike.config.gcpProject)
		if err != nil {
			log.Fatalf("failed to create client: %v", err)
		}
		r.Use(mtrace.Trace(traceClient))
	}

	// does not cache and defaults to json
	v1 := r.Group("/v1")
	{
		v1.Use(middleware.ContentNegotiation(middleware.JsonContentType))
		v1.Use(middleware.Decorator)

		v1.GET("/universities", universitiesHandler(0))
		v1.GET("/university/:topic", universityHandler(0))
		v1.GET("/subjects/:topic/:season/:year", subjectsHandler(0))
		v1.GET("/subject/:topic", subjectHandler(0))
		v1.GET("/courses/:topic", coursesHandler(0))
		v1.GET("/course/:topic", courseHandler(0))

		v1.GET("/section/:topic", sectionHandler(0))
		v1.GET("/subscription", subscriptionHandler())
		v1.GET("/notification", notificationHandler())
	}

	// v2 caches responses and defaults to protobuf
	v2 := r.Group("/v2")
	{
		v2.Use(middleware.ContentNegotiation(middleware.ProtobufContentType))
		v2.Use(middleware.Decorator)

		v2.GET("/universities", universitiesHandler(time.Minute))
		v2.GET("/university/:topic", universityHandler(time.Minute))
		v2.GET("/subjects/:topic/:season/:year", subjectsHandler(time.Minute))
		v2.GET("/subject/:topic", subjectHandler(10*time.Second))
		v2.GET("/courses/:topic", coursesHandler(10*time.Second))
		v2.GET("/course/:topic", courseHandler(10*time.Second))
		v2.GET("/section/:topic", sectionHandler(10*time.Second))
		v2.POST("/subscription", subscriptionHandler())
		v2.POST("/notification", notificationHandler())
	}

	static := r.Group("/static")
	static.GET("/:file", serveStaticFromGithub)

	r.Run(":" + strconv.Itoa(int(spike.config.port)))
}
