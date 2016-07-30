package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/contrib/cache"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"strconv"
	"time"
	uct "uct/common"
)

var (
	app    = kingpin.New("spiegal", "A command-line application to serve university course information")
	port   = app.Flag("port", "port to start server on").Short('o').Default("9876").Uint16()
	server = app.Flag("pprof", "host:port to start profiling on").Short('p').Default(uct.SPIEGAL_DEBUG_SERVER).TCP()
)

const CacheDuration = 10 * time.Second

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Start profiling
	go uct.StartPprof(*server)

	var err error

	// Open database connection
	database, err = uct.InitDB(uct.GetUniversityDB())
	uct.CheckError(err)

	// Prepare database connections
	database.SetMaxOpenConns(50)
	PrepareAllStmts()

	// Open cache
	store := cache.NewInMemoryStore(CacheDuration)
	// recovery and logging
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, true))

	// Json
	v1 := r.Group("/v1")
	v1.Use(jsonWriter())
	v1.Use(errorWriter())

	// Protocol Buffers
	v2 := r.Group("/v2")
	v2.Use(protobufWriter())
	v2.Use(errorWriter())

	// Json (caching)
	v3 := r.Group("/v3")
	v3.Use(jsonWriter())
	v3.Use(errorWriter())

	// Protocol Buffers (caching)
	v4 := r.Group("/v4")
	v4.Use(protobufWriter())
	v4.Use(errorWriter())

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
