package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/conf"
	"github.com/tevjef/uct-backend/common/model"
)

var (
	subjectsHtml, _ = ioutil.ReadFile(conf.WorkingDir + "testdata/subjects.dat")
	searchHtml, _   = ioutil.ReadFile(conf.WorkingDir + "testdata/courses.dat")
	sectionHtml, _  = ioutil.ReadFile(conf.WorkingDir + "testdata/section.dat")
	defaultSemester = model.Semester{Year: 2017, Season: "spring"}
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
}

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
		if got, err := parseSemester(tt.args.semester); got != tt.want && err != nil {
			t.Errorf("%q. parseSemester() = %v, want %v err=%v", tt.name, got, tt.want, err)
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

	doc, err := sr.scrapeSubjects()
	if err != nil {
		t.Fatal(err.Error())
	}

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

	cr := courseScraper{semester: defaultSemester, url: ts.URL + "/search", scraper: defaultScraper}
	doc, err := cr.scrapeCourses()
	if err != nil {
		t.Fatal(err.Error())
	}
	_ = cr.parseCourses(doc)
}

func TestParseCourse(t *testing.T) {
	mux := setUpTestMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	cr := &courseScraper{semester: defaultSemester, url: ts.URL + "/search", scraper: defaultScraper}

	doc, err := cr.scrapeCourses()
	if err != nil {
		t.Fatal(err.Error())
	}

	findCourses(doc.Selection, func(index int, s *goquery.Selection) {
		findSections(s, func(index int, s *goquery.Selection) {
			section := cr.parseSection(s, &model.Course{})

			if section.CallNumber == "" {
				t.Fatalf("section.CallNumber is empty: %#v", section)
			}

			if section.Status == "" {
				t.Fatalf("section.Status is empty: %#v", section)
			}

			if section.Number == "" {
				t.Fatalf("section.Number is empty: %#v", section)
			}

			if section.Credits == "" {
				t.Fatalf("section section.Credits is empty: %#v", section)
			}

			if len(section.Metadata) < 1 && section.Metadata != nil {
				t.Fatalf("invalid section.Metadata:%#v", section)
			}
		})
	})
}

func TestParseSection(t *testing.T) {
	mux := setUpTestMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	sr := sectionScraper{url: ts.URL + "/section", scraper: defaultScraper}

	doc := sr.scrapeSection("")

	section := sr.parseAdditionalSectionInfo(doc, &model.Course{})

	if section.Credits == "" {
		t.Fatalf("section.Credits is empty: %#v", section)
	}

	if len(section.Metadata) < 1 {
		t.Fatalf("invalid section.Metadata: %#v", section)
	}
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
