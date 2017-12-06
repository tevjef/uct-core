package store

import (
	"github.com/tevjef/uct-core/common/database"
	mtrace "github.com/tevjef/uct-core/spike/middleware/trace"

	"golang.org/x/net/context"
)

type Data struct {
	Data []byte `db:"data"`
}

func Select(ctx context.Context, query string, dest interface{}, args interface{}) error {
	span := mtrace.NewSpan(ctx, "database.Select")
	span.SetLabel("query", query)
	defer span.Finish()

	if err := database.FromContext(ctx).Select(query, dest, args); err != nil {
		return err
	}
	return nil
}

func Get(ctx context.Context, query string, dest interface{}, args interface{}) error {
	span := mtrace.NewSpan(ctx, "database.Get")
	span.SetLabel("query", query)
	defer span.Finish()

	if err := database.FromContext(ctx).Get(query, dest, args); err != nil {
		return err
	}
	return nil
}

func Insert(ctx context.Context, query string, data interface{}) error {
	span := mtrace.NewSpan(ctx, "database.Insert")
	span.SetLabel("query", query)
	defer span.Finish()

	database.FromContext(ctx).Insert(query, data)
	return nil
}

var Queries = []string{
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
	InsertSubscriptionQuery,
	InsertNotificationQuery,
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

	InsertSubscriptionQuery = `INSERT INTO subscription (topic_name, fcm_token, is_subscribed)
                    VALUES  (:topic_name, :fcm_token, :is_subscribed)
                    RETURNING subscription.id`

	InsertNotificationQuery = `INSERT INTO acknowledge (topic_name, receive_at)
                    VALUES  (:topic_name, :receive_at)
                    RETURNING acknowledge.id`

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
