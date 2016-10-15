package main

import (
	"uct/servers"
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"strconv"
	"time"
	"uct/common/model"
)

var (
	app    = kingpin.New("faye", "A command-line application to serve university course information")
	port   = app.Flag("port", "port to start server on").Short('o').Default("9877").Uint16()
	server = app.Flag("pprof", "host:port to start profiling on").Short('p').Default(model.FAYE_DEBUG_SERVER).TCP()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Start profiling
	go model.StartPprof(*server)

	var err error

	// Open database connection
	database, err = model.InitDB(model.GetUniversityDB())
	model.CheckError(err)

	// Prepare database connections
	database.SetMaxOpenConns(50)
	PrepareAllStmts()

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



	v1.POST("/notification", notificationHandler)

	v2.POST("/notification", notificationHandler)

	r.Run(":" + strconv.Itoa(int(*port)))
}