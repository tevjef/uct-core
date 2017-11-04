package main

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/tevjef/uct-core/common/database"
	"github.com/tevjef/uct-core/common/model"
	"github.com/tevjef/uct-core/spike/middleware"

	"sync"

	"github.com/pquerna/ffjson/ffjson"
	"golang.org/x/net/context"
)

type Data struct {
	Data []byte `db:"data"`
}

func SelectUniversity(ctx context.Context, topicName string) (university model.University, err error) {
	defer model.TimeTrack(time.Now(), "SelectUniversity")
	m := map[string]interface{}{"topic_name": topicName}
	d := Data{}
	if err = Get(ctx, SelectUniversityCTE, &d, m); err != nil {
		return
	}

	if err = ffjson.Unmarshal([]byte(d.Data), &university); err != nil {
		return
	}
	return
}

func SelectUniversities(ctx context.Context) (universities []*model.University, err error) {
	var topics []string
	m := map[string]interface{}{}
	if err = Select(ctx, ListUniversitiesQuery, &topics, m); err != nil {
		return
	}

	if err == nil && len(topics) == 0 {
		err = middleware.ErrNoRows{Uri: "No data found a list of universities"}
	}

	uniChan := make(chan model.University)
	go func() {
		var wg sync.WaitGroup
		for i := range topics {
			wg.Add(1)
			u := topics[i]
			go func() {
				defer wg.Done()
				uni, err := SelectUniversity(ctx, u)
				if err != nil {
					panic(err)
				}
				uniChan <- uni
			}()
		}
		wg.Wait()
		close(uniChan)
	}()

	for uni := range uniChan {
		u := uni
		universities = append(universities, &u)
	}

	sort.Sort(model.UniversityByName(universities))
	return
}

func GetResolvedSemesters(ctx context.Context, topicName string, university *model.University) error {
	if r, err := SelectResolvedSemesters(ctx, topicName); err != nil {
		return err
	} else {
		university.ResolvedSemesters = &r
		return err
	}
}

func GetAvailableSemesters(ctx context.Context, topicName string, university *model.University) error {
	if s, err := SelectAvailableSemesters(ctx, topicName); err != nil {
		return err
	} else {
		university.AvailableSemesters = s
		return err
	}
}

func SelectAvailableSemesters(ctx context.Context, topicName string) (semesters []*model.Semester, err error) {
	defer model.TimeTrack(time.Now(), "GetAvailableSemesters")
	m := map[string]interface{}{"topic_name": topicName}
	err = Select(ctx, SelectAvailableSemestersQuery, &semesters, m)
	sort.Sort(model.SemesterSorter(semesters))
	return
}

