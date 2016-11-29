package main

import (
	"bufio"
	"io"
	"os"
	"uct/common/model"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	log "github.com/Sirupsen/logrus"
)

var (
	app      = kingpin.New("print", "An application to print and translate json and protobuf")
	logLevel = app.Flag("log-level", "Log level").Short('l').Default("debug").String()
	format   = app.Flag("format", "choose file input format").Short('f').HintOptions(model.PROTOBUF, model.JSON).PlaceHolder("[protobuf, json]").Required().String()
	out      = app.Flag("output", "output format").Short('o').HintOptions(model.PROTOBUF, model.JSON).PlaceHolder("[protobuf, json]").String()
	file     = app.Arg("input", "file to clean").File()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
	if lvl, err := log.ParseLevel(*logLevel); err != nil {
		log.WithField("loglevel", *logLevel).Fatal(err)
	} else {
		log.SetLevel(lvl)
	}
	if *format != model.JSON && *format != model.PROTOBUF {
		log.WithField("format", *format).Fatal("Invalid format")
	}

	var input *bufio.Reader
	if *file != nil {
		input = bufio.NewReader(*file)
	} else {
		input = bufio.NewReader(os.Stdin)
	}

	var university model.University

	if err := model.UnmarshalMessage(*format, input, &university); err != nil {
		log.WithError(err).Fatalf("Failed to unmarshall message")
	}

	if err := model.ValidateAll(&university); err != nil {
		log.WithError(err).Fatalf("Failed to validate message")
	}

	if *out != "" {
		if output, err := model.MarshalMessage(*out, university); err != nil {
			log.WithError(err).Fatal()
		} else {
			io.Copy(os.Stdout, output)
		}
	} else {
		if output, err := model.MarshalMessage(*format, university); err != nil {
			log.WithError(err).Fatal()
		} else {
			io.Copy(os.Stdout, output)
		}
	}
}
