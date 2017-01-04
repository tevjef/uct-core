package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/uct-core/common/conf"
	"github.com/tevjef/uct-core/common/model"
	"github.com/tevjef/uct-core/common/proxy"
	"golang.org/x/net/context"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type cuny struct {
	app    *kingpin.ApplicationModel
	config *cunyConfig
	ctx    context.Context
}

type cunyConfig struct {
	service      conf.Config
	university   CunyUniversity
	full         bool
	outputFormat string
	latest       bool
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
}

func main() {
	app := kingpin.New("cuny", "A program for scraping information from CunyFirst.")

	abbr := app.Flag("university", "Choose a cuny university"+abbrMap()).
		Short('u').
		Required().
		String()

	section := app.Flag("section", "Get all section data").
		Short('s').
		Envar("CUNY_ALL_SECTION").
		Bool()

	format := app.Flag("format", "Choose output format").
		Short('f').
		HintOptions(model.Json, model.Protobuf).
		PlaceHolder("[protobuf, json]").
		Default("protobuf").
		String()

	configFile := app.Flag("config", "Configuration file for the application").
		Required().
		Short('c').
		File()

	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Parse configuration file
	config := conf.OpenConfigWithName(*configFile, app.Name)

	// Start profiling
	go model.StartPprof(config.DebugSever(app.Name))

	upperAbbr := strings.ToUpper(*abbr)

	(&cuny{
		app: app.Model(),
		config: &cunyConfig{
			service:      config,
			full:         *section,
			outputFormat: *format,
			university:   abbrToCunyUniversity(upperAbbr),
		},
		ctx: context.TODO(),
	}).init()
}

func (cuny *cuny) init() {
	cs := &cunyScraper{
		university: cuny.config.university,
		client: &CunyFirstClient{
			values:     map[string][]string{},
			httpClient: newClient(),
		},
		full: cuny.config.full,
	}

	sr := &subjectScraper{
		scraper: cs,
		url:     initalPage,
		semester: model.Semester{
			Year:   2017,
			Season: model.Spring,
		},
	}

	subjects := sr.parseSubjects(sr.scrapeSubjects())

	var wg sync.WaitGroup

	sem := make(chan *sync.WaitGroup, 20)

	wg.Add(len(subjects))
	for i := range subjects {
		sem <- &wg
		go func(subject *model.Subject) {
			cs.run(subject)
			wg := <-sem
			wg.Done()
		}(subjects[i])
	}

	wg.Wait()

	university := model.University{}
	university.Subjects = subjects

	if reader, err := model.MarshalMessage(cuny.config.outputFormat, university); err != nil {
		log.WithError(err).Fatalln()
	} else {
		io.Copy(os.Stdout, reader)
	}
}

type cunyScraper struct {
	client     *CunyFirstClient
	university CunyUniversity
	full       bool
}

func (cs *cunyScraper) run(subject *model.Subject) {
	httpClient := newClient()

	newMap := map[string][]string{}
	for k, v := range cs.client.values {
		newMap[k] = v
	}

	scraper := &cunyScraper{
		university: cs.university,
		client: &CunyFirstClient{
			values:     newMap,
			httpClient: httpClient,
		},
		full: cs.full,
	}

	scraper.client.Get(initalPage)

	sr := &subjectScraper{
		scraper: scraper,
		url:     initalPage,
		semester: model.Semester{
			Year:   2017,
			Season: model.Spring,
		},
		full: cs.full,
	}

	sr.scrapeSubjects()

	cr := courseScraper{scraper, initalPage, subject.Number, cs.full}
	doc := cr.scrapeCourses()

	if doc == nil {
		log.Warnln("skipping courses for", subject.Name)
	} else {
		subject.Courses = cr.parseCourses(doc)
	}

}

type subjectScraper struct {
	scraper  *cunyScraper
	url      string
	semester model.Semester
	full     bool
}