func SelectResolvedSemesters(ctx context.Context, topicName string) (semesters model.ResolvedSemester, err error) {
	defer model.TimeTrack(time.Now(), "SelectResolvedSemesters")
	m := map[string]interface{}{"topic_name": topicName}
	rs := model.DBResolvedSemester{}
	if err = Get(ctx, SelectResolvedSemestersQuery, &rs, m); err != nil {
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

func SelectSubject(ctx context.Context, subjectTopicName string) (subject model.Subject, b []byte, err error) {
	defer model.TimeTrack(time.Now(), "SelectProtoSubject")
	m := map[string]interface{}{"topic_name": subjectTopicName}
	d := Data{}
	if err = Get(ctx, SelectProtoSubjectQuery, &d, m); err != nil {
		return
	}
	b = d.Data
	err = subject.Unmarshal(d.Data)
	return
}

func SelectSubjects(ctx context.Context, uniTopicName, season, year string) (subjects []*model.Subject, err error) {
	defer model.TimeTrack(time.Now(), "SelectSubjects")
	m := map[string]interface{}{"topic_name": uniTopicName, "subject_season": season, "subject_year": year}
	err = Select(ctx, ListSubjectQuery, &subjects, m)
	if err == nil && len(subjects) == 0 {
		err = middleware.ErrNoRows{Uri: fmt.Sprintf("No data subjects found for university=%s, season=%s, year=%s", uniTopicName, season, year)}
	}
	return
}

func SelectCourse(ctx context.Context, courseTopicName string) (course model.Course, b []byte, err error) {
	defer model.TimeTrack(time.Now(), "SelectCourse")
	d := Data{}
	m := map[string]interface{}{"topic_name": courseTopicName}
	if err = Get(ctx, SelectCourseQuery, &d, m); err != nil {
		return
	}
	b = d.Data
	err = course.Unmarshal(b)
	return
}

func SelectCourses(ctx context.Context, subjectTopicName string) (courses []*model.Course, err error) {
	defer model.TimeTrack(time.Now(), "SelectCourses")
	d := []Data{}
	m := map[string]interface{}{"topic_name": subjectTopicName}
	if err = Select(ctx, ListCoursesQuery, &d, m); err != nil {
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

func SelectSection(ctx context.Context, sectionTopicName string) (section model.Section, b []byte, err error) {
	defer model.TimeTrack(time.Now(), "SelectSection")
	d := Data{}
	m := map[string]interface{}{"topic_name": sectionTopicName}
	if err = Get(ctx, SelectProtoSectionQuery, &d, m); err != nil {
		return
	}
	b = d.Data
	err = section.Unmarshal(b)
	return
}

func SelectMetadata(ctx context.Context, universityId, subjectId, courseId, sectionId, meetingId int64) (metadata []*model.Metadata, err error) {
	defer model.TimeTrack(time.Now(), "SelectMetadata")
	m := map[string]interface{}{
		"university_id": universityId,
		"subject_id":    subjectId,
		"course_id":     courseId,
		"section_id":    sectionId,
		"meeting_id":    meetingId,
	}
	if universityId != 0 {
		err = Select(ctx, UniversityMetadataQuery, &metadata, m)
	} else if subjectId != 0 {
		err = Select(ctx, SubjectMetadataQuery, &metadata, m)
	} else if courseId != 0 {
		err = Select(ctx, CourseMetadataQuery, &metadata, m)
	} else if sectionId != 0 {
		err = Select(ctx, SectionMetadataQuery, &metadata, m)
	} else if meetingId != 0 {
		err = Select(ctx, MeetingMetadataQuery, &metadata, m)
	}
	return
}

func Select(ctx context.Context, query string, dest interface{}, args interface{}) error {
	if err := database.FromContext(ctx).Select(query, dest, args); err != nil {
		if err == sql.ErrNoRows {
			err = middleware.ErrNoRows{Uri: err.Error()}
		}
		return err
	}
	return nil
}

func Get(ctx context.Context, query string, dest interface{}, args interface{}) error {
	if err := database.FromContext(ctx).Get(query, dest, args); err != nil {
		if err == sql.ErrNoRows {
			err = middleware.ErrNoRows{Uri: err.Error()}
		}
		return err
	}
	return nil
}

var queries = []string{
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
	SelectUniversityCTE,
}

const (
	SelectUniversityQuery         = `SELECT id, name, abbr, home_page, registration_page, main_color, accent_color, topic_name, topic_id FROM university WHERE topic_name = :topic_name ORDER BY name`
	ListUniversitiesQuery         = `SELECT topic_name FROM university ORDER BY name`
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

	SelectUniversityCTE = `WITH resolved_semesters AS (
    SELECT json_build_object(
        'current', json_build_object(
            'year', cast(s.current_year as INT),
            'season', s.current_season
        ),
        'next', json_build_object(
            'year', cast(s.next_year as INT),
            'season', s.next_season
        ),
        'last', json_build_object(
            'year', cast(s.last_year as INT),
            'season', s.last_season
        )
    )
    FROM semester s
      JOIN university ON university.id = s.university_id
    WHERE university.topic_name = :topic_name
), metadata AS (
    SELECT json_build_array(json_build_object(
        'title', m.title,
        'content', m.content))
    FROM metadata m
      LEFT JOIN university ON university.id = m.university_id
    WHERE university.topic_name = :topic_name
), available_semesters AS (
    SELECT array_to_json(array_agg(rawSemesters))
    FROM (SELECT
            s.season,
            cast(s.year as INT)
          FROM subject s
            JOIN university ON university.id = s.university_id
          WHERE university.topic_name = :topic_name
		  GROUP BY season, year
		  ORDER BY s.year DESC) rawSemesters
)
SELECT json_build_object(
    'name', u.name,
    'abbr', u.abbr,
    'home_page', u.home_page,
    'registration_page', u.registration_page,
    'main_color', u.main_color,
    'accent_color', u.accent_color,
    'topic_name', u.topic_name,
    'topic_id', u.topic_id,
    'available_semesters', (SELECT *
                           FROM available_semesters),
    'resolved_semesters', (SELECT *
                           FROM resolved_semesters),
    'metadata', (SELECT *
                 FROM metadata)
) as data
FROM university u
WHERE u.topic_name = :topic_name
GROUP BY u.id;
`
)
