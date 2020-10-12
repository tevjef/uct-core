package main

import (
	"bufio"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/model"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	app    = kingpin.New("print", "An application to print and translate json and protobuf")
	format = app.Flag("format", "choose file input format").Short('f').HintOptions(model.Protobuf, model.Json).PlaceHolder("[protobuf, json]").Required().String()
	out    = app.Flag("output", "output format").Short('o').
		HintOptions(model.Protobuf, model.Json).
		Default("json").
		PlaceHolder("[protobuf, json]").String()
	file = app.Arg("input", "file to print").File()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *format != model.Json && *format != model.Protobuf {
		log.Fatalln("Invalid format:", *format)
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

	if *out != "" {
		if output, err := model.MarshalMessage(*out, university); err != nil {
			log.WithError(err).Fatal()
		} else {
			io.Copy(os.Stdout, output)
		}
	}
}
