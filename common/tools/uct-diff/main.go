package main

import (
	"bufio"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"log"
	"os"
	"uct/common/model"
)

var (
	app     = kingpin.New("model-diff", "An application to filter unchanged objects")
	format  = app.Flag("format", "choose file input format").Short('f').HintOptions(model.PROTOBUF, model.JSON).PlaceHolder("[protobuf, json]").Required().String()
	old     = app.Arg("old", "the first file to compare").Required().File()
	new     = app.Arg("new", "the second file to compare").File()
	verbose = app.Flag("verbose", "Verbose log of object representations.").Short('v').Bool()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *format != model.JSON && *format != model.PROTOBUF {
		log.Fatalln("Invalid format:", *format)
	}

	var firstFile = bufio.NewReader(*old)
	var secondFile *bufio.Reader

	if *new != nil {
		secondFile = bufio.NewReader(*new)
	} else {
		secondFile = bufio.NewReader(os.Stdin)
	}

	var oldUniversity model.University

	model.UnmarshallMessage(*format, firstFile, &oldUniversity)

	var newUniversity model.University

	model.UnmarshallMessage(*format, secondFile, &newUniversity)

	filteredUniversity := model.DiffAndFilter(oldUniversity, newUniversity)

	buf := model.MarshalMessage(*format, filteredUniversity)

	io.Copy(os.Stdout, buf)
}
