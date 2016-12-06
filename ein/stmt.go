package main

var queries = []string{UniversityInsertQuery,
	UniversityUpdateQuery,
	SemesterInsertQuery,
	SemesterUpdateQuery,
	SubjectExistQuery,
	SubjectInsertQuery,
	SubjectUpdateQuery,
	CourseUpdateQuery,
	CourseExistQuery,
	CourseInsertQuery,
	SectionInsertQuery,
	SectionUpdateQuery,
	MeetingUpdateQuery,
	MeetingInsertQuery,
	MeetingExistQuery,
	InstructorExistQuery,
	InstructorUpdateQuery,
	InstructorInsertQuery,
	BookUpdateQuery,
	BookInsertQuery,
	RegistrationUpdateQuery,
	RegistrationInsertQuery,
	MetaUniExistQuery,
	MetaUniUpdateQuery,
	MetaUniInsertQuery,
	MetaSubjectExistQuery,
	MetaSubjectUpdateQuery,
	MetaSubjectInsertQuery,
	MetaCourseExistQuery,
	MetaCourseUpdateQuery,
	MetaCourseInsertQuery,
	MetaSectionExistQuery,
	MetaSectionInsertQuery,
	MetaSectionUpdateQuery,
	MetaSectionExistQuery,
	MetaMeetingInsertQuery,
	MetaMeetingUpdateQuery,
	SerialSubjectUpdateQuery,
	SerialCourseUpdateQuery,
	SerialSectionUpdateQuery,
}

