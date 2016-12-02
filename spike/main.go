package main

import (
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"
	"uct/common/conf"
	"uct/common/database"
	"uct/common/model"
	"uct/common/redis"
	"uct/spike/middleware"
	"uct/spike/middleware/cache"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"golang.org/x/net/context"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type spike struct {
	app      *kingpin.ApplicationModel
	config   *spikeConfig
	postgres database.Handler
	redis    *redis.Helper
	ctx      context.Context
}

type spikeConfig struct {
	service conf.Config
	port    uint16
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
}

func main() {
	app := kingpin.New("spike", "A command-line application to serve university course information")
	port := app.Flag("port", "port to start server on").Short('o').Default("9876").Uint16()
	configFile := app.Flag("config", "configuration file for the application").Short('c').File()
	kingpin.MustParse(app.Parse(os.Args[1:]))

	config := conf.OpenConfigWithName(*configFile, app.Name)

	// Start profiling
	go model.StartPprof(config.DebugSever(app.Name))

	// Open database connection
	pgdb, err := model.OpenPostgres(config.DatabaseConfig(app.Name))
	if err != nil {
		log.WithError(err).Fatalln("failed to open connection to database")
	}
	pgdb.SetMaxOpenConns(config.Postgres.ConnMax)

	(&spike{
		app: app.Model(),
		config: &spikeConfig{
			service: config,
			port:    *port,
		},
		redis:    redis.NewHelper(config, app.Name),
		postgres: database.NewHandler(app.Name, pgdb, queries),
		ctx:      context.TODO(),
	}).init()
}

func (spike *spike) init() {
	// recovery and logging
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Ginrus(log.StandardLogger(), time.RFC3339, true))
	r.Use(middleware.Database(spike.postgres))
	r.Use(cache.Cache(cache.NewRedisCache(
		spike.config.service.RedisAddr(),
		spike.config.service.Redis.Password,
		spike.config.service.Spike.RedisDb,
		10*time.Second)))

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

	r.Run(":" + strconv.Itoa(int(spike.config.port)))
}
