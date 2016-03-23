package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	uct "uct/common"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"fmt"
	"reflect"
)

var database *sqlx.DB

func init() {
	var err error
	database, err = sqlx.Connect("postgres",
		fmt.Sprintf("postgres://%s:%s@%s:5432/%s",
			uct.DbUser, uct.DbPassword, uct.DbHost, uct.DbName))
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	decoder := json.NewDecoder(os.Stdin)

	var university uct.University

	for {
		if err := decoder.Decode(&university); err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("%+v", err)
		}
		insertUniversity(database, university)
	}

	defer database.Close()
}

func insertUniversity(db *sqlx.DB, uni uct.University) {
	uni.VetAndBuild()
	var university_id int64


	insertQuery := `INSERT INTO university (name, abbr, home_page, registration_page, main_color, accent_color, topic_name)
                    VALUES (:name, :abbr, :home_page, :registration_page, :main_color, :accent_color, :topic_name)
                    RETURNING university.id`

	updateQuery := `UPDATE university SET (abbr, home_page, registration_page, main_color, accent_color, topic_name) =
	                (:abbr, :home_page, :registration_page, :main_color, :accent_color, :topic_name)
	                WHERE name = :name
	                RETURNING university.id`

	university_id = upsert(db, insertQuery, updateQuery, uni)

	for _, subject := range uni.Subjects {
		subject.UniversityId = university_id
		subId := insertSubject(db, subject)

		for _, course := range subject.Courses {
			course.SubjectId = subId
			courseId := insertCourse(db, course)

			for _, section := range course.Sections {
				section.CourseId = courseId
				sectionId := insertSection(db, section)

				for _, instructor := range section.Instructors {
					instructor.SectionId = sectionId
					insertInstructor(db, instructor)
				}

				for _, meeting := range section.Meetings {
					meeting.SectionId = sectionId
					meetingId := insertMeeting(db, meeting)

					for _, metadata := range meeting.Metadata {
						metadata.MeetingId = meetingId
						insertMetadata(db, metadata)
					}
				}

				for _, book := range section.Books {
					book.SectionId = sectionId
					insertBook(db, book)
				}

				for _, metadata := range section.Metadata {
					metadata.SectionId = sectionId
					insertMetadata(db, metadata)
				}
			}

			for _, metadata := range course.Metadata {
				metadata.CourseId = courseId
				insertMetadata(db, metadata)
			}
		}

		for _, metadata := range subject.Metadata {
			metadata.SubjectId = subId
			insertMetadata(db, metadata)
		}
	}

	for _, registrations := range uni.Registrations {
		registrations.UniversityId = university_id
		insertRegistration(db, registrations)
	}

	for _, metadata := range uni.Metadata {
		metadata.UniversityId = university_id
		insertMetadata(db, metadata)
	}
}

func insertSubject(db *sqlx.DB, sub uct.Subject) (subject_id int64) {
	sub.VetAndBuild()

	insertQuery := `INSERT INTO subject (university_id, name, number, season, year, hash, topic_name)
                   	VALUES  (:university_id, :name, :number, :season, :year, :hash, :topic_name) 
                   	RETURNING subject.id`

	updateQuery := `UPDATE subject SET (name, season, year, topic_name) = (:name, :season, :year, :topic_name)
					WHERE hash = :hash
                   	RETURNING subject.id`
	
	subject_id = upsert(db, insertQuery, updateQuery, sub)
	
	return subject_id
}

func insertCourse(db *sqlx.DB, course uct.Course) (course_id int64) {
	course.VetAndBuild()
	
	updateQuery := `UPDATE course SET (name, synopsis, topic_name) = (:name, :synopsis, :topic_name) WHERE hash = :hash
                    RETURNING course.id`

	insertQuery := `INSERT INTO course (subject_id, name, number, synopsis, hash, topic_name)
                    VALUES  (:subject_id, :name, :number, :synopsis, :hash, :topic_name) 
                    RETURNING course.id`


	course_id = upsert(db, insertQuery, updateQuery, course)

	return course_id
}

func insertSection(db *sqlx.DB, section uct.Section) (section_id int64) {
	section.VetAndBuild()
	
	updateQuery := `UPDATE section SET (max, nox, status, credits, topic_name) = (:max, :now, :status, :creadits, :topic_name)
					WHERE course_id = :course_id
                    RETURNING section.id`

	insertQuery := `INSERT INTO section (course_id, number, call_number, max, nox, status, credits, topic_name)
                    VALUES  (:course_id, :number, :call_number, :max, :nox, :status, :credits, :topic_name) 
                    RETURNING section.id`


	section_id = upsert(db, insertQuery, updateQuery, section)

	return section_id
}

func insertMeeting(db *sqlx.DB, meeting uct.Meeting) (meeting_id int64) {
	meeting.VetAndBuild()
	
	updateQuery := `UPDATE meeting SET (room, day, start_time, end_time) = (:room, :day, :start_time, :end_time)
					WHERE section_id = :section_id
                    RETURNING meeting.id`

	insertQuery := `INSERT INTO meeting (section_id, room, day, start_time, end_time)
                    VALUES  (:section_id, :room, :day, :start_time, :end_time)
                    RETURNING meeting.id`


	meeting_id = upsert(db, insertQuery, updateQuery, meeting)

	return meeting_id
}