const (
	UniversityInsertQuery = `INSERT INTO university (name, abbr, home_page, registration_page, main_color, accent_color, topic_name, topic_id)
                    VALUES (:name, :abbr, :home_page, :registration_page, :main_color, :accent_color, :topic_name, :topic_id)
                    RETURNING university.id`
	UniversityUpdateQuery = `UPDATE university SET (abbr, home_page, registration_page, main_color, accent_color, topic_name, topic_id) =
	                (:abbr, :home_page, :registration_page, :main_color, :accent_color, :topic_name, :topic_id)
	                WHERE name = :name
	                RETURNING university.id`

	SemesterInsertQuery = `INSERT INTO semester (university_id, current_season, current_year, last_season, last_year, next_season, next_year)
							VALUES (:university_id, :current_season, :current_year, :last_season, :last_year, :next_season, :next_year) RETURNING semester.id`
	SemesterUpdateQuery = `UPDATE semester SET (current_season, current_year, last_season, last_year, next_season, next_year) =
						(:current_season, :current_year, :last_season, :last_year, :next_season, :next_year) WHERE university_id = :university_id RETURNING semester.id`

	SubjectExistQuery = `SELECT subject.id FROM subject WHERE topic_name = :topic_name`

	SubjectInsertQuery = `INSERT INTO subject (university_id, name, number, season, year, topic_name, topic_id)
                   	VALUES  (:university_id, :name, :number, :season, :year, :topic_name, :topic_id)
                   	RETURNING subject.id`

	SubjectUpdateQuery = SubjectExistQuery

	CourseInsertQuery = `INSERT INTO course (subject_id, name, number, synopsis, topic_name, topic_id) VALUES  (:subject_id, :name, :number, :synopsis, :topic_name, :topic_id) RETURNING course.id`
	CourseExistQuery  = `SELECT course.id FROM course WHERE topic_name = :topic_name`
	CourseUpdateQuery = `UPDATE course SET (synopsis) = (:synopsis) WHERE topic_name = :topic_name RETURNING course.id`

	SectionInsertQuery = `INSERT INTO section (course_id, number, call_number, max, now, status, credits, topic_name, topic_id)
                    VALUES (:course_id, :number, :call_number, :max, :now, :status, :credits, :topic_name, :topic_id)
                    RETURNING section.id`

	SectionUpdateQuery = `UPDATE section SET (max, now, status, credits) = (:max, :now, :status, :credits) WHERE topic_name = :topic_name RETURNING section.id`

	MeetingExistQuery = `SELECT id FROM meeting WHERE section_id = :section_id AND index = :index`

	MeetingUpdateQuery = `UPDATE meeting SET (room, day, start_time, end_time, class_type) = (:room, :day, :start_time, :end_time, :class_type)
					WHERE section_id = :section_id AND index = :index
                    RETURNING meeting.id`

	MeetingInsertQuery = `INSERT INTO meeting (section_id, room, day, start_time, end_time, class_type, index)
                    VALUES  (:section_id, :room, :day, :start_time, :end_time, :class_type, :index)
                    RETURNING meeting.id`

	InstructorExistQuery = `SELECT id FROM instructor
				WHERE section_id = :section_id AND index = :index`

	InstructorUpdateQuery = `UPDATE instructor SET name = :name WHERE section_id = :section_id AND index = :index
				RETURNING instructor.id`

	InstructorInsertQuery = `INSERT INTO instructor (section_id, name, index)
				VALUES  (:section_id, :name, :index)
				RETURNING instructor.id`

	BookUpdateQuery = `UPDATE book SET (title, url) = (:title, :url)
					WHERE section_id = :section_id AND title = :title
                    RETURNING book.id`

	BookInsertQuery = `INSERT INTO book (section_id, title, url)
                    VALUES  (:section_id, :title, :url)
                    RETURNING book.id`

	RegistrationUpdateQuery = `UPDATE registration SET (period_date) = (:period_date)
					WHERE university_id = :university_id AND period = :period
	                RETURNING registration.id`

	RegistrationInsertQuery = `INSERT INTO registration (university_id, period, period_date)
                    VALUES (:university_id, :period, :period_date)
                    RETURNING registration.id`

	MetaUniExistQuery = `SELECT id FROM metadata
						WHERE metadata.university_id = :university_id AND metadata.title = :title`
	MetaUniUpdateQuery = `UPDATE metadata SET (title, content) = (:title, :content)
					   WHERE metadata.university_id = :university_id AND metadata.title = :title
		               RETURNING metadata.id`
	MetaUniInsertQuery = `INSERT INTO metadata (university_id, title, content)
                       VALUES (:university_id, :title, :content)
                       RETURNING metadata.id`

	MetaSubjectExistQuery = `SELECT id FROM metadata
						WHERE metadata.subject_id = :subject_id AND metadata.title = :title`
	MetaSubjectUpdateQuery = `UPDATE metadata SET (title, content) = (:title, :content)
					   WHERE metadata.subject_id = :subject_id AND metadata.title = :title
		               RETURNING metadata.id`
	MetaSubjectInsertQuery = `INSERT INTO metadata (subject_id, title, content)
                       VALUES (:subject_id, :title, :content)
                       RETURNING metadata.id`

	MetaCourseExistQuery = `SELECT id FROM metadata
						WHERE metadata.course_id = :course_id AND metadata.title = :title`

	MetaCourseUpdateQuery = `UPDATE metadata SET (title, content) = (:title, :content)
					   WHERE metadata.course_id = :course_id AND metadata.title = :title
		               RETURNING metadata.id`

	MetaCourseInsertQuery = `INSERT INTO metadata (course_id, title, content)
                       VALUES (:course_id, :title, :content)
                       RETURNING metadata.id`

	MetaSectionExistQuery = `SELECT id FROM metadata
						WHERE metadata.section_id = :section_id AND metadata.title = :title`
	MetaSectionInsertQuery = `INSERT INTO metadata (section_id, title, content)
                       VALUES (:section_id, :title, :content)
                       RETURNING metadata.id`
	MetaSectionUpdateQuery = `UPDATE metadata SET (title, content) = (:title, :content)
					   WHERE metadata.section_id = :section_id AND metadata.title = :title
		               RETURNING metadata.id`

	MetaMeetingExistQuery = `SELECT id FROM metadata
						WHERE metadata.meeting_id = :meeting_id AND metadata.title = :title`
	MetaMeetingInsertQuery = `INSERT INTO metadata (meeting_id, title, content)
                       VALUES (:meeting_id, :title, :content)
                       RETURNING metadata.id`
	MetaMeetingUpdateQuery = `UPDATE metadata SET (title, content) = (:title, :content)
					   WHERE metadata.meeting_id = :meeting_id AND metadata.title = :title
		               RETURNING metadata.id`

	SerialSubjectUpdateQuery = `UPDATE subject SET data = :data WHERE topic_name = :topic_name RETURNING subject.id`
	SerialCourseUpdateQuery  = `UPDATE course SET data = :data WHERE topic_name = :topic_name RETURNING course.id`
	SerialSectionUpdateQuery = `UPDATE section SET data = :data WHERE topic_name = :topic_name RETURNING section.id`
)
