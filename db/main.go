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
	upsertUni := `INSERT INTO university (name, abbr, home_page, registration_page, main_color, accent_color, topic_name)
					VALUES (:name, :abbr, :home_page, :registration_page, :main_color, :accent_color, :topic_name)
					ON CONFLICT
						ON CONSTRAINT unique_university_name
						DO UPDATE SET (abbr, home_page, registration_page, main_color, accent_color, topic_name) = (EXCLUDED.abbr, EXCLUDED.home_page, EXCLUDED.registration_page, EXCLUDED.main_color, EXCLUDED.accent_color, EXCLUDED.topic_name)
					RETURNING university.id`

	if rows, err := db.NamedQuery(upsertUni, uni); err != nil {
		log.Panicln(err)
	} else {
		for rows.Next() {
			if err = rows.Scan(&university_id); err != nil {
				log.Panicln(err)
			}
		}
		uct.Log("UPSERT university_id ", university_id)
	}

	for _, subject := range uni.Subjects {
		subject.UniversityId = university_id
		subId := insertSubject(db, subject)

		for _, course := range subject.Courses {
			course.SubjectId = subId
			insertCourse(db, course)
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

func insertSubject(db *sqlx.DB, sub uct.Subject) int64 {
	sub.VetAndBuild()
	var subject_id int64

	upsertSub := `INSERT INTO subject (university_id, name, number, season, year, topic_name)
					VALUES  (:university_id, :name, :number, :season, :year, :topic_name)
					ON CONFLICT
						ON CONSTRAINT unique_subject_name_number__university_id_season_year
						DO UPDATE SET (name, season, year, topic_name) = (EXCLUDED.name, EXCLUDED.season, EXCLUDED.year, EXCLUDED.topic_name)
					RETURNING subject.id`
	if rows, err := db.NamedQuery(upsertSub, sub); err != nil {
		log.Panicln(err)
	} else {
		for rows.Next() {
			if err = rows.Scan(&subject_id); err != nil {
				log.Panicln(err)
			}
		}
		uct.Log("UPSERT subject_id ", subject_id)
	}
	return subject_id
}

func insertRegistration(db *sqlx.DB, registration uct.Registration) int64 {
	var registration_id int64

	upsertRegis := `INSERT INTO registration (university_id, period, period_date)
					VALUES (:university_id, :period, :period_date)
					ON CONFLICT
						ON CONSTRAINT unique_registration_period__university_id
						DO UPDATE SET (period_date) = (:period_date)
					RETURNING registration.id`

	if rows, err := db.NamedQuery(upsertRegis, registration); err != nil {
		log.Panicln(err)
	} else {
		for rows.Next() {
			if err = rows.Scan(&registration_id); err != nil {
				log.Panicln(err)
			}
		}
		uct.Log("UPSERT registration_id ", registration_id)
	}
	return registration_id
}

func insertMetadata(db *sqlx.DB, metadata uct.Metadata) int64 {
	var metadata_id int64
	var upsertMetaData string

	if metadata.UniversityId != 0 {
		upsertMetaData = `INSERT INTO metadata (university_id, title, content)
					VALUES (:university_id, :title, :content)
					ON CONFLICT
						ON CONSTRAINT unique_metadata_title__university_id
						DO UPDATE SET (title, content) = (:title, :content)
					RETURNING id`
	} else if metadata.SubjectId != 0 {
		upsertMetaData = `INSERT INTO metadata (subject_id, title, content)
					VALUES (:subject_id, :title, :content)
					ON CONFLICT
						ON CONSTRAINT unique_metadata_title__subject_id
						DO UPDATE SET (metadata.title, metadata.content) = (:title, :content)
					RETURNING id`
	} else if metadata.CourseId != 0 {
		upsertMetaData = `INSERT INTO metadata (course_id, title, content)
					VALUES (:course_id, :title, :content)
					ON CONFLICT
						ON CONSTRAINT unique_metadata_title__course_id
						DO UPDATE SET (metadata.title, metadata.content) = (:title, :content)
					RETURNING id`
	} else if metadata.SectionId != 0 {
		upsertMetaData = `INSERT INTO metadata (section_id, title, content)
					VALUES (:section_id, :title, :content)
					ON CONFLICT
						ON CONSTRAINT unique_metadata_title__section_id
						DO UPDATE SET (metadata.title, metadata.content) = (:title, :content)
					RETURNING id`
	} else if metadata.MeetingId != 0 {
		upsertMetaData = `INSERT INTO metadata (meeting_id, title, content)
					VALUES (:meeting_id, :title, :content)
					ON CONFLICT
						ON CONSTRAINT unique_metadata_title__meeting_id
						DO UPDATE SET (metadata.title, metadata.content) = (:title, :content)
					RETURNING id`
	}


	if rows, err := db.NamedQuery(upsertMetaData, metadata); err != nil {
		log.Panicln(err)
	} else {
		for rows.Next() {
			if err = rows.Scan(&metadata_id); err != nil {
				log.Panicln(err)
			}
		}
		uct.Log("UPSERT metadata_id ", metadata_id)
	}
	return metadata_id
}

func insertCourse(db *sqlx.DB, course uct.Course) int64 {
	course.VetAndBuild()
	var course_id int64

	upsertSub := `INSERT INTO course (subject_id, name, number, synopsis, topic_name)
					VALUES  (:subject_id, :name, :number, :synopsis, :topic_name)
					ON CONFLICT
						ON CONSTRAINT unique_course_name__number__subject_id
						DO UPDATE SET (name, synopsis, topic_name) = (EXCLUDED.name, EXCLUDED.synopsis, EXCLUDED.topic_name)
					RETURNING course.id`
	if rows, err := db.NamedQuery(upsertSub, course); err != nil {
		/*
		b , _ := json.Marshal(course)
		*/
		log.Printf("Name: %x\n", []byte(course.Name))
		log.Printf("Name: %s\n", course.Name)
		log.Panicln(err)
	} else {
		for rows.Next() {
			if err = rows.Scan(&course_id); err != nil {
				log.Panicln(err)
			}
		}
		uct.Log("UPSERT course_id ", course_id)
	}
	return course_id
}