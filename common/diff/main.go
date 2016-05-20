package main

import (
	"bufio"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/pquerna/ffjson/ffjson"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"log"
	"os"
	uct "uct/common"
)

var (
	app     = kingpin.New("model-diff", "An application to filter unchanged sections")
	format  = app.Flag("format", "choose input format").Short('f').HintOptions("protobuf", "json").PlaceHolder("[protobuf, json]").Required().String()
	first   = app.Arg("FILE", "the first file to compare").Required().File()
	second  = app.Arg("FILE2", "the second file to compare").File()
	verbose = app.Flag("verbose", "verbose log of object representations.").Short('v').Bool()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *format != "json" && *format != "protobuf" {
		log.Fatalln("Invalid format:", *format)
	}

	var firstFile = bufio.NewReader(*first)
	var secondFile *bufio.Reader

	if *second != nil {
		secondFile = bufio.NewReader(*second)
	} else if *second != nil {
		secondFile = bufio.NewReader(os.Stdin)
	}

	var university1 uct.University

	if *format == "json" {
		dec := ffjson.NewDecoder()
		if err := dec.DecodeReader(firstFile, &university1); err != nil {
			log.Fatalln("Failed to unmarshal first university:", err)
		}
	} else if *format == "protobuf" {
		data, err := ioutil.ReadAll(firstFile)
		if err = proto.Unmarshal(data, &university1); err != nil {
			log.Fatalln("Failed to unmarshal first university:", err)
		}
	}

	var university2 uct.University

	if *format == "json" {
		dec := ffjson.NewDecoder()
		if err := dec.DecodeReader(secondFile, &university2); err != nil {
			log.Fatalln("Failed to unmarshal second university", err)
		}
	} else if *format == "protobuf" {
		data, err := ioutil.ReadAll(secondFile)
		if err = proto.Unmarshal(data, &university2); err != nil {
			log.Fatalln("Failed to unmarshal second university:", err)
		}
	}
}

func diffAndFilter(uni, uni2 uct.University) {
	oldSubjects := uni.Subjects
	newSubjects := uni2.Subjects

	if len(newSubjects) != len(oldSubjects) {
		log.Println("Subjects:", len(uni.Subjects), "Subjects:", len(uni2))
	}

	for i := range newSubjects {
		newSubjects[i].VerboseEqual(oldSubjects[i])
	}

}