func insertInstructor(db *sqlx.DB, instructor uct.Instructor) (instructor_id int64) {
	instructor.VetAndBuild()

	updateQuery := `UPDATE instructor SET (name) = (:name)
					WHERE section_id = :section_id
                    RETURNING instructor.id`

	insertQuery := `INSERT INTO meeting (section_id, name)
                    VALUES  (:section_id, :name)
                    RETURNING instructor.id`


	instructor_id = upsert(db, insertQuery, updateQuery, instructor)

	return instructor_id
}

func insertBook(db *sqlx.DB, book uct.Book) (book_id int64) {
	updateQuery := `UPDATE book SET (title, url) = (:title, :url)
					WHERE section_id = :section_id
                    RETURNING book.id`

	insertQuery := `INSERT INTO meeting (section_id, title, url)
                    VALUES  (:section_id, :title, :url)
                    RETURNING book.id`


	book_id = upsert(db, insertQuery, updateQuery, book)

	return book_id
}

func insertRegistration(db *sqlx.DB, registration uct.Registration) int64 {
	var registration_id int64

	updateQuery := `UPDATE registration SET (period_date) = (:period_date) 
					WHERE university_id = :university_id
	                RETURNING registration.id`
	
	insertQuery := `INSERT INTO registration (university_id, period, period_date)
                    VALUES (:university_id, :period, :period_date) 
                    RETURNING registration.id`

	registration_id = upsert(db, insertQuery, updateQuery, registration)

	return registration_id
}

func insertMetadata(db *sqlx.DB, metadata uct.Metadata) (metadata_id int64) {
	var insertQuery string
	var updateQuery string

	if metadata.UniversityId != 0 {
		updateQuery = `UPDATE metadata SET (title, content) = (:title, :content) 
					   WHERE metadata.university_id = :university_id 
		               RETURNING metadata.id`
		
		insertQuery = `INSERT INTO metadata (university_id, title, content)
                       VALUES (:university_id, :title, :content) 
                       RETURNING metadata.id`

	} else if metadata.SubjectId != 0 {
		updateQuery = `UPDATE metadata SET (title, content) = (:title, :content) 
					   WHERE metadata.subject_id = :subject_id 
		               RETURNING metadata.id`
		
		insertQuery = `INSERT INTO metadata (subject_id, title, content)
                       VALUES (:subject_id, :title, :content) 
                       RETURNING metadata.id`

	} else if metadata.CourseId != 0 {

		updateQuery = `UPDATE metadata SET (title, content) = (:title, :content) 
					   WHERE metadata.course_id = :course_id 
		               RETURNING metadata.id`
		
		insertQuery = `INSERT INTO metadata (course_id, title, content)
                       VALUES (:course_id, :title, :content) 
                       RETURNING metadata.id`

	} else if metadata.SectionId != 0 {
		updateQuery = `UPDATE metadata SET (title, content) = (:title, :content) 
					   WHERE metadata.section_id = :section_id 
		               RETURNING metadata.id`
		
		insertQuery = `INSERT INTO metadata (section_id, title, content)
                       VALUES (:section_id, :title, :content) 
                       RETURNING metadata.id`
		
	} else if metadata.MeetingId != 0 {
		updateQuery = `UPDATE metadata SET (title, content) = (:title, :content) 
					   WHERE metadata.meeting_id = :meeting_id 
		               RETURNING metadata.id`
		
		insertQuery = `INSERT INTO metadata (meeting_id, title, content)
                       VALUES (:meeting_id, :title, :content) 
                       RETURNING metadata.id`
	}

	metadata_id = upsert(db, insertQuery, updateQuery, metadata)
	
	return metadata_id
}

func insert(db *sqlx.DB, query string, data interface{}) int64 {
	var id int64

	if rows, err := db.NamedQuery(query, data); err != nil {
		log.Panicln(err)
	} else {
		for rows.Next() {
			if err = rows.Scan(&id); err != nil {
				log.Panicln(err)
			}
		}
	}
	return id
}

func update(db *sqlx.DB, query string, data interface{}) int64 {
	typeName := reflect.TypeOf(data).Name()
	b, _ := json.Marshal(data)
	dataString := string(b)

	var id int64
	if id == 0 {
		if rows, err := db.NamedQuery(query, data); err != nil {
			log.Panicln(err,typeName, dataString)
		} else {
			for rows.Next() {
				if err = rows.Scan(&id); err != nil {
					log.Panicln(err, typeName, dataString)
				}
			}
		}
	}
	return id
}
func upsert(db *sqlx.DB, insertQuery, updateQuery string, data interface{}) (id int64) {
	typeName := reflect.TypeOf(data).Name()
	if id = update(db, updateQuery, data); id != 0 {
		uct.Log("Update:",typeName, id)
	} else {
		id = insert(db, insertQuery, data)
		uct.Log("Insert:",typeName, id)
	}
	return
}