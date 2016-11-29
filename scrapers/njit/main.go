package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"uct/common/conf"
	"uct/common/model"
	"uct/common/proxy"
	"uct/common/try"
	"uct/scrapers/njit/cookie"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

var (
	app        = kingpin.New("njit", "A program for scraping information from NJIT serrvers.")
	configFile = app.Flag("config", "configuration file for the application").Required().Short('c').File()
	config     = conf.Config{}
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	log.SetLevel(log.DebugLevel)

	// Parse configuration file
	config = conf.OpenConfig(*configFile)
	config.AppName = app.Name

	// Start profiling
	go model.StartPprof(config.GetDebugSever(app.Name))

	var university model.University

	university = njit

	university.ResolvedSemesters = model.ResolveSemesters(time.Now(), university.Registrations)

	semesters := []*model.Semester{
		university.ResolvedSemesters.Last,
		university.ResolvedSemesters.Current,
		university.ResolvedSemesters.Next}

	for _, semester := range semesters {
		CreateCookieQueue(parseSemester(*semester))

		subjectRequest := subjectRequest{semester: *semester, paginated: paginated{offset: 1, max: 10}}
		subjects := subjectRequest.requestSubjects()

		var wg sync.WaitGroup
		control := make(chan struct{}, 10)
		wg.Add(len(subjects))
		for i := range subjects {
			control <- struct{}{}
			go func(subjects *NSubject) {
				defer func() { wg.Done() }()
				courseRequest := courseRequest{semester: *semester, subject: subjects.Code, paginated: paginated{offset: 0, max: 20}}
				subjects.courses = courseRequest.requestSearch()
				<-control
			}(subjects[i])
		}
		wg.Wait()

		university.Subjects = append(university.Subjects, buildSubjects(subjects)...)
	}

	if reader, err := model.MarshalMessage(model.JSON, university); err != nil {
		io.Copy(os.Stdout, reader)
	}
}

const baseUrl = "https://myhub.njit.edu/StudentRegistrationSsb/ssb/"

type paginated struct {
	offset int
	max    int
	sync.Mutex
}

type termRequest struct {
	paginated
}

type subjectRequest struct {
	semester model.Semester
	paginated
}

func (sr *subjectRequest) buildUrl() string {
	sr.Lock()
	defer sr.Unlock()
	return fmt.Sprintf("%sclassSearch/get_subject?term=%s&offset=%d&max=%d&searchTerm=&_=%d&startDatepicker=&endDatepicker=&sortColumn=subjectDescription",
		baseUrl, parseSemester(sr.semester), sr.offset, sr.max, time.Now().Unix())
}

func (sr *subjectRequest) paginate() {
	sr.Lock()
	defer sr.Unlock()
	sr.offset++
}

func (sr *subjectRequest) requestSubjects() (subjects []*NSubject) {
	var url = sr.buildUrl()

	page := []*NSubject{}

	for firstPage := true; len(page) == sr.max || firstPage; {
		page = []*NSubject{}
		if err := getData(url, &page); err == nil {
			for i := range page {
				subject := page[i]
				subject.clean()
				subject.season = sr.semester.Season
				subject.year = int(sr.semester.Year)
			}

			subjects = append(subjects, page...)
		}

		firstPage = false
		sr.paginate()
		url = sr.buildUrl()
	}

	return
}

type courseRequest struct {
	subject  string
	semester model.Semester
	paginated
}

func (cr *courseRequest) buildUrl() string {
	cr.Lock()
	defer cr.Unlock()
	return fmt.Sprintf("%ssearchResults/searchResults?txt_subject=%s&txt_term=%s&pageOffset=%d&pageMax=%d&sortDirection=asc",
		baseUrl, cr.subject, parseSemester(cr.semester), cr.offset, cr.max)
}

func (cr *courseRequest) paginate() {
	cr.Lock()
	defer cr.Unlock()
	cr.offset += cr.max
	httpClient.PostForm("https://myhub.njit.edu/StudentRegistrationSsb/ssb/classSearch/resetDataForm", url.Values{})
}

func (cr *courseRequest) requestSearch() (courses []*NCourse) {
	var url = cr.buildUrl()
	page := NSearch{}
	for firstPage := true; len(page.Data) >= cr.max || firstPage; {

		page = NSearch{}
		if err := getData(url, &page); err == nil {
			courses = append(courses, page.Data...)
		}

		firstPage = false
		cr.paginate()
		url = cr.buildUrl()
	}

	courses = cleanCourseList(courses)

	return
}

