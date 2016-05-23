package main

import (
	"bufio"
	"github.com/gogo/protobuf/proto"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"log"
	"os"
	uct "uct/common"
)

var (
	app    = kingpin.New("print", "An application to print and translate json and protobuf")
	format = app.Flag("format", "choose file input format").Short('f').HintOptions("protobuf", "json").PlaceHolder("[protobuf, json]").Required().String()
	output = app.Flag("output", "output format").Short('o').HintOptions("protobuf", "json").PlaceHolder("[protobuf, json]").String()
	file   = app.Arg("input", "file to print").File()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *format != "json" && *format != "protobuf" {
		log.Fatalln("Invalid format:", *format)
	}

	var input bufio.Reader
	if *file != nil {
		input = bufio.NewReader(*file)
	} else {
		input = bufio.NewReader(os.Stdin)
	}

	var university uct.University
	if *format == "json" {
		io.Copy(os.Stdout, input)
	} else if *format == "protobuf" {
		uct.UnmarshallMessage(*format, input, &university)
		proto.MarshalTextString(university)
	}

	output := uct.MarshalMessage(*output, university)
	io.Copy(os.Stdout, output)
}
