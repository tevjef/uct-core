package main

import ()
import (
	"fmt"
	"strconv"
	"sync"
	"time"
	"uct/common/model"
	"net/http"
	log "github.com/Sirupsen/logrus"
	"encoding/json"
)


func main() {

	var university model.University

	university = njit

	university.ResolvedSemesters = model.ResolveSemesters(time.Now(), university.Registrations)

	_ = []*model.Semester{
		university.ResolvedSemesters.Last,
		university.ResolvedSemesters.Current,
		university.ResolvedSemesters.Next}

}

const baseUrl = "https://myhub.njit.edu/StudentRegistrationSsb/ssb/"

type paginated struct {
	offset int
	max    int
	limit int
	sync.Mutex
}

type termRequest struct {
	paginated
}

type subjectRequest struct {
	term string
	paginated
}

func (sr subjectRequest) buildUrl() string {
	sr.Lock()
	defer sr.Unlock()
	return fmt.Sprintf("%sclassSearch/get_subject?term=%s&offset=%d&max=%d",
		baseUrl, sr.term, sr.offset, sr.max)
}

type courseRequest struct {
	term    string
	subject string
	paginated
}

func (cr courseRequest) buildUrl() string {
	cr.Lock()
	defer cr.Unlock()
	return fmt.Sprintf("%ssearchResults/searchResults?txt_subject=%s&txt_term=%spageOffset=%d&pageMaxSize=%d&sortDirection=asc",
		baseUrl, cr.subject, cr.term, cr.offset, cr.max)
}

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
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
			log.WithFields(log.Fields{"content-length": resp.ContentLength, "response_status": resp.StatusCode, "url": url, "response_time": time.Since(startTime).Seconds()}).Debug(modeType + " response")
			success = true
			break
		}

	}

	if !success {
		return fmt.Errorf("Unable to retrieve resource at %s", url)
	}
	return nil
}

func buildSubjects(njitSubjects []*NSubject) (s []*model.Subject) {

	for _, subject := range njitSubjects {
		newSubject := &model.Subject{
			Name:   subject.Description,
			Number: subject.Code,
			Season: subject.season,
			Year:   strconv.Itoa(subject.year),
			Courses: buildCourses(subject.courses),
		}
		s = append(s, newSubject)
	}
	return
}

func buildCourses(njitCourses []NCourse) (c []*model.Course) {
	for _, course := range collapseCourses(njitCourses) {
		newCourse := &model.Course{}
		courseSection := course[0]
		newCourse.Name = courseSection.CourseNumber
		newCourse.Number = courseSection.CourseNumber
		newCourse.Sections = buildSections(course)
		c = append(c, newCourse)
	}

	return
}

func buildSections(njitCourses []NCourse) (s []*model.Section) {
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

func buildMeetings(njitCourse NCourse) (meetings []*model.Meeting) {
	for _, val := range njitCourse.meetingTimes {
		newMeeting := &model.Meeting{
			StartTime: &val.BeginTime,
			EndTime: &val.EndTime,
			Room: &val.Room,
			ClassType: &val.MeetingScheduleType,
			Day: &val.day,
		}

		meetings = append(meetings, newMeeting)
	}

	return
}

func buildInstructors(njitCourse NCourse) (instructors []*model.Instructor) {
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
