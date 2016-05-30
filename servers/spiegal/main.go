package main

import (
	_ "github.com/lib/pq"
	"github.com/tevjef/gin"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"strconv"
	uct "uct/common"
)

var (
	app    = kingpin.New("spiegal", "A command-line application to serve university course information")
	port   = app.Flag("port", "port to start server on").Short('o').Default("9876").Uint16()
	server = app.Flag("pprof", "host:port to start profiling on").Short('p').Default(uct.SPIEGAL_DEBUG_SERVER).TCP()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Start profiling
	go uct.StartPprof(*server)

	var err error
	database, err = uct.InitDB(uct.GetUniversityDB())
	uct.CheckError(err)

	database.SetMaxOpenConns(50)

	PrepareAllStmts()

	r := gin.Default()
	r.Use(errorWriter())
	v1 := r.Group("/v1")
	v1.Use(jsonWriter())
	v2 := r.Group("/v2")
	v2.Use(protobufWriter())

	v1.GET("/university/:topic", universityHandler)
	v1.GET("/subject/:topic", subjectHandler)
	v1.GET("/course/:topic", courseHandler)
	v1.GET("/section/:topic", sectionHandler)
	v1.GET("/universities", universitiesHandler)
	v1.GET("/subjects/:topic/:season/:year", subjectsHandler)
	v1.GET("/courses/:topic", coursesHandler)

	v2.GET("/university/:topic", universityHandler)
	v2.GET("/subject/:topic", subjectHandler)
	v2.GET("/course/:topic", courseHandler)
	v2.GET("/section/:topic", sectionHandler)
	v2.GET("/universities", universitiesHandler)
	v2.GET("/subjects/:topic/:season/:year", subjectsHandler)
	v2.GET("/courses/:topic", coursesHandler)

	r.Run(":" + strconv.Itoa(int(*port)))
}
