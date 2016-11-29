package main

import (
	"database/sql"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"sort"
	"strconv"
	"time"
	"uct/common/model"
	"uct/spike/middleware"
)

var (
	database      *sqlx.DB
	preparedStmts = make(map[string]*sqlx.NamedStmt)
)

type Data struct {
	Data []byte `db:"data"`
}

func SelectUniversity(topicName string) (university model.University, err error) {
	defer model.TimeTrack(time.Now(), "SelectUniversity")
	m := map[string]interface{}{"topic_name": topicName}
	if err = Get(SelectUniversityQuery, &university, m); err != nil {
		return
	}
	if err = GetAvailableSemesters(topicName, &university); err != nil {
		return
	}
	if err = GetResolvedSemesters(topicName, &university); err != nil {
		return
	}
	return
}

func SelectUniversities() (universities []*model.University, err error) {
	m := map[string]interface{}{}
	if err = Select(ListUniversitiesQuery, &universities, m); err != nil {
		return
	}
	if err == nil && len(universities) == 0 {
		err = middleware.ErrNoRows{Uri: "No data found a list of universities"}
	}

	for i := range universities {
		if err = GetAvailableSemesters(universities[i].TopicName, universities[i]); err != nil {
			return
		}

		if err = GetResolvedSemesters(universities[i].TopicName, universities[i]); err != nil {
			return
		}
	}

	return
}

func GetResolvedSemesters(topicName string, university *model.University) error {
	if r, err := SelectResolvedSemesters(topicName); err != nil {
		return err
	} else {
		university.ResolvedSemesters = &r
		return err
	}
}

func GetAvailableSemesters(topicName string, university *model.University) error {
	if s, err := SelectAvailableSemesters(topicName); err != nil {
		return err
	} else {
		university.AvailableSemesters = s
		university.Metadata, err = SelectMetadata(university.Id, 0, 0, 0, 0)
		return err
	}
}

func SelectAvailableSemesters(topicName string) (semesters []*model.Semester, err error) {
	defer model.TimeTrack(time.Now(), "GetAvailableSemesters")
	m := map[string]interface{}{"topic_name": topicName}
	err = Select(SelectAvailableSemestersQuery, &semesters, m)
	sort.Sort(model.SemesterSorter(semesters))
	return
}

func SelectResolvedSemesters(topicName string) (semesters model.ResolvedSemester, err error) {
	defer model.TimeTrack(time.Now(), "SelectResolvedSemesters")
	m := map[string]interface{}{"topic_name": topicName}
	rs := model.DBResolvedSemester{}
	if err = Get(SelectResolvedSemestersQuery, &rs, m); err != nil {
		return
	}
	curr, _ := strconv.ParseInt(rs.CurrentYear, 10, 32)
	last, _ := strconv.ParseInt(rs.LastYear, 10, 32)
	next, _ := strconv.ParseInt(rs.NextYear, 10, 32)
	semesters.Current = &model.Semester{Year: int32(curr), Season: rs.CurrentSeason}
	semesters.Last = &model.Semester{Year: int32(last), Season: rs.LastSeason}
	semesters.Next = &model.Semester{Year: int32(next), Season: rs.NextSeason}
	return
}

func SelectSubject(subjectTopicName string) (subject model.Subject, b []byte, err error) {
	defer model.TimeTrack(time.Now(), "SelectProtoSubject")
	m := map[string]interface{}{"topic_name": subjectTopicName}
	d := Data{}
	if err = Get(SelectProtoSubjectQuery, &d, m); err != nil {
		return
	}
	b = d.Data
	err = subject.Unmarshal(d.Data)
	return
}

func SelectSubjects(uniTopicName, season, year string) (subjects []*model.Subject, err error) {
	defer model.TimeTrack(time.Now(), "SelectSubjects")
	m := map[string]interface{}{"topic_name": uniTopicName, "subject_season": season, "subject_year": year}
	err = Select(ListSubjectQuery, &subjects, m)
	if err == nil && len(subjects) == 0 {
		err = middleware.ErrNoRows{Uri: fmt.Sprintf("No data subjects found for university=%s, season=%s, year=%s", uniTopicName, season, year)}
	}
	return
}

