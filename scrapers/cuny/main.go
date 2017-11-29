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
	"github.com/pkg/errors"
	"github.com/tevjef/uct-core/common/conf"
	"github.com/tevjef/uct-core/common/model"
	"github.com/tevjef/uct-core/common/proxy"
	"github.com/tevjef/uct-core/common/try"
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
	university   string
	full         bool
	outputFormat string
	latest       bool
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
}

func main() {
	cconf := &cunyConfig{}

	app := kingpin.New("cuny", "A program for scraping information from CunyFirst.")

	app.Flag("campus", "Choose a cuny campus"+abbrMap()).
		Short('u').
		Required().
		Envar("CUNY_UNIVERSITY").
		HintOptions(cunyUniversityAbbr...).
		EnumVar(&cconf.university, cunyUniversityAbbr...)

	app.Flag("section", "Get all section data").
		Short('s').
		Envar("CUNY_ALL_SECTION").
		BoolVar(&cconf.full)

	app.Flag("format", "Choose output format").
		Short('f').
		HintOptions(model.Json, model.Protobuf).
		PlaceHolder("[protobuf, json]").
		Default("protobuf").
		Envar("CUNY_OUTPUT_FORMAT").
		EnumVar(&cconf.outputFormat, "protobuf", "json")

	configFile := app.Flag("config", "Configuration file for the application").
		Short('c').
		Required().
		Envar("CUNY_CONFIG").
		File()

	kingpin.MustParse(app.Parse(os.Args[1:]))
	app.Name = app.Name + "-" + strings.ToLower(cconf.university)

	// Parse configuration file
	cconf.service = conf.OpenConfigWithName(*configFile, app.Name)

	// Start profiling
	go model.StartPprof(cconf.service.DebugSever(app.Name))

	(&cuny{
		app:    app.Model(),
		config: cconf,
		ctx:    context.TODO(),
	}).init()
}

