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
	app     = kingpin.New("model-diff", "An application to filter unchanged sections")
	format  = app.Flag("format", "choose input format").Short('f').HintOptions("protobuf", "json").PlaceHolder("[protobuf, json]").Required().String()
	old = app.Arg("old", "the first file to compare").Required().File()
	new = app.Arg("new", "the second file to compare").File()
	verbose = app.Flag("verbose", "verbose log of object representations.").Short('v').Bool()
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
	} else if *new != nil {
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

	var filteredUniversity uct.University

	filteredUniversity = diffAndFilter(oldUniversity, newUniversity)

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


func diffAndFilter(uni, uni2 uct.University) (filteredUniversity uct.University) {
	filteredUniversity = uni2
	oldSubjects := uni.Subjects
	newSubjects := uni2.Subjects
	var filteredSubjects []uct.Subject
	for s := range newSubjects {
		if err := newSubjects[s].VerboseEqual(oldSubjects[s]); err != nil {
			uct.Log(err)
			oldCourses := oldSubjects[s].Courses
			newCourses := newSubjects[s].Courses
			var filteredCourses []uct.Course
			for c := range newCourses {
				if err := newCourses[c].VerboseEqual(oldCourses[c]); err != nil {
					uct.Log(err)
					oldSections := oldCourses[c].Sections
					newSections := newCourses[c].Sections
					var filteredSections []uct.Section
					for e := range newSections {
						if err := newSections[e].VerboseEqual(oldSections[e]); err != nil {
							uct.Log(err)
							filteredSections = append(filteredSections, newSections[e])
						}
					}
					newCourses[c].Sections = filteredSections
					filteredCourses = append(filteredCourses, newCourses[c])
				}
			}
			newSubjects[s].Courses = filteredCourses
			filteredSubjects = append(filteredSubjects, newSubjects[s])
		}
	}
	filteredUniversity.Subjects = filteredSubjects
	return
}
