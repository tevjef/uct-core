package main

import (
	"bufio"
	"bytes"
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

	filteredUniversity := DiffAndFilter(oldUniversity, newUniversity)

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

func DiffAndFilter(uni, uni2 uct.University) (filteredUniversity uct.University) {
	if UniEqual(&uni, uni2) {
		uni2.MainColor = ""
		uni2.AccentColor = ""
		uni2.HomePage = ""
		uni2.RegistrationPage = ""
		uni2.AvailableSemesters = []uct.Semester{}
		uni2.Registrations = []uct.Registration{}
		uni2.Metadata = []uct.Metadata{}
	}

	filteredUniversity = uni2
	oldSubjects := uni.Subjects
	newSubjects := uni2.Subjects
	var filteredSubjects []uct.Subject
	// For each newer subject

	for s := range newSubjects {
		Log(fmt.Sprintf("%s Subject index: %d \t %s | %s %s\n", oldSubjects[s].Season, s, oldSubjects[s].Name, newSubjects[s].Name, oldSubjects[s].Number == newSubjects[s].Number))
		// If current index is out of range of the old subjects[] break and add every subject
		if s >= len(oldSubjects) {
			filteredSubjects = newSubjects
			break
		}
		// If newSubject != oldSubject, log why, then drill deeper to filter out ones tht haven't changed
		if err := newSubjects[s].VerboseEqual(oldSubjects[s]); err != nil {
			Log("Subject differs!")
			oldCourses := oldSubjects[s].Courses
			newCourses := newSubjects[s].Courses
			var filteredCourses []uct.Course
			for c := range newCourses {
				//Log(fmt.Sprintf("Course index: %d \t %s | %s %s\n", c, oldCourses[c].Name, newCourses[c].Name, oldCourses[c].Number == newCourses[c].Number))
				if oldCourses[c].Number != newCourses[c].Number {
					fmt.Printf("%s %s", oldCourses[c].Name, newCourses[c].Name)
				}

				if c >= len(oldCourses) {
					filteredCourses = newCourses
					break
				}
				if err := newCourses[c].VerboseEqual(oldCourses[c]); err != nil {
					Log("Course differs!")
					oldSections := oldCourses[c].Sections
					newSections := newCourses[c].Sections
					var filteredSections []uct.Section
					for e := range newSections {
						//Log(fmt.Sprintf("Section index: %d \t %s | %s %s\n", e, oldSections[e].Number, newSections[e].Number, oldSections[e].Number == newSections[e].Number))
						if e >= len(oldSections) {
							filteredSections = newSections
							break
						}
						if err := newSections[e].VerboseEqual(oldSections[e]); err != nil {
							Log("Section: ", err)
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

func UniEqual(this *uct.University, that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*uct.University)
	if !ok {
		that2, ok := that.(uct.University)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if this.Id != that1.Id {
		return false
	}
	if this.Name != that1.Name {
		return false
	}
	if this.Abbr != that1.Abbr {
		return false
	}
	if this.HomePage != that1.HomePage {
		return false
	}
	if this.RegistrationPage != that1.RegistrationPage {
		return false
	}
	if this.MainColor != that1.MainColor {
		return false
	}
	if this.AccentColor != that1.AccentColor {
		return false
	}
	if this.TopicName != that1.TopicName {
		return false
	}
	if len(this.AvailableSemesters) != len(that1.AvailableSemesters) {
		return false
	}
	for i := range this.AvailableSemesters {
		if !this.AvailableSemesters[i].Equal(&that1.AvailableSemesters[i]) {
			return false
		}
	}
	if len(this.Registrations) != len(that1.Registrations) {
		return false
	}
	for i := range this.Registrations {
		if !this.Registrations[i].Equal(&that1.Registrations[i]) {
			return false
		}
	}
	if len(this.Metadata) != len(that1.Metadata) {
		return false
	}
	for i := range this.Metadata {
		if !this.Metadata[i].Equal(&that1.Metadata[i]) {
			return false
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}

func Log(v ...interface{}) {
	if *verbose {
		uct.Log(v)
	}
}
