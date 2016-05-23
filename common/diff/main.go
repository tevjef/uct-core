package main

import (
	"bufio"
	"github.com/gogo/protobuf/proto"
	"github.com/pquerna/ffjson/ffjson"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
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

	if *format == "json" {
		dec := ffjson.NewDecoder()
		if err := dec.DecodeReader(firstFile, &oldUniversity); err != nil {
			log.Fatalln("Failed to unmarshal first university:", err)
		}
	} else if *format == "protobuf" {
		data, err := ioutil.ReadAll(firstFile)
		if err = proto.Unmarshal(data, &oldUniversity); err != nil {
			log.Fatalln("Failed to unmarshal first university:", err)
		}
	}

	var newUniversity uct.University

	if *format == "json" {
		dec := ffjson.NewDecoder()
		if err := dec.DecodeReader(secondFile, &newUniversity); err != nil {
			log.Fatalln("Failed to unmarshal second university", err)
		}
	} else if *format == "protobuf" {
		data, err := ioutil.ReadAll(secondFile)
		if err = proto.Unmarshal(data, &newUniversity); err != nil {
			log.Fatalln("Failed to unmarshal second university:", err)
		}
	}

	filteredUniversity := uct.DiffAndFilter(oldUniversity, newUniversity)

	if *format == "json" {
		enc := ffjson.NewEncoder(os.Stdout)
		err := enc.Encode(filteredUniversity)
		uct.CheckError(err)
	} else if *format == "protobuf" {
		out, err := proto.Marshal(&filteredUniversity)
		if err != nil {
			log.Fatalln("Failed to encode university:", err)
		}
		if _, err := os.Stdout.Write(out); err != nil {
			log.Fatalln("Failed to write university:", err)
		}
	}
}

func Log(v ...interface{}) {
	if *verbose {
		uct.Log(v)
	}
}
