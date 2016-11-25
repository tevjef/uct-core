package main

import (
	"bufio"
	log "github.com/Sirupsen/logrus"
	"github.com/gogo/protobuf/proto"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"os"
	"uct/common/model"
)

var (
	app    = kingpin.New("push", "An application to print and translate json and protobuf")
	format = app.Flag("format", "choose file input format").Short('f').HintOptions(model.PROTOBUF, model.JSON).PlaceHolder("[protobuf, json]").Required().String()
	out    = app.Flag("output", "output format").Short('o').HintOptions(model.PROTOBUF, model.JSON).PlaceHolder("[protobuf, json]").String()
	file   = app.Arg("input", "file to print").File()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *format != model.JSON && *format != model.PROTOBUF {
		log.Fatalln("Invalid format:", *format)
	}

	var input *bufio.Reader
	if *file != nil {
		input = bufio.NewReader(*file)
	} else {
		input = bufio.NewReader(os.Stdin)
	}

	var university model.University

	if err := model.UnmarshallMessage(*format, input, &university); err != nil {
		log.WithError(err).Fatalf("Failed to unmarshall message")
	}

	if *format == model.JSON {
		if *out != "" {
			io.Copy(os.Stdout, input)
		}
	} else if *format == model.PROTOBUF {
		if *out != "" {
			log.Println(proto.MarshalTextString(&university))
		}
	}

	if *out != "" {
		output := model.MarshalMessage(*out, university)
		io.Copy(os.Stdout, output)
	} else {
		output := model.MarshalMessage(*format, university)
		io.Copy(os.Stdout, output)
	}
}
