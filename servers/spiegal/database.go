package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"time"
	uct "uct/common"
)

var (
	database      *sqlx.DB
	preparedStmts = make(map[string]*sqlx.NamedStmt)
)

type Data struct {
	Data []byte `db:"data"`
}


func GetSemesters(topicName string, university *uct.University) error {
	if s, err := SelectAvailableSemesters(topicName); err != nil {
		return err
	} else {
		university.AvailableSemesters = s
		university.Metadata, err = SelectMetadata(university.Id, 0, 0, 0, 0)
		return err
	}
}

func SelectUniversity(topicName string) (university uct.University, err error) {
	defer uct.TimeTrack(time.Now(), "SelectUniversity")
	m := map[string]interface{}{"topic_name": topicName}
	if err = Get(SelectUniversityQuery, &university, m); err != nil {
		return
	}
	err = GetSemesters(topicName, &university)
	return
}

func SelectUniversities() (universities []*uct.University, err error) {
	m := map[string]interface{}{}
	if err = Select(ListUniversitiesQuery, &universities, m); err != nil {
		return
	}

	for i := range universities {
		err = GetSemesters(universities[i].TopicName, universities[i])
	}
	return
}

func SelectAvailableSemesters(topicName string) (semesters []*uct.Semester, err error) {
	defer uct.TimeTrack(time.Now(), "GetAvailableSemesters")
	m := map[string]interface{}{"topic_name": topicName}
	err = Select(GetAvailableSemestersQuery, &semesters, m)
	return
}

func SelectSubject(subjectTopicName string) (subject uct.Subject, b []byte, err error) {
	defer uct.TimeTrack(time.Now(), "SelectProtoSubject")
	m := map[string]interface{}{"topic_name": subjectTopicName}
	d := Data{}
	if err = Get(SelectProtoSubjectQuery, &d, m); err != nil {
		return
	}
	b = d.Data
	err = subject.Unmarshal(d.Data)
	return
}

func SelectSubjects(uniTopicName, season, year string) (subjects []*uct.Subject, err error) {
	defer uct.TimeTrack(time.Now(), "SelectSubjects")
	m := map[string]interface{}{"topic_name": uniTopicName, "subject_season": season, "subject_year": year}
	err = Select(ListSubjectQuery, &subjects, m)
	return
}

func SelectCourse(courseTopicName string) (course uct.Course, b []byte, err error) {
	defer uct.TimeTrack(time.Now(), "SelectCourse")
	d := Data{}
	m := map[string]interface{}{"topic_name": courseTopicName}
	if err = Get(SelectCourseQuery, &d, m); err != nil {
		return
	}
	b = d.Data
	err = course.Unmarshal(b)
	return
}

func SelectCourses(subjectTopicName string) (courses []*uct.Course, err error) {
	defer uct.TimeTrack(time.Now(), "SelectCourses")
	d := []Data{}
	m := map[string]interface{}{"topic_name": subjectTopicName}

	log.WithFields(log.Fields{"query": ListCoursesQuery, "args": m}).Debug()
	if err = Select(ListCoursesQuery, &d, m); err != nil {
		return
	}

	for i := range d {
		c := uct.Course{}
		if err = c.Unmarshal(d[i].Data); err != nil {
			return
		}
		courses = append(courses, &c)
	}

	return
}

func SelectSection(sectionTopicName string) (section uct.Section, err error) {
	defer uct.TimeTrack(time.Now(), "SelectSection")

	m := map[string]interface{}{"topic_name": sectionTopicName}

	if err = Get(SelectSectionQuery, &section, m); err != nil {
		return
	}

	err = deepSelectSection(&section)
	return
}

func deepSelectSection(section *uct.Section) (err error) {
	section.Meetings, err = SelectMeetings(section.Id)
	section.Books, err = SelectBooks(section.Id)
	section.Instructors, err = SelectInstructors(section.Id)
	section.Metadata, err = SelectMetadata(0, 0, 0, section.Id, 0)
	return
}

func SelectMeetings(sectionId int64) (meetings []*uct.Meeting, err error) {
	defer uct.TimeTrack(time.Now(), "SelectMeetings")
	m := map[string]interface{}{"section_id": sectionId}
	if err = Select(SelectMeeting, &meetings, m); err != nil {
		return
	}
	for i := range meetings {
		if meetings[i].Metadata, err = SelectMetadata(0, 0, 0, 0, meetings[i].Id); err != nil {
			return
		}
	}
	return
}

func SelectInstructors(sectionId int64) (instructors []*uct.Instructor, err error) {
	defer uct.TimeTrack(time.Now(), "SelectInstructors")
	m := map[string]interface{}{"section_id": sectionId}
	err = Select(SelectInstructor, &instructors, m)
	return
}

func SelectBooks(sectionId int64) (books []*uct.Book, err error) {
	defer uct.TimeTrack(time.Now(), "SelectInstructors")
	m := map[string]interface{}{"section_id": sectionId}
	err = Select(SelectBook, &books, m)
	return
}

