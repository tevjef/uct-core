package main

import (
	"bufio"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"log"
	"os"
	uct "uct/common"
)

var (
	app     = kingpin.New("model-diff", "An application to filter unchanged objects")
	format  = app.Flag("format", "choose file input format").Short('f').HintOptions("protobuf", "json").PlaceHolder("[protobuf, json]").Required().String()
	old     = app.Arg("old", "the first file to compare").Required().File()
	new     = app.Arg("new", "the second file to compare").File()
	verbose = app.Flag("verbose", "Verbose log of object representations.").Short('v').Bool()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *format != "json" && *format != "protobuf" {
		log.Fatalln("Invalid format:", *format)
	}

	var firstFile = bufio.NewReader(*old)
	var secondFile *bufio.Reader

	if *new != nil {
		secondFile = bufio.NewReader(*new)
	} else {
		secondFile = bufio.NewReader(os.Stdin)
	}

	var oldUniversity uct.University

	uct.UnmarshallMessage(*format, firstFile, &oldUniversity)

	var newUniversity uct.University

	uct.UnmarshallMessage(*format, secondFile, &newUniversity)

	filteredUniversity := uct.DiffAndFilter(oldUniversity, newUniversity)

	buf := uct.MarshalMessage(*format, filteredUniversity)

	io.Copy(os.Stdout, buf)
}
