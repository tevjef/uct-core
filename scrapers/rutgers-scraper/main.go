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
)

var (
	app        = kingpin.New("rutgers", "A web scraper that retrives course information for Rutgers University's servers.")
	campusFlag = app.Flag("campus", "Choose campus code. NB=New Brunswick, CM=Camden, NK=Newark").HintOptions("CM", "NK", "NB").Short('u').PlaceHolder("[CM, NK, NB]").Required().String()
	latest     = app.Flag("latest", "Only output the current and next semester").Short('l').Bool()
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
	log.SetLevel(log.InfoLevel)

	if reader, err := model.MarshalMessage(model.JSON, getCampus(*campusFlag)); err != nil {
		log.WithError(err).Fatal()
	} else {
		io.Copy(os.Stdout, reader)
	}
}

func getCampus(campus string) model.University {
	var university model.University

	university = getRutgers(campus)

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
	var success bool
	modeType := fmt.Sprintf("%T", model)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("User-Agent", "Go/rutgers-scraper")
	for i := 0; i < 3; i++ {
		startTime := time.Now()
		log.WithFields(log.Fields{"retry": i, "url": url}).Debug(modeType + " request")
		time.Sleep(time.Duration(i*2) * time.Second)
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Errorf("Retrying %d after error: %s\n", i, err)
			continue
		} else if err := json.NewDecoder(resp.Body).Decode(model); err != nil {
			log.Errorf("Retrying %d after error: %s\n", i, err)
			continue
		} else {
			log.WithFields(log.Fields{"content-length": resp.ContentLength, "status": resp.Status, "url": url, "response_time": time.Since(startTime).Seconds()}).Debug(modeType + " response")
			success = true
			break
		}

	}

	if !success {
		return fmt.Errorf("Unable to retrieve resource at %s", url)
	}
	return nil
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
