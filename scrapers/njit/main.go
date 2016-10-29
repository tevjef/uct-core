package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"uct/common/model"
	"io"
	"os"
)

func main() {

	var university model.University

	log.SetFormatter(&log.TextFormatter{})
	university = njit

	university.ResolvedSemesters = model.ResolveSemesters(time.Now(), university.Registrations)

	semesters := []*model.Semester{
		university.ResolvedSemesters.Last,
		university.ResolvedSemesters.Current,
		university.ResolvedSemesters.Next}[:1]

	for _, semester := range semesters {
		jar, _ = cookiejar.New(nil)

		subjectRequest := subjectRequest{semester: *semester, paginated: paginated{offset: 1, max: 20}}
		subjects := subjectRequest.requestSubjects()

		for i := range subjects {
			courseRequest := courseRequest{semester: *semester, subject: subjects[i].Code, paginated: paginated{offset: 0, max: 20}}
			subjects[i].courses = courseRequest.requestSearch()
		}

		university.Subjects = append(university.Subjects, buildSubjects(subjects)...)
	}

	reader := model.MarshalMessage(model.JSON, university)
	io.Copy(os.Stdout, reader)
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
		if err := getData(url, parseSemester(sr.semester), &page); err == nil {
			for i := range subjects {
				subject := subjects[i]
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
	cr.offset+=cr.max
	httpClient.PostForm("https://myhub.njit.edu/StudentRegistrationSsb/ssb/classSearch/resetDataForm", url.Values{})
}

func (cr *courseRequest) requestSearch() (courses []*NCourse) {
	var url = cr.buildUrl()
	page := NSearch{}
	for firstPage := true; len(page.Data) >= cr.max || firstPage; {

		page = NSearch{}
		if err := getData(url, parseSemester(cr.semester), &page); err == nil {
			courses = append(courses, page.Data...)
		}

		firstPage = false
		cr.paginate()
		url = cr.buildUrl()

		log.Debugf("%s len in loop=%d offser=%d max=%d\n", cr.subject, len(page.Data), cr.offset, cr.max )
	}

	courses = cleanCourseList(courses)

	return
}

var proxyUrl, _ = url.Parse("http://localhost:8888")

var httpClient = &http.Client{
	Timeout:   15 * time.Second,
	Jar:       jar,
	Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
}

var jar, _ = cookiejar.New(nil)

var token = uuid.NewV4().String()

func doRequest(rawUrl string) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodGet, rawUrl, nil)

	req.Header.Set("Accept", "application/json, text/javascript, */*; 0.01")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("X-Synchronizer-Token", token)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36")
	req.Header.Set("Origin", "https://myhub.njit.edu")
	if strings.Contains(rawUrl, "searchResults") || strings.Contains(rawUrl, "classSearch") {
		req.Header.Set("Referer", "https://myhub.njit.edu/StudentRegistrationSsb/ssb/classSearch/classSearch")
	}

	return httpClient.Do(req)
}

var cookieMap = make(map[string]*http.Cookie)

func getCookie(term string) *http.Cookie {
	if cookieMap[term] == nil {
		cookie := &http.Cookie{
			Name:   "JSESSIONID",
			Path:   "/StudentRegistrationSsb/",
			Domain: "myhub.njit.edu",
		}

		resp, _ := httpClient.PostForm("https://myhub.njit.edu/StudentRegistrationSsb/ssb/term/search?mode=search", url.Values{"term": []string{term}})
		if len(resp.Cookies()) > 0 && cookie.Value == "" {
			cookie.Value = resp.Cookies()[0].Value
			resp, _ = doRequest("https://myhub.njit.edu/StudentRegistrationSsb/ssb/classSearch/classSearch")
		}

		cookieMap[term] = cookie

	}

	return cookieMap[term]
}

func getData(rawUrl string, term string, model interface{}) error {
	var success bool
	modeType := fmt.Sprintf("%T", model)

	u, _ := url.Parse(rawUrl)
	jar.SetCookies(u, []*http.Cookie{getCookie(term)})

	for i := 0; i < 3; i++ {
		startTime := time.Now()
		log.WithFields(log.Fields{"retry": i, "url": rawUrl}).Debug(modeType + " request")
		time.Sleep(time.Duration(i*2) * time.Second)
		resp, err := doRequest(rawUrl)
		if err != nil {
			log.Errorf("Retrying %d after error: %s\n", i, err)
			continue
		} else if data, err := ioutil.ReadAll(resp.Body); err != nil {
			log.Errorf("Retrying %d after error: %s\n", i, err)
			continue
		} else if err := json.Unmarshal(data, model); err != nil {
			log.Errorf("Retrying %d after error: %s\n", i, err)
			continue
		} else {
			log.WithFields(log.Fields{"content-length": len(data), "response_status": resp.StatusCode, "url": rawUrl, "response_time": time.Since(startTime).Seconds()}).Debug(modeType + " response")
			success = true
			break
		}

	}

	if !success {
		return fmt.Errorf("Unable to retrieve resource at %s", rawUrl)
	}

	return nil
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
		newSection.Credits = strconv.Itoa(course.CreditHours)
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
