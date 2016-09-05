package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/contrib/cache"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"gopkg.in/alecthomas/kingpin.v2"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"
	uct "uct/common"
	"uct/servers"
)

var (
	app      = kingpin.New("spike", "A command-line application to serve university course information")
	port     = app.Flag("port", "port to start server on").Short('o').Default("9876").Uint16()
	logLevel = app.Flag("log-level", "Log level").Short('l').Default("debug").String()
	configFile    = app.Flag("config", "configuration file for the application").Short('c').File()
	config = uct.Config{}
)

const CacheDuration = 10 * time.Second

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if lvl, err := log.ParseLevel(*logLevel); err != nil {
		log.WithField("loglevel", *logLevel).Fatal(err)
	} else {
		log.SetLevel(lvl)
	}

	config = uct.NewConfig(*configFile)

	// Start profiling
	go uct.StartPprof(config.GetDebugSever(app.Name))

	var err error

	// Open database connection
	database, err = uct.InitDB(config.GetDbConfig(app.Name))
	uct.CheckError(err)

	// Prepare database connections
	database.SetMaxOpenConns(config.Db.ConnMax)
	PrepareAllStmts()

	// Open cache
	store := cache.NewInMemoryStore(CacheDuration)
	// recovery and logging
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))

	// Json
	v1 := r.Group("/v1")
	v1.Use(servers.JsonWriter())
	v1.Use(servers.ErrorWriter())

	// Protocol Buffers
	v2 := r.Group("/v2")
	v2.Use(servers.ProtobufWriter())
	v2.Use(servers.ErrorWriter())

	// Json (caching)
	v3 := r.Group("/v3")
	v3.Use(servers.JsonWriter())
	v3.Use(servers.ErrorWriter())

	// Protocol Buffers (caching)
	v4 := r.Group("/v4")
	v4.Use(servers.ProtobufWriter())
	v4.Use(servers.ErrorWriter())

	v1.GET("/universities", universitiesHandler)
	v1.GET("/university/:topic", universityHandler)
	v1.GET("/subjects/:topic/:season/:year", subjectsHandler)
	v1.GET("/subject/:topic", subjectHandler)
	v1.GET("/courses/:topic", coursesHandler)
	v1.GET("/course/:topic", courseHandler)
	v1.GET("/section/:topic", sectionHandler)

	v2.GET("/universities", universitiesHandler)
	v2.GET("/university/:topic", universityHandler)
	v2.GET("/subjects/:topic/:season/:year", subjectsHandler)
	v2.GET("/subject/:topic", subjectHandler)
	v2.GET("/courses/:topic", coursesHandler)
	v2.GET("/course/:topic", courseHandler)
	v2.GET("/section/:topic", sectionHandler)

	v3.GET("/universities", cache.CachePage(store, time.Minute, universitiesHandler))
	v3.GET("/university/:topic", cache.CachePage(store, time.Minute, universityHandler))
	v3.GET("/subjects/:topic/:season/:year", cache.CachePage(store, time.Minute, subjectsHandler))
	v3.GET("/subject/:topic", cache.CachePage(store, CacheDuration, subjectHandler))
	v3.GET("/courses/:topic", cache.CachePage(store, CacheDuration, coursesHandler))
	v3.GET("/course/:topic", cache.CachePage(store, CacheDuration, courseHandler))
	v3.GET("/section/:topic", cache.CachePage(store, CacheDuration, sectionHandler))

	v4.GET("/universities", cache.CachePage(store, time.Minute, universitiesHandler))
	v4.GET("/university/:topic", cache.CachePage(store, time.Minute, universityHandler))
	v4.GET("/subjects/:topic/:season/:year", cache.CachePage(store, time.Minute, subjectsHandler))
	v4.GET("/subject/:topic", cache.CachePage(store, CacheDuration, subjectHandler))
	v4.GET("/courses/:topic", cache.CachePage(store, CacheDuration, coursesHandler))
	v4.GET("/course/:topic", cache.CachePage(store, CacheDuration, courseHandler))
	v4.GET("/section/:topic", cache.CachePage(store, CacheDuration, sectionHandler))

	r.Run(":" + strconv.Itoa(int(*port)))
}
