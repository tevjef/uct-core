package main

import (
	"context"
	"io"
	_ "net/http/pprof"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/conf"
	"github.com/tevjef/uct-backend/common/model"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
}

func main() {
	rconf := &rutgersConfig{}

	app := kingpin.New("rutgers", "A web scraper that retrives course information for Rutgers University's servers.")

	app.Flag("campus", "Choose campus code. NB=New Brunswick, CM=Camden, NK=Newark").
		Short('u').
		PlaceHolder("[CM, NK, NB]").
		Required().
		Envar("RUTGERS_CAMPUS").
		EnumVar(&rconf.campus, "CM", "NK", "NB")

	app.Flag("format", "choose output format").
		Short('f').
		HintOptions(model.Json, model.Protobuf).
		PlaceHolder("[protobuf, json]").
		Default("protobuf").
		Envar("RUTGERS_OUTPUT_FORMAT").
		EnumVar(&rconf.outputFormat, "protobuf", "json")

	app.Flag("latest", "Only output the current and next semester").
		Short('l').
		Envar("RUTGERS_LATEST").
		BoolVar(&rconf.latest)

	configFile := app.Flag("config", "configuration file for the application").
		Required().
		Short('c').
		Envar("RUTGERS_CONFIG").
		File()

	kingpin.MustParse(app.Parse(os.Args[1:]))
	app.Name = app.Name + "-" + rconf.campus

	// Parse configuration file
	rconf.service = conf.OpenConfigWithName(*configFile, app.Name)

	// Start profiling
	go model.StartPprof(rconf.service.DebugSever(app.Name))

	(&rutgers{
		app:    app.Model(),
		config: rconf,
		ctx:    context.TODO(),
	}).init()
}

func (rutgers *rutgers) init() {
	if reader, err := model.MarshalMessage(rutgers.config.outputFormat, rutgers.getCampus(rutgers.config.campus)); err != nil {
		log.WithError(err).Fatal()
	} else {
		io.Copy(os.Stdout, reader)
	}
}