func (sr *subjectScraper) scrapeSubjects() *goquery.Document {
	sr.scraper.client.Get(initalPage)

	cunyForm := cunyForm{}
	cunyForm.setUniversity(sr.scraper.university)
	cunyForm.setAction(universityKey)
	cunyForm.setTerm(parseSemester(sr.semester))

	doc := sr.scraper.client.Post(sr.url, url.Values(cunyForm))

	return doc
}

func (sr *subjectScraper) parseSubjects(doc *goquery.Document) (subjects []*model.Subject) {
	str := doc.Find(selectSubjects).Text()
	subs := strings.Split(str, "\n")

	for _, s := range subs {
		s = strings.TrimSpace(s)

		if s != "" && strings.Contains(s, "-") {
			pair := strings.Split(s, " - ")

			if len(pair) != 2 {
				log.Fatalln("expected subject number-name pair")
			}

			sub := &model.Subject{
				Name:   pair[1],
				Number: pair[0],
				Season: sr.semester.Season,
				Year:   strconv.Itoa(int(sr.semester.Year)),
			}

			subjects = append(subjects, sub)
		}
	}
	return
}

type courseScraper struct {
	scraper   *cunyScraper
	url       string
	subjectId string
	full      bool
}

func (cr *courseScraper) scrapeCourses() *goquery.Document {
	form := cunyForm{}
	form.setAction(searchAction)
	form.setSubject(cr.subjectId)

	return cr.scraper.client.Post(cr.url, url.Values(form))
}

func (cr *courseScraper) resetCourses() *goquery.Document {
	form := cunyForm{}
	form.setAction(modifySearchAction)

	return cr.scraper.client.Post(cr.url, url.Values(form))
}

func (cr *courseScraper) parseCourses(doc *goquery.Document) (courses []*model.Course) {
	findCourses(doc.Selection, func(index int, s *goquery.Selection) {
		if index%2 == 0 {
			c := cr.parseCourse(s)
			namenum := strings.Split(strings.TrimSpace(s.Find(".PAGROUPBOXLABELLEVEL1").Text()), " - ")

			if len(namenum) == 1 {
				var last rune
				index = strings.IndexFunc(namenum[0], func(r rune) bool {
					if unicode.IsSpace(r) && unicode.IsSpace(last) {
						return true
					}
					last = r
					return false
				})

				namenum = []string{namenum[0][:index], namenum[0][index+1:]}
			}

			if len(namenum) > 2 {
				namenum = []string{namenum[0], strings.Trim(fmt.Sprint(namenum[1:]), "[]")}
			}

			if len(namenum) != 2 {
				log.Errorln(namenum, len(namenum))
				for _, val := range namenum {
					log.Errorln(val)
				}
			}

			namenum[0] = strings.Split(namenum[0], " ")[1]

			c.Number = namenum[0]
			c.Name = namenum[1]
			courses = append(courses, c)
		}
	})
	return
}

func (cr *courseScraper) parseCourse(doc *goquery.Selection) (course *model.Course) {
	course = &model.Course{}
	sections := []*model.Section{}

	findSections(doc, func(index int, s *goquery.Selection) {
		sections = append(sections, cr.parseSection(s, course))
	})

	course.Sections = sections
	return
}

func findCourses(doc *goquery.Selection, call func(index int, s *goquery.Selection)) {
	doc.Find(selectCourses).Each(call)
}

func findSections(doc *goquery.Selection, call func(index int, s *goquery.Selection)) {
	doc.Find(".PSLEVEL1GRIDNBONBO").Each(call)
}

func parseInstructor(s []string) (instructors []*model.Instructor) {
	for _, val := range s {
		for _, instr := range instructors {
			if instr.Name != val {
				instructors = append(instructors, &model.Instructor{
					Name: val,
				})
				break
			}
		}
	}
	return
}