func (cuny *cuny) init() {
	university := cunyMetadata[cuny.config.university]
	university.ResolvedSemesters = model.ResolveSemesters(time.Now(), registrations)

	semesters := [2]*model.Semester{university.ResolvedSemesters.Current, university.ResolvedSemesters.Next}

	for _, sem := range semesters {
		// Do not scrape winter and summer semesters
		if sem.Season == model.Winter || sem.Season == model.Summer {
			continue
		}

		scraper := &cunyScraper{
			university: abbrToCunyUniversity(cuny.config.university),
			client: &CunyFirstClient{
				values:     map[string][]string{},
				httpClient: newClient(),
			},
			full: cuny.config.full,
		}

		sr := &subjectScraper{
			scraper:  scraper,
			url:      initalPage,
			semester: *sem,
		}

		doc, err := sr.scrapeSubjects()
		if err != nil {
			continue
		}

		subjects := sr.parseSubjects(doc)

		var wg sync.WaitGroup
		sem := make(chan *sync.WaitGroup, 5)
		wg.Add(len(subjects))
		for i := range subjects {
			sem <- &wg
			go func(subject *model.Subject) {
				// Retry when a request times out. A timeout makes the cookie jar go stale,
				// blocking future requests from completing. Retry the entire procedure
				// request in that case
				try.DoN(func(attempt int) (bool, error) {
					err := scraper.run(subject)
					if err != nil && err == ErrTimeout {
						log.Warningln("retrying a timed out request")
						return true, err
					}
					return false, nil
				}, 3)
				wg := <-sem
				wg.Done()
			}(subjects[i])
		}
		wg.Wait()

		university.Subjects = append(university.Subjects, subjects...)
	}

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

func (cs *cunyScraper) run(subject *model.Subject) error {
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

	sr := &subjectScraper{
		scraper: scraper,
		url:     initalPage,
		semester: model.Semester{
			Year:   2017,
			Season: model.Spring,
		},
		full: cs.full,
	}

	// Scrape subjects again to initialize cookies necessary to scrape courses
	_, err := sr.scrapeSubjects()
	if err != nil {
		return err
	}

	cr := courseScraper{
		scraper:   scraper,
		url:       initalPage,
		subjectId: subject.Number,
		full:      cs.full,
		semester:  sr.semester}

	doc, err := cr.scrapeCourses()
	if err != nil {
		return err
	}

	if doc == nil {
		log.Warnln("skipping courses for", subject.Name)
	} else {
		subject.Courses = cr.parseCourses(doc)
	}

	return nil
}

type subjectScraper struct {
	scraper  *cunyScraper
	semester model.Semester
	url      string
	full     bool
}

func (sr *subjectScraper) scrapeSubjects() (*goquery.Document, error) {
	defer func(start time.Time) {
		log.WithFields(log.Fields{
			"season":  sr.semester.GetSeason(),
			"year":    sr.semester.GetYear(),
			"elapsed": time.Since(start).Seconds()}).
			Debugln("scrape subjects")
	}(time.Now())

	sr.scraper.client.Get(initalPage)

	form := cunyForm{}
	form.setUniversity(sr.scraper.university)
	form.setAction(universityKey)
	form.setTerm(parseSemester(sr.semester))

	return sr.scraper.client.Post(sr.url, url.Values(form))
}

func (sr *subjectScraper) parseSubjects(doc *goquery.Document) (subjects []*model.Subject) {
	str := doc.Find(selectSubjects).Text()
	subs := strings.Split(str, "\n")

	for _, s := range subs {
		s = strings.TrimSpace(s)

		if s != "" && strings.Contains(s, "-") {
			pair := strings.SplitN(s, "-", 2)

			if len(pair) != 2 {
				log.Fatalln("unexpected subject number-name pair", s)
			}

			sub := &model.Subject{
				Name:   strings.TrimSpace(pair[1]),
				Number: strings.TrimSpace(pair[0]),
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
	semester  model.Semester
	url       string
	subjectId string
	full      bool
}

func (cr *courseScraper) scrapeCourses() (*goquery.Document, error) {
	defer func(start time.Time) {
		log.WithFields(log.Fields{
			"subject": cr.subjectId,
			"season":  cr.semester.GetSeason(),
			"year":    cr.semester.GetYear(),
			"elapsed": time.Since(start).Seconds()}).
			Debugln("scrape courses")
	}(time.Now())

	form := cunyForm{}
	form.setAction(searchAction)
	form.setSubject(cr.subjectId)
	form.setUniversity(cr.scraper.university)
	form.setTerm(parseSemester(cr.semester))

	return cr.scraper.client.Post(cr.url, url.Values(form))
}

func (cr *courseScraper) resetCourses() (*goquery.Document, error) {
	form := cunyForm{}
	form.setAction(modifySearchAction)

	return cr.scraper.client.Post(cr.url, url.Values(form))
}

func (cr *courseScraper) parseCourses(doc *goquery.Document) (courses []*model.Course) {
	findCourses(doc.Selection, func(index int, s *goquery.Selection) {
		// Every even index contains the course data
		if index%2 == 0 {
			courses = append(courses, cr.parseCourse(s))
		}
	})
	return
}

func (cr *courseScraper) parseCourse(s *goquery.Selection) (course *model.Course) {
	course = &model.Course{}

	// Attempts to find the course name and number
	rawstr := strings.TrimSpace(s.Find(".PAGROUPBOXLABELLEVEL1").Text())
	namenum := strings.Split(rawstr, "-")

	if len(namenum) == 1 {
		var last rune
		index := strings.IndexFunc(namenum[0], func(r rune) bool {
			if unicode.IsSpace(r) && unicode.IsSpace(last) {
				return true
			}
			last = r
			return false
		})

		namenum = []string{namenum[0][:index], namenum[0][index+1:]}
	}

	if len(namenum) >= 2 {
		namenum = []string{model.TrimAll(namenum[0]), strings.Trim(fmt.Sprint(namenum[1:]), "[] ")}
	}

	if len(namenum) != 2 {
		log.WithError(errors.New("failed to find course name and number")).Errorln(rawstr)
	}

	namenum[0] = strings.Split(namenum[0], " ")[1]

	course.Number = namenum[0]
	course.Name = namenum[1]

	if course.Number == "" {
		fmt.Print(s.Html())
		log.Fatalln()
	}

	// Compile all the sections for this course
	findSections(s, func(index int, s *goquery.Selection) {
		course.Sections = append(course.Sections, cr.parseSection(s, course))
	})

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

// Within the context of some course, parses all sections.
func (cr *courseScraper) parseSection(doc *goquery.Selection, course *model.Course) (section *model.Section) {
	section = &model.Section{}

	// findSection returns an array like ["01", "LEC\nRegular"] where the first index is the section number
	rawSection := cr.findSection(doc.Find(selectGridRow).Eq(1))
	section.Number = rawSection[0]

	// Section meeting times, contains an unexpanded meeting time string e.g MoWe 8:00AM - 9:50AM
	m := cr.findMeetings(doc.Find(selectGridRow).Eq(2))

	// Section call number, should uniquely identify a section within a course
	section.CallNumber = cr.findClass(doc.Find(selectGridRow).Eq(0))[0]

	// Section room number
	room := cr.findRoom(doc.Find(selectGridRow).Eq(3))

	// Section status,
	section.Status = cr.findStatus(doc.Find(selectGridRow).Eq(6))

	// Section instructor collects all instructors and returns a []model.Instructor
	section.Instructors = parseInstructor(cr.findInstructor(doc.Find(selectGridRow).Eq(4)))

	// Only numeric types allowed
	section.Credits = "-1"

	for i := range m {

		// Expand meetings from e.g MoWe 8:00AM - 9:50AM to e.g ["Monday 8:00AM - 9:50AM", "Wednesday 8:00AM - 9:50AM"]
		meetings := []string{}
		meetings = append(meetings, expandMeeting(m[i])...)

		// For each expanded meeting, create a model.Meeting and append it to the list
		for j := range meetings {
			meeting := &model.Meeting{}

			// Split a meeting into its components Day, Start, End
			tups := splitMeeting(meetings[j])

			meeting.Day = &tups[0]
			meeting.StartTime = &tups[1]
			meeting.EndTime = &tups[2]

			if room[i] != "" {
				meeting.Room = &room[i]
			}

			if *meeting.Day == "" && *meeting.StartTime == "" && *meeting.EndTime == "" {
				continue
			}

			section.Meetings = append(section.Meetings, meeting)
		}
	}

	// This routine with retrieve all data about a section and will increase the time to scrape significantly.
	// This is disables by default
	if cr.full {
		// Find the id of the section to be scraped
		sectionId := doc.Find(selectGridRow).Eq(1).Find(selectSectionLink).AttrOr("name", "")
		sr := &sectionScraper{cr.scraper, cr.url}

		//
		sectionDoc := sr.scrapeSection(sectionId)

		// Parses section in the context of some course
		extraSectionInfo := sr.parseSection(sectionDoc, course)

		// Merge new section returned after parsing.
		section.Now = extraSectionInfo.Now
		section.Max = extraSectionInfo.Max
		section.Metadata = extraSectionInfo.Metadata
		section.Credits = extraSectionInfo.Credits
	}

	return
}

func (cr *courseScraper) parseRoom(room string) string {
	switch cr.scraper.university {
	case BaruchCollege:
		return strings.Replace(room, "B - Vert ", "V", 1)
	default:
		room = strings.Replace(room, "Bldg ", "", 1)
		room = strings.Replace(room, "Hall", "H", 1)

		return room
	}
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
	form := cunyForm{}
	form.setAction(sectionId)

	doc, _ := sr.scraper.client.Post(sr.url, url.Values(form))

	// Go back after search. Is necessary since the website
	// uses cookies to track resource requests
	form = cunyForm{}
	form.setAction(sectionBackAction)

	sr.scraper.client.Post(sr.url, url.Values(form))

	return doc
}

// Parses section in the context of some course, returns a section containing
// information that could not be parse from the course page
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
		Timeout:   30 * time.Second,
		Jar:       jar,
		Transport: &http.Transport{Proxy: proxy.ProxyUrl(), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
}
