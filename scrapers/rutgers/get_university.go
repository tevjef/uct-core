package rutgers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/conf"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/common/try"
	"gopkg.in/alecthomas/kingpin.v2"
)

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

type rutgersRequest struct {
	semester model.Semester
	host     string
	campus   string
}
type rutgers struct {
	app    *kingpin.ApplicationModel
	config *rutgersConfig
	ctx    context.Context
}

type rutgersConfig struct {
	service       conf.Config
	campus        string
	outputHttpUrl string
	outputFormat  string
	latest        bool
}

func (rutgers *rutgers) getCampus(campus string) model.University {
	var university model.University

	university = getRutgers(campus)

	// Rutgers servers go down for maintenance between 2 and 5 AM UTC every day.
	// Data scraped from this time would be inaccurate and may lead to unforeseen errors
	if currentHour := time.Now().In(time.FixedZone("EST", -18000)).Hour(); currentHour >= 2 && currentHour < 5 {
		return university
	}

	university.ResolvedSemesters = model.ResolveSemesters(time.Now(), university.Registrations)

	semesters := []*model.Semester{
		university.ResolvedSemesters.Last,
		university.ResolvedSemesters.Current,
		university.ResolvedSemesters.Next}

	if rutgers.config.latest {
		semesters = semesters[1:]
	}

	for _, semester := range semesters {
		if semester.Season == model.Winter {
			semester.Year++
		}

		subjects := getSubjects(campus, semester)

		university.Subjects = append(university.Subjects, buildSubjects(subjects)...)
	}

	return university
}

func getSubjects(campus string, semester *model.Semester) []*RSubject {
	rr := rutgersRequest{
		host:     "http://sis.rutgers.edu/soc/api",
		semester: *semester,
		campus:   campus,
	}

	subjects := rr.requestCourses()

	return subjects
}

func (sr rutgersRequest) requestCourses() (subjects []*RSubject) {
	var url = fmt.Sprintf(
		"%s/courses.gzip?year=%s&term=%s&campus=%s",
		sr.host,
		parseYear(sr.semester),
		parseTerm(sr.semester),
		sr.campus,
	)

	var courses []*RCourseNew

	if err := getData(url, &courses); err == nil {
		subjectsMap := map[string]*RSubject{}
		for i := range courses {
			course := courses[i]
			if value, ok := subjectsMap[course.Subject]; ok {
				value.Courses = append(value.Courses, course.buildCourse())
			} else {
				var coursesSlice []*RCourse
				subject := &RSubject{
					Name:    course.SubjectDescription,
					Number:  course.Subject,
					Season:  sr.semester.Season,
					Courses: append(coursesSlice, course.buildCourse()),
					Year:    int(sr.semester.Year),
				}
				subject.clean()
				subjectsMap[course.Subject] = subject
			}
		}

		for _, value := range subjectsMap {
			subjects = append(subjects, value)
		}
	}

	return
}

func parseTerm(semester model.Semester) string {
	if semester.Season == model.Fall {
		return "9"
	} else if semester.Season == model.Summer {
		return "7"
	} else if semester.Season == model.Spring {
		return "1"
	} else if semester.Season == model.Winter {
		return "0"
	}
	return ""
}

func parseYear(semester model.Semester) string {
	return strconv.Itoa(int(semester.Year))
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

		for i := range section.MeetingTimes {
			meeting := section.MeetingTimes[i]
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
	//req.Header.Add("User-Agent", "Go/rutgers-scrape")

	fields := log.WithFields(log.Fields{"url": url, "model_type": fmt.Sprintf("%T", model)})
	fields.Debugln()

	err := try.Do(func(attempt int) (bool, error) {
		startTime := time.Now()
		time.Sleep(time.Duration((attempt-1)*2) * time.Second)
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
			"content-length":  len(data),
			"response_status": resp.StatusCode,
			"response_time":   time.Since(startTime).Seconds()}).Debugln()

		return false, nil
	})

	return err
}