func (cr *courseScraper) parseSection(doc *goquery.Selection, course *model.Course) (section *model.Section) {
	section = &model.Section{}
	rawSection := cr.findSection(doc.Find(".PSLEVEL3GRIDROW").Eq(1))
	section.Number = rawSection[0]

	m := cr.findMeetings(doc.Find(".PSLEVEL3GRIDROW").Eq(2))

	section.CallNumber = cr.findClass(doc.Find(".PSLEVEL3GRIDROW").Eq(0))[0]

	room := cr.findRoom(doc.Find(".PSLEVEL3GRIDROW").Eq(3))

	section.Status = cr.findStatus(doc.Find(".PSLEVEL3GRIDROW").Eq(6))
	section.Instructors = parseInstructor(cr.findInstructor(doc.Find(".PSLEVEL3GRIDROW").Eq(4)))

	for i := range m {
		meetings := []string{}
		meetings = append(meetings, parseMeeting(m[i])...)
		for j := range meetings {
			meeting := &model.Meeting{}
			tups := splitMeeting(meetings[j])
			meeting.Day = &tups[0]
			meeting.StartTime = &tups[1]
			meeting.EndTime = &tups[2]

			if len(room) < i && room[i] != "" {
				meeting.Room = &room[i]
			}

			section.Meetings = append(section.Meetings, meeting)
		}
	}

	if cr.full {
		sectionId := doc.Find(".PSLEVEL3GRIDROW").Eq(1).Find("a.PSHYPERLINK").AttrOr("name", "")
		sr := &sectionScraper{cr.scraper, cr.url}
		sectionDoc := sr.scrapeSection(sectionId)
		extraSectionInfo := sr.parseSection(sectionDoc, course)
		section.Now = extraSectionInfo.Now
		section.Max = extraSectionInfo.Max
		section.Metadata = extraSectionInfo.Metadata
		section.Credits = extraSectionInfo.Credits
	}

	return
}

func (cr *courseScraper) findClass(s *goquery.Selection) []string {
	return []string{strings.TrimSpace(s.Find("a").Text())}
}

func (cr *courseScraper) findSection(s *goquery.Selection) []string {
	str := strings.TrimSpace(s.Find("a").Text())
	return strings.Split(str, "-")[:2]
}

func (cr *courseScraper) findMeetings(s *goquery.Selection) []string {
	str := strings.TrimSpace(s.Find("span").Text())
	m := strings.Split(str, "\n")
	return m
}

func (cr *courseScraper) findRoom(s *goquery.Selection) []string {
	str := strings.TrimSpace(s.Find("span").Text())
	return strings.Split(str, "\n")
}

func (cr *courseScraper) findInstructor(s *goquery.Selection) []string {
	str := strings.Trim(s.Find("span").Text(), ", ")
	return strings.Split(str, "\n")
}

func (cr *courseScraper) findStatus(s *goquery.Selection) string {
	str := s.Find("img").AttrOr("alt", "")
	if str != model.Open.String() && str != model.Closed.String() {
		str = model.Closed.String()
	}
	return str
}

type sectionScraper struct {
	scraper *cunyScraper
	url     string
}

func (sr *sectionScraper) scrapeSection(sectionId string) *goquery.Document {
	formValues := cunyForm{}
	formValues.setAction(sectionId)

	doc := sr.scraper.client.Post(sr.url, url.Values(formValues))
	// Go back after search
	formValues = cunyForm{}
	formValues.setAction(sectionBackAction)

	_ = sr.scraper.client.Post(sr.url, url.Values(formValues))

	return doc
}

func (sr *sectionScraper) parseSection(doc *goquery.Document, course *model.Course) (section model.Section) {
	if course.Synopsis == nil {
		s := sr.findDescription(doc.Selection)
		course.Synopsis = &s
	}

	section.Max = int64(sr.findMax(doc.Selection))
	section.Now = int64(sr.findNow(doc.Selection))
	section.Credits = sr.findUnits(doc.Selection)

	meta := []func(s *goquery.Selection) *model.Metadata{
		sr.findRequirements,
		sr.findClassAttributes,
		sr.findDesignation,
		sr.findWaitlist,
	}

	for _, fn := range meta {
		if m := fn(doc.Selection); m != nil {
			section.Metadata = append(section.Metadata, m)
		}
	}

	return
}