var httpClient = &http.Client{
	Timeout:   15 * time.Second,
	Transport: &http.Transport{Proxy: proxy.GetProxyUrl(), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
}

func getData(rawUrl string, model interface{}) error {
	modeType := fmt.Sprintf("%T", model)
	req, _ := http.NewRequest(http.MethodGet, rawUrl, nil)
	startTime := time.Now()

	err := try.Do(func(attempt int) (retry bool, err error) {
		// Get cookie
		bc := cookieCutter.Pop(nil)
		req.AddCookie(bc.Get())

		resp, err := httpClient.Do(req)

		// Put cookie back on queue
		cookieCutter.Push(bc, func(baked *cookie.BakedCookie) error {
			resetCookie(*bc.Get())
			return nil
		})

		var data []byte

		if err != nil {
			return true, err
		} else if data, err = ioutil.ReadAll(resp.Body); err != nil {
			return true, err
		} else if err = json.Unmarshal(data, model); err != nil {
			return true, err
		}

		log.WithFields(log.Fields{"content-length": len(data),
			"response_status": resp.StatusCode,
			"url":             rawUrl,
			"response_time":   time.Since(startTime).Seconds(),
		}).Debug(modeType + " response")

		return false, nil
	})

	if err != nil {
		return errors.Wrap(err, "Unable to retrieve resource at "+rawUrl)
	} else {
		return nil

	}
}

func buildSubjects(njitSubjects []*NSubject) (s []*model.Subject) {
	for _, subject := range njitSubjects {
		newSubject := &model.Subject{
			Name:    subject.Description,
			Number:  subject.Code,
			Season:  subject.season,
			Year:    strconv.Itoa(subject.year),
			Courses: buildCourses(subject.courses),
		}
		s = append(s, newSubject)
	}
	return
}

func buildCourses(njitCourses []*NCourse) (c []*model.Course) {
	for _, course := range collapseCourses(njitCourses) {
		newCourse := &model.Course{}
		courseSection := course[0]
		newCourse.Name = courseSection.CourseTitle
		newCourse.Number = courseSection.CourseNumber
		newCourse.Sections = buildSections(course)
		c = append(c, newCourse)
	}

	return
}

func buildSections(njitCourses []*NCourse) (s []*model.Section) {
	for _, course := range njitCourses {
		newSection := &model.Section{}
		newSection.Number = course.SequenceNumber
		newSection.Status = course.status
		newSection.CallNumber = course.CourseReferenceNumber
		newSection.Credits = fmt.Sprintf("%.1f", course.CreditHours)
		newSection.Now = int64(course.Enrollment)
		newSection.Max = int64(course.MaximumEnrollment)
		// Instructors
		newSection.Instructors = buildInstructors(course)
		// Meetings
		newSection.Meetings = buildMeetings(course)
		// Metadata instructor name-email
		s = append(s, newSection)
	}

	return
}

func buildMeetings(njitCourse *NCourse) (meetings []*model.Meeting) {
	for _, val := range njitCourse.meetingTimes {
		newMeeting := &model.Meeting{
			StartTime: &val.BeginTime,
			EndTime:   &val.EndTime,
			Room:      &val.Room,
			ClassType: &val.MeetingScheduleType,
			Day:       &val.day,
		}

		meetings = append(meetings, newMeeting)
	}

	return
}

func buildInstructors(njitCourse *NCourse) (instructors []*model.Instructor) {
	for _, val := range njitCourse.Faculty {
		newInstructor := &model.Instructor{
			Name: val.DisplayName,
		}

		instructors = append(instructors, newInstructor)
	}
	return
}

func parseSemester(semester model.Semester) string {
	year := strconv.Itoa(int(semester.Year))
	if semester.Season == model.Fall {
		return year + "90"
	} else if semester.Season == model.Summer {
		return year + "70"
	} else if semester.Season == model.Spring {
		return year + "10"
	} else if semester.Season == model.Winter {
		return year + "95"
	}
	return ""
}

var cookieCutter *cookie.CookieCutter

func CreateCookieQueue(term string) {
	var queueSize = 30
	var cookies []*http.Cookie
	for i := 0; i < queueSize; i++ {
		cookies = append(cookies, &http.Cookie{
			Name:   "JSESSIONID",
			Path:   "/StudentRegistrationSsb/",
			Domain: "myhub.njit.edu",
		})
	}

	cc := cookie.New(cookies, func(bc *cookie.BakedCookie) error {
		bc.SetValue(prepareCookie(term))
		return nil
	})

	cookieCutter = cc
}

func resetCookie(cookie http.Cookie) {
	err := try.Do(func(attempt int) (retry bool, err error) {
		req, _ := http.NewRequest(http.MethodPost, "https://myhub.njit.edu/StudentRegistrationSsb/ssb/classSearch/classSearch", strings.NewReader(url.Values{}.Encode()))
		req.AddCookie(&cookie)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		_, err = httpClient.Do(req)

		if err != nil {
			log.WithError(err).Errorln("Failed to validate cookie", attempt)
			return true, err
		}
		return false, nil
	})

	if err != nil {
		log.WithError(err).Fatalln("Failed to reset cookie")
	}
}

func prepareCookie(term string) (value string) {
	err := try.Do(func(attempt int) (retry bool, err error) {
		resp, err := httpClient.PostForm("https://myhub.njit.edu/StudentRegistrationSsb/ssb/term/search?mode=search", url.Values{"term": []string{term}})

		if err != nil {
			log.WithError(err).Errorln("Failed get cookie", attempt)
			return true, err
		}

		if len(resp.Cookies()) > 0 {
			cookie := resp.Cookies()[0]
			req, _ := http.NewRequest(http.MethodGet, "https://myhub.njit.edu/StudentRegistrationSsb/ssb/classSearch/classSearch", nil)
			req.AddCookie(cookie)
			_, err := httpClient.Do(req)

			if err != nil {
				log.WithError(err).Errorln("Failed to validate cookie", attempt)
				return true, err
			}

			value = cookie.Value
		}

		return false, nil
	})

	if err != nil {
		log.WithError(err).Fatalln("Failed to prepare cookie")
	}

	return
}
