package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"gopkg.in/alecthomas/kingpin.v2"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"
	"uct/common/conf"
	"uct/common/model"
	"uct/servers/spike/cache"
	"uct/servers/spike/middleware"
)

var (
	app        = kingpin.New("spike", "A command-line application to serve university course information")
	port       = app.Flag("port", "port to start server on").Short('o').Default("9876").Uint16()
	logLevel   = app.Flag("log-level", "Log level").Short('l').Default("info").String()
	configFile = app.Flag("config", "configuration file for the application").Short('c').File()
	config     = conf.Config{}
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if lvl, err := log.ParseLevel(*logLevel); err != nil {
		log.WithField("loglevel", *logLevel).Fatal(err)
	} else {
		log.SetLevel(lvl)
	}

	config = conf.OpenConfig(*configFile)
	config.AppName = app.Name

	// Start profiling
	go model.StartPprof(config.GetDebugSever(app.Name))

	var err error

	// Open database connection
	if database, err = model.InitDB(config.GetDbConfig(app.Name)); err != nil {
		log.WithError(err).Fatalln("failed to open connection to database")
	}

	// Prepare database connections
	database.SetMaxOpenConns(config.Postgres.ConnMax)
	prepareAllStmts()

	// recovery and logging
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Ginrus(log.StandardLogger(), time.RFC3339, true))
	r.Use(cache.Cache(cache.NewRedisCache(config.GetRedisAddr(), config.Redis.Password, config.Spike.RedisDb, 10*time.Second)))

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
	}

	// Legacy, some version android and iOS clients use this endpoint. Investigate redirecting traffic to /v2 with nginx
	v4 := r.Group("/v4")
	{
		v4.Use(middleware.ContentNegotiation(middleware.ProtobufContentType))
		v4.Use(middleware.Decorator)

		v4.GET("/universities", universitiesHandler(time.Minute))
		v4.GET("/university/:topic", universityHandler(time.Minute))
		v4.GET("/subjects/:topic/:season/:year", subjectsHandler(time.Minute))
		v4.GET("/subject/:topic", subjectHandler(10*time.Second))
		v4.GET("/courses/:topic", coursesHandler(10*time.Second))
		v4.GET("/course/:topic", courseHandler(10*time.Second))
		v4.GET("/section/:topic", sectionHandler(10*time.Second))
	}

	r.Run(":" + strconv.Itoa(int(*port)))
}