func (sr *sectionScraper) findDescription(s *goquery.Selection) string {
	str := strings.TrimSpace(s.Find(selectDesc).Text())
	derivedDesc := strings.TrimSpace(s.Find(selectDescTop).Text())

	// If the only description is the name of the class itself, do not return.
	if strings.Contains(derivedDesc, str) {
		return ""
	}

	return str
}

func (sr *sectionScraper) findUnits(s *goquery.Selection) string {
	str := strings.TrimSpace(s.Find(selectUnits).Text())
	if c := strings.Split(str, " "); len(c) >= 2 {
		return c[0]
	} else {
		log.Warningln("unexpected credits format", str)
	}

	return str
}

func (sr *sectionScraper) findInstructionMode(s *goquery.Selection) string {
	str := strings.TrimSpace(s.Find(selectInstructionMode).Text())
	return str
}

func (sr *sectionScraper) findClassComponents(s *goquery.Selection) string {
	str := strings.TrimSpace(s.Find(selectClassComponents).Text())
	return str
}

func (sr *sectionScraper) findRequirements(s *goquery.Selection) *model.Metadata {
	str := strings.TrimSpace(s.Find(selectRequirements).Text())
	if str != "" {
		return &model.Metadata{
			Title:   "Enrollment Requirement",
			Content: str,
		}
	}
	return nil
}

func (sr *sectionScraper) findDesignation(s *goquery.Selection) *model.Metadata {
	str := strings.TrimSpace(s.Find(selectDesignation).Text())
	if str != "" {
		return &model.Metadata{
			Title:   "Requirement Designation",
			Content: str,
		}
	}

	return nil
}

func (sr *sectionScraper) findClassAttributes(s *goquery.Selection) *model.Metadata {
	str := strings.TrimSpace(s.Find(selectClassAttributes).Text())
	if str != "" {
		return &model.Metadata{
			Title:   "Class Attributes",
			Content: str,
		}
	}

	return nil
}

func (sr *sectionScraper) findMax(s *goquery.Selection) int {
	return sr.findAvailability(s, selectEnrollmentCap)
}

func (sr *sectionScraper) findNow(s *goquery.Selection) int {
	return sr.findAvailability(s, selectEnrollmentTotal)
}

func (sr *sectionScraper) findAvailableSeats(s *goquery.Selection) int {
	return sr.findAvailability(s, selectAvailableSeats)
}

func (sr *sectionScraper) findWaitlist(s *goquery.Selection) *model.Metadata {
	capacity := sr.findWaitlistCap(s)
	total := sr.findWaitlistTotal(s)

	if capacity != 0 && total != 0 {
		return &model.Metadata{
			Title:   "Waitlist",
			Content: fmt.Sprintf("Capacity: %d Total: %d", capacity, total),
		}
	}

	return nil
}

func (sr *sectionScraper) findWaitlistCap(s *goquery.Selection) int {
	return sr.findAvailability(s, selectWaitCap)
}

func (sr *sectionScraper) findWaitlistTotal(s *goquery.Selection) int {
	return sr.findAvailability(s, selectWaitTotal)
}

func (sr *sectionScraper) findAvailability(s *goquery.Selection, id string) int {
	str := strings.TrimSpace(s.Find("span#" + id).Text())
	if str != "" {
		i, err := strconv.Atoi(str)
		if err == nil {
			return i
		}
	}

	return 0
}

func parseSemester(semester model.Semester) string {
	year := strconv.Itoa(int(semester.Year))[2:]
	if semester.Season == model.Fall {
		return "1" + year + "9"
	} else if semester.Season == model.Summer {
		return "1" + year + "7"
	} else if semester.Season == model.Spring {
		return "1" + year + "2"
	} else {
		return "1" + year + "0"
	}
}

func newClient() *http.Client {
	jar, _ := cookiejar.New(nil)

	return &http.Client{
		Timeout:   15 * time.Second,
		Jar:       jar,
		Transport: &http.Transport{Proxy: proxy.ProxyUrl(), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
}