func SelectCourse(courseTopicName string) (course model.Course, b []byte, err error) {
	defer model.TimeTrack(time.Now(), "SelectCourse")
	d := Data{}
	m := map[string]interface{}{"topic_name": courseTopicName}
	if err = Get(SelectCourseQuery, &d, m); err != nil {
		return
	}
	b = d.Data
	err = course.Unmarshal(b)
	return
}

func SelectCourses(subjectTopicName string) (courses []*model.Course, err error) {
	defer model.TimeTrack(time.Now(), "SelectCourses")
	d := []Data{}
	m := map[string]interface{}{"topic_name": subjectTopicName}
	if err = Select(ListCoursesQuery, &d, m); err != nil {
		return
	}
	if err == nil && len(courses) == 0 {
		err = middleware.ErrNoRows{Uri: fmt.Sprintf("No courses found for %s", subjectTopicName)}
	}
	for i := range d {
		c := model.Course{}
		if err = c.Unmarshal(d[i].Data); err != nil {
			return
		}
		courses = append(courses, &c)
	}

	return
}

func SelectSection(sectionTopicName string) (section model.Section, b []byte, err error) {
	defer model.TimeTrack(time.Now(), "SelectSection")
	d := Data{}
	m := map[string]interface{}{"topic_name": sectionTopicName}
	if err = Get(SelectProtoSectionQuery, &d, m); err != nil {
		return
	}
	b = d.Data
	err = section.Unmarshal(b)
	return
}

func SelectMetadata(universityId, subjectId, courseId, sectionId, meetingId int64) (metadata []*model.Metadata, err error) {
	defer model.TimeTrack(time.Now(), "SelectMetadata")
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
		if err == sql.ErrNoRows {
			err = middleware.ErrNoRows{Uri: err.Error()}
		}
		return err
	}
	return nil
}

func Get(query string, dest interface{}, args interface{}) error {
	if err := GetCachedStmt(query).Get(dest, args); err != nil {
		if err == sql.ErrNoRows {
			err = middleware.ErrNoRows{Uri: err.Error()}
		}
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

func prepareAllStmts() {
	queries := []string{
		SelectUniversityQuery,
		ListUniversitiesQuery,
		SelectAvailableSemestersQuery,
		SelectResolvedSemestersQuery,
		SelectProtoSubjectQuery,
		SelectProtoSectionQuery,
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
	SelectUniversityQuery         = `SELECT id, name, abbr, home_page, registration_page, main_color, accent_color, topic_name, topic_id FROM university WHERE topic_name = :topic_name ORDER BY name`
	ListUniversitiesQuery         = `SELECT id, name, abbr, home_page, registration_page, main_color, accent_color, topic_name, topic_id FROM university ORDER BY name`
	SelectAvailableSemestersQuery = `SELECT season, year FROM subject JOIN university ON university.id = subject.university_id
									WHERE university.topic_name = :topic_name GROUP BY season, year`

	SelectResolvedSemestersQuery = `SELECT current_season, current_year, last_season, last_year, next_season, next_year FROM semester JOIN university ON university.id = semester.university_id
	WHERE university.topic_name = :topic_name`

	SelectProtoSubjectQuery = `SELECT data FROM subject WHERE topic_name = :topic_name`

	SelectProtoSectionQuery = `SELECT data FROM section WHERE topic_name = :topic_name`

	ListSubjectQuery = `SELECT subject.id, university_id, subject.name, subject.number, subject.season, subject.year, subject.topic_name, subject.topic_id FROM subject JOIN university ON university.id = subject.university_id
									AND university.topic_name = :topic_name
									AND season = :subject_season
									AND year = :subject_year ORDER BY subject.name`

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
