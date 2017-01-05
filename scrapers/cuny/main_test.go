package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/uct-core/common/conf"
	"github.com/tevjef/uct-core/common/model"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
}

var subjectsHtml, _ = ioutil.ReadFile(conf.WorkingDir + "testdata/subjects.dat")
var searchHtml, _ = ioutil.ReadFile(conf.WorkingDir + "testdata/courses.dat")
var sectionHtml, _ = ioutil.ReadFile(conf.WorkingDir + "testdata/section.dat")

func TestParseSemester(t *testing.T) {
	type args struct {
		semester model.Semester
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{args: args{model.Semester{Year: 2016, Season: model.Spring}}, want: "1162"},
		{args: args{model.Semester{Year: 2015, Season: model.Spring}}, want: "1152"},
		{args: args{model.Semester{Year: 2017, Season: model.Fall}}, want: "1179"},
		{args: args{model.Semester{Year: 2017, Season: model.Summer}}, want: "1177"},
		{args: args{model.Semester{Year: 2017, Season: model.Winter}}, want: "1170"},
	}
	for _, tt := range tests {
		if got := parseSemester(tt.args.semester); got != tt.want {
			t.Errorf("%q. parseSemester() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestParseSubjects(t *testing.T) {
	mux := setUpTestMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	sr := &subjectScraper{
		scraper: defaultScraper,
		url:     ts.URL + "/subjects",
		semester: model.Semester{
			Year:   2017,
			Season: model.Spring,
		}}

	doc := sr.scrapeSubjects()

	subjects := sr.parseSubjects(doc)

	if len(subjects) == 0 {
		t.Error("number of subjects found is zero")
	}

	for _, val := range subjects {
		if val.Name == "" {
			t.Error("subject did not have a name")
		}

		if val.Number == "" {
			t.Error("subject did not have a numvber")
		}
	}
}

func TestParseCourses(t *testing.T) {
	mux := setUpTestMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	cr := courseScraper{url: ts.URL + "/search", scraper: defaultScraper}
	doc := cr.scrapeCourses()

	_ = cr.parseCourses(doc)
}

func TestParseCourse(t *testing.T) {
	mux := setUpTestMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	cr := &courseScraper{url: ts.URL + "/search", scraper: defaultScraper}

	doc := cr.scrapeCourses()

	findCourses(doc.Selection, func(index int, s *goquery.Selection) {
		findSections(s, func(index int, s *goquery.Selection) {
			a, _ := json.Marshal(cr.parseSection(s, &model.Course{}))
			fmt.Println(string(a))
		})
	})
}

func TestParseSection(t *testing.T) {
	mux := setUpTestMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	sr := sectionScraper{url: ts.URL + "/section", scraper: defaultScraper}

	doc := sr.scrapeSection("")

	fmt.Printf("%#v", sr.parseSection(doc, &model.Course{}))
}

func setUpTestMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/subjects", func(w http.ResponseWriter, req *http.Request) {
		w.Write(subjectsHtml)
	})
	mux.HandleFunc("/search", func(w http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.FormValue("ICAction"), "MTG_CLASSNAME") {
			w.Write(sectionHtml)
			return
		}
		w.Write(searchHtml)
	})
	mux.HandleFunc("/section", func(w http.ResponseWriter, req *http.Request) {
		w.Write(sectionHtml)
	})
	return mux
}
