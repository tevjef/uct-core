package main

import (
	"bufio"
	"io"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/gogo/protobuf/proto"
	"github.com/tevjef/uct-backend/common/model"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	app    = kingpin.New("push", "An application to print and translate json and protobuf")
	format = app.Flag("format", "choose file input format").Short('f').HintOptions(model.Protobuf, model.Json).PlaceHolder("[protobuf, json]").Required().String()
	out    = app.Flag("output", "output format").Short('o').HintOptions(model.Protobuf, model.Json).PlaceHolder("[protobuf, json]").String()
	file   = app.Arg("input", "file to print").File()
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

	if *format == model.Json {
		if *out != "" {
			io.Copy(os.Stdout, input)
		}
	} else if *format == model.Protobuf {
		if *out != "" {
			log.Println(proto.MarshalTextString(&university))
		}
	}

	if *out != "" {
		output, err := model.MarshalMessage(*out, university)
		if err != nil {
			panic(err)
		}
		io.Copy(os.Stdout, output)
	} else {
		output, err := model.MarshalMessage(*format, university)
		if err != nil {
			panic(err)
		}
		io.Copy(os.Stdout, output)
	}
}
