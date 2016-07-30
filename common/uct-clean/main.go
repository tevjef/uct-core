package main

import (
	"bufio"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"os"
	uct "uct/common"
)

var (
	app      = kingpin.New("print", "An application to print and translate json and protobuf")
	logLevel = app.Flag("log-level", "Log level").Short('l').Default("debug").String()
	format   = app.Flag("format", "choose file input format").Short('f').HintOptions(uct.PROTOBUF, uct.JSON).PlaceHolder("[protobuf, json]").Required().String()
	out      = app.Flag("output", "output format").Short('o').HintOptions(uct.PROTOBUF, uct.JSON).PlaceHolder("[protobuf, json]").String()
	file     = app.Arg("input", "file to clean").File()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
	if lvl, err := log.ParseLevel(*logLevel); err != nil {
		log.WithField("loglevel", *logLevel).Fatal(err)
	} else {
		log.SetLevel(lvl)
	}
	if *format != uct.JSON && *format != uct.PROTOBUF {
		log.WithField("format", *format).Fatal("Invalid format")
	}

	var input *bufio.Reader
	if *file != nil {
		input = bufio.NewReader(*file)
	} else {
		input = bufio.NewReader(os.Stdin)
	}

	var university uct.University
	uct.UnmarshallMessage(*format, input, &university)

	uct.ValidateAll(&university)

	if *out != "" {
		output := uct.MarshalMessage(*out, university)
		io.Copy(os.Stdout, output)
	} else {
		output := uct.MarshalMessage(*format, university)
		io.Copy(os.Stdout, output)
	}
}