func SelectMetadata(universityId, subjectId, courseId, sectionId, meetingId int64) (metadata []*uct.Metadata, err error) {
	defer uct.TimeTrack(time.Now(), "SelectMetadata")
	m := map[string]interface{}{
		"university_id": universityId,
		"subject_id":    subjectId,
		"course_id":     courseId,
		"section_id":    sectionId,
		"meeting_id":    meetingId,
	}
	if universityId != 0 {
		err = Select(UniversityMetadataQuery, &metadata, m)
	} else if subjectId != 0 {
		err = Select(SubjectMetadataQuery, &metadata, m)
	} else if courseId != 0 {
		err = Select(CourseMetadataQuery, &metadata, m)
	} else if sectionId != 0 {
		err = Select(SectionMetadataQuery, &metadata, m)
	} else if meetingId != 0 {
		err = Select(MeetingMetadataQuery, &metadata, m)
	}
	return
}

func Select(query string, dest interface{}, args interface{}) error {
	if err := GetCachedStmt(query).Select(dest, args); err != nil {
		return err
	}
	return nil
}

func Get(query string, dest interface{}, args interface{}) error {
	if err := GetCachedStmt(query).Get(dest, args); err != nil {
		return err
	}
	return nil
}

func GetCachedStmt(query string) *sqlx.NamedStmt {
	if stmt := preparedStmts[query]; stmt == nil {
		preparedStmts[query] = Prepare(query)
	}
	return preparedStmts[query]
}

func Prepare(query string) *sqlx.NamedStmt {
	if named, err := database.PrepareNamed(query); err != nil {
		log.Panicln(fmt.Errorf("Error: %s Query: %s", query, err))
		return nil
	} else {
		return named
	}
}

func PrepareAllStmts() {
	queries := []string{
		SelectUniversityQuery,
		ListUniversitiesQuery,
		GetAvailableSemestersQuery,
		SelectProtoSubjectQuery,
		ListSubjectQuery,
		SelectCourseQuery,
		ListCoursesQuery,
		SelectSectionQuery,
		SelectMeeting,
		SelectInstructor,
		SelectBook,
		UniversityMetadataQuery,
		SubjectMetadataQuery,
		CourseMetadataQuery,
		SectionMetadataQuery,
		MeetingMetadataQuery,
	}

	for _, query := range queries {
		preparedStmts[query] = Prepare(query)
	}
}

var (
	SelectUniversityQuery      = `SELECT id, name, abbr, home_page, registration_page, main_color, accent_color, topic_name, topic_id FROM university WHERE topic_name = :topic_name ORDER BY name`
	ListUniversitiesQuery      = `SELECT id, name, abbr, home_page, registration_page, main_color, accent_color, topic_name, topic_id FROM university ORDER BY name`
	GetAvailableSemestersQuery = `SELECT season, year FROM subject JOIN university ON university.id = subject.university_id
									WHERE university.topic_name = :topic_name GROUP BY season, year`

	SelectProtoSubjectQuery = `SELECT data FROM subject WHERE topic_name = :topic_name`

	ListSubjectQuery = `SELECT subject.id, university_id, subject.name, subject.number, subject.season, subject.year, subject.topic_name, subject.topic_id FROM subject JOIN university ON university.id = subject.university_id
									AND university.topic_name = :topic_name
									AND season = :subject_season
									AND year = :subject_year ORDER BY subject.id`

	SelectCourseQuery = `SELECT data FROM course WHERE course.topic_name = :topic_name ORDER BY course.id`

	ListCoursesQuery = `SELECT course.data FROM course JOIN subject ON subject.id = course.subject_id WHERE subject.topic_name = :topic_name ORDER BY course.number`

	SelectSectionQuery = `SELECT id, course_id, number, call_number, now, max, status, credits, topic_name FROM section WHERE section.topic_name = :topic_name`

	SelectMeeting    = `SELECT section.id, section_id, room, day, start_time, end_time FROM meeting JOIN section ON section.id = meeting.section_id WHERE section_id = :section_id ORDER BY meeting.id`
	SelectInstructor = `SELECT name FROM instructor WHERE section_id = :section_id ORDER BY index`
	SelectBook       = `SELECT title, url FROM book WHERE section_id = :section_id`

	UniversityMetadataQuery = `SELECT title, content FROM metadata WHERE university_id = :university_id ORDER BY id`
	SubjectMetadataQuery    = `SELECT title, content FROM metadata WHERE subject_id = :subject_id ORDER BY id`
	CourseMetadataQuery     = `SELECT title, content FROM metadata WHERE course_id = :course_id ORDER BY id`
	SectionMetadataQuery    = `SELECT title, content FROM metadata WHERE section_id = :section_id ORDER BY id`
	MeetingMetadataQuery    = `SELECT title, content FROM metadata WHERE meeting_id = :meeting_id ORDER BY id`
)
