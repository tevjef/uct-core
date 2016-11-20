package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"sync"
	"uct/common/model"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"uct/common/conf"
	"uct/common/try"
	"io/ioutil"
	"net"
)

var (
	app        = kingpin.New("rutgers", "A web scraper that retrives course information for Rutgers University's servers.")
	campusFlag = app.Flag("campus", "Choose campus code. NB=New Brunswick, CM=Camden, NK=Newark").HintOptions("CM", "NK", "NB").Short('u').PlaceHolder("[CM, NK, NB]").Required().String()
	configFile = app.Flag("config", "configuration file for the application").Required().Short('c').File()
	format     = app.Flag("format", "choose input format").Short('f').HintOptions(model.JSON, model.PROTOBUF).PlaceHolder("[protobuf, json]").Default("protobuf").String()
	latest     = app.Flag("latest", "Only output the current and next semester").Short('l').Bool()
	config     conf.Config
)

type rutgersRequest struct {
	semester model.Semester
	host     string
	campus   string
}

type subjectRequest struct {
	rutgersRequest
}

type courseRequest struct {
	subject string
	rutgersRequest
}

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
	*campusFlag = strings.ToUpper(*campusFlag)
	app.Name = app.Name + "-" + strings.ToLower(*campusFlag)
	log.SetLevel(log.DebugLevel)

	// Parse configuration file
	config = conf.OpenConfig(*configFile)
	config.AppName = app.Name

	// Start profiling
	go model.StartPprof(config.GetDebugSever(app.Name))

	if reader, err := model.MarshalMessage(*format, getCampus(*campusFlag)); err != nil {
		log.WithError(err).Fatal()
	} else {
		io.Copy(os.Stdout, reader)
	}
}

func getCampus(campus string) model.University {
	var university model.University

	university = getRutgers(campus)

	// Rutgers servers go down for maintenance between 2 and 5 AM UTC every day.
	// Data scraped from this time would be inaccurate and may lead to unforeseen errors
	if currentHour := time.Now().UTC().Hour(); currentHour >= 6 && currentHour < 9 {
		return model.University{}
	}

	university.ResolvedSemesters = model.ResolveSemesters(time.Now(), university.Registrations)

	semesters := []*model.Semester{
		university.ResolvedSemesters.Last,
		university.ResolvedSemesters.Current,
		university.ResolvedSemesters.Next}

	if *latest {
		semesters = semesters[1:]
	}

	for _, semester := range semesters {
		if semester.Season == model.Winter {
			semester.Year++
		}

		rr := rutgersRequest{
			host:     "http://sis.rutgers.edu/soc",
			semester: *semester,
			campus:   campus,
		}

		subjects := subjectRequest{rutgersRequest: rr}.requestSubjects()

		var wg sync.WaitGroup

		control := make(chan struct{}, 10)
		for i := range subjects {
			wg.Add(1)
			control <- struct{}{}
			go func(sub *RSubject) {
				defer func() { wg.Done() }()
				sub.Courses = courseRequest{rutgersRequest: rr, subject: sub.Number}.requestCourses()
				<-control
			}(subjects[i])
		}
		wg.Wait()

		university.Subjects = append(university.Subjects, buildSubjects(subjects)...)
	}

	return university
}

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		//TLSHandshakeTimeout:   10 * time.Second,
		//ExpectContinueTimeout: 1 * time.Second,
	},
}

func (sr subjectRequest) requestSubjects() (subjects []*RSubject) {
	var url = fmt.Sprintf("%s/subjects.json?semester=%s&campus=%s&level=U%%2CG", sr.host, parseSemester(sr.semester), sr.campus)
	if err := getData(url, &subjects); err == nil {
		for i := range subjects {
			subject := subjects[i]
			subject.Season = sr.semester.Season
			subject.Year = int(sr.semester.Year)
			subject.clean()
		}
	}

	return
}

func (cr courseRequest) requestCourses() (courses []*RCourse) {
	var url = fmt.Sprintf("%s/courses.json?subject=%s&semester=%s&campus=%s&level=U%%2CG", cr.host, cr.subject, parseSemester(cr.semester), cr.campus)
	if err := getData(url, &courses); err == nil {
		for i := range courses {
			course := courses[i]
			course.clean()
		}
	}

	courses = filterCourses(courses, func(course *RCourse) bool {
		return len(course.Sections) > 0
	})

	sort.Sort(CourseSorter{courses})
	return
}

func buildSubjects(rutgersSubjects []*RSubject) (s []*model.Subject) {
	// Filter subjects that don't have a course
	rutgersSubjects = filterSubjects(rutgersSubjects, func(subject *RSubject) bool {
		return len(subject.Courses) > 0
	})

	for _, subject := range rutgersSubjects {
		newSubject := &model.Subject{
			Name:    subject.Name,
			Number:  subject.Number,
			Season:  subject.Season,
			Year:    strconv.Itoa(subject.Year),
			Courses: buildCourses(subject.Courses)}
		s = append(s, newSubject)
	}
	return
}

func buildCourses(rutgersCourses []*RCourse) (c []*model.Course) {
	for _, course := range rutgersCourses {
		newCourse := &model.Course{
			Name:     course.ExpandedTitle,
			Number:   course.CourseNumber,
			Synopsis: &course.CourseDescription,
			Metadata: course.metadata(),
			Sections: buildSections(course.Sections)}

		c = append(c, newCourse)
	}
	return
}

func buildSections(rutgerSections []*RSection) (s []*model.Section) {
	for _, section := range rutgerSections {
		newSection := &model.Section{
			Number:     section.Number,
			CallNumber: section.Index,
			Status:     section.status,
			Credits:    section.creditsFloat,
			Metadata:   section.metadata()}

		for _, instructor := range section.Instructor {
			newInstructor := &model.Instructor{Name: instructor.Name}
			newSection.Instructors = append(newSection.Instructors, newInstructor)
		}

		for _, meeting := range section.MeetingTimes {
			newMeeting := &model.Meeting{
				Room:      &meeting.RoomNumber,
				Day:       &meeting.MeetingDay,
				StartTime: &meeting.StartTime,
				EndTime:   &meeting.EndTime,
				ClassType: &meeting.ClassType}

			newSection.Meetings = append(newSection.Meetings, newMeeting)
		}
		s = append(s, newSection)
	}
	return
}

func getData(url string, model interface{}) error {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("User-Agent", "Go/rutgers-scraper")

	fields := log.WithFields(log.Fields{"url": url, "model_type": fmt.Sprintf("%T", model)})
	fields.Debugln()

	err := try.Do(func(attempt int) (retry bool, err error) {
		startTime := time.Now()
		time.Sleep(time.Duration((attempt - 1 )*2) * time.Second)
		resp, err := httpClient.Do(req)

		var data []byte
		if err != nil {
			return true, err
		} else if data, err = ioutil.ReadAll(resp.Body); err != nil {
			return true, err
		} else if err = json.Unmarshal(data, model); err != nil {
			return true, err
		}

		fields.WithFields(log.Fields{
			"content-length": len(data),
			"response_status": resp.StatusCode,
			"response_time": time.Since(startTime).Seconds()}).Debugln()

		return false, nil
	})

	return err
}

func parseSemester(semester model.Semester) string {
	year := strconv.Itoa(int(semester.Year))
	if semester.Season == model.Fall {
		return "9" + year
	} else if semester.Season == model.Summer {
		return "7" + year
	} else if semester.Season == model.Spring {
		return "1" + year
	} else if semester.Season == model.Winter {
		return "0" + year
	}
	return ""
}
