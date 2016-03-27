package main

import (
	"errors"
	"github.com/alecthomas/template/parse"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"strconv"
	"strings"
	"uct/common"
)

var (
	database *sqlx.DB
)

func init() {
	database = initDB(common.GetUniversityDB(false))
}

func initDB(connection string) *sqlx.DB {
	database, err := sqlx.Open("postgres", connection)
	if err != nil {
		log.Fatalln(err)
	}
	return database
}

func main() {
	r := gin.Default()
	r.GET("/university", universityHandler)
	r.GET("/subject", subjectHandler)
	r.GET("/course", courseHandler)
	r.GET("/section", sectionHandler)
	r.Run(":80")
}

func universityHandler(c *gin.Context) {
	dirtyDeep := c.DefaultQuery("deep", "true")
	dirtyId := c.DefaultQuery("id", "0")

	var deep bool
	var id int64
	var err error

	if deep, err = strconv.ParseBool(dirtyDeep); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: deep=%s", dirtyDeep)
	}

	if id, err = strconv.ParseInt(dirtyId, 10, 64); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: id=%s", dirtyId)
	}

	if s := c.Request.Header.Get("Accept"); s == "" || s != "application/json" {
		c.Error(errors.New("Missing header, Accept: application/json"))
		c.String(http.StatusBadRequest, "Missing header, Accept: application/json")
	}

	if id != 0 {
		return c.JSON(http.StatusOK, SelectUniversity(id, deep))
	} else {
		return c.JSON(http.StatusOK, SelectUniversities(deep))
	}

}

func subjectHandler(c *gin.Context) {
	dirtyDeep := c.DefaultQuery("deep", "true")
	dirtyUniversityId := c.Query("university_id")
	dirtyId := c.DefaultQuery("id", "0")
	dirtySeason := c.Query("season")
	dirtyYear := c.Query("year")

	var season common.Season
	var year string
	var deep bool
	var id int64
	var universityId int64
	var err error

	if deep, err = strconv.ParseBool(dirtyDeep); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: deep=%s", dirtyDeep)
	}

	if universityId, err = strconv.ParseInt(dirtyUniversityId, 10, 64); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: university_id=%s", dirtyUniversityId)
	}

	if id, err = strconv.ParseInt(dirtyId, 10, 64); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: id=%s", dirtyId)
	}

	if season, err = ParseSeason(dirtySeason); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: season=%s", dirtyId)
	}

	if year, err := strconv.ParseInt(dirtyYear, 10, 64); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: year=%s", dirtySeason)
	} else {
		year = year
	}

	if s := c.Request.Header.Get("Accept"); s == "" || s != "application/json" {
		c.Error(errors.New("Missing header, Accept: application/json"))
		c.String(http.StatusBadRequest, "Missing header, Accept: application/json")
	}

	if id != 0 {
		c.JSON(http.StatusOK, SelectSubject(id, deep))
	} else {
		c.JSON(http.StatusOK, SelectSubjects(universityId, season, year, deep))
	}
}

func courseHandler(c *gin.Context) {
	dirtyDeep := c.DefaultQuery("deep", "true")
	dirtyId := c.DefaultQuery("id", "0")
	dirstySubjectId := c.Query("subject_id")

	var deep bool
	var id int64
	var err error
	var subjectId int64

	if deep, err = strconv.ParseBool(dirtyDeep); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: deep=%s", dirtyDeep)
	}

	if id, err = strconv.ParseInt(dirtyId, 10, 64); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: id=%s", dirtyId)
	}

	if subjectId, err = strconv.ParseInt(dirstySubjectId, 10, 64); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: subject_id=%s", dirstySubjectId)
	}

	if s := c.Request.Header.Get("Accept"); s == "" || s != "application/json" {
		c.Error(errors.New("Missing header, Accept: application/json"))
		c.String(http.StatusBadRequest, "Missing header, Accept: application/json")
	}

	if id != 0 {
		return c.JSON(http.StatusOK, SelectCourse(id, deep))
	} else {
		return c.JSON(http.StatusOK, SelectCourses(subjectId, deep))
	}
}

func sectionHandler(c *gin.Context) {

}

func SelectUniversity(university_id int64, deep bool) (university common.University) {
	query := `SELECT * FROM university WHERE id = $1 ORDER BY name`
	PrepareAndGet(query, university, university_id)
	if deep && university != nil {
		deepSelectUniversities(university)
	}
	return
}

func SelectUniversities(deep bool) (universities []common.University) {
	query := `SELECT * FROM university ORDER BY name`
	PrepareAndSelect(query, universities)
	if deep {
		for i, _ := range universities {
			deepSelectUniversities(universities[i])
		}
	}
	return
}

func deepSelectUniversities(university common.University) {
	university.Registrations = SelectRegistrations(university.Id)
	university.Metadata = SelectMetadata(university.Id, 0, 0, 0, 0)
}

func SelectSubject(subject_id int64, deep bool) (subject common.Subject) {
	query := `SELECT * FROM subject WHERE id = $1 ORDER BY number`
	PrepareAndSelect(query, subject, subject_id)
	if deep && subject != nil {
		deepSelectSubject(subject)
	}
	return
}

func SelectSubjects(university_id int64, season, year string, deep bool) (subjects []*common.Subject) {
	query := `SELECT * FROM subject WHERE university_id = $1 AND season = $2 AND year = $3 ORDER BY number`
	PrepareAndSelect(query, subjects, university_id, season, year)
	if deep {
		for i := range subjects {
			deepSelectSubject(subjects[i])
		}
	}
	return
}

func deepSelectSubject(subject *common.Subject) {
	subject.Courses = SelectCourses(subject.Id, true)
	subject.Metadata = SelectMetadata(0, subject.Id, 0, 0, 0)
}

func SelectCourse(course_id int64, deep bool) (course common.Course) {
	query := `SELECT * FROM course WHERE id = $1 ORDER BY name`
	PrepareAndGet(query, course, course_id)
	if deep && course != nil {
		deepSelectCourse(course)
	}
	return
}

func SelectCourses(subjectId int64, deep bool) (courses []*common.Course) {
	query := `SELECT * FROM course WHERE subject_id = $1 ORDER BY name`
	PrepareAndSelect(query, courses, subjectId)
	if deep {
		for i := range courses {
			deepSelectCourse(courses[i])
		}
	}
	return
}

func deepSelectCourse(course *common.Course) {
	course.Sections = SelectSections(course.Id, true)
	course.Metadata = SelectMetadata(0, 0, course.Id, 0, 0)
}

func SelectSection(section_id int64, deep bool) (section *common.Section) {
	query := `SELECT * FROM section WHERE id = $1 ORDER BY number`
	PrepareAndGet(query, section, section_id)
	if deep && section != nil {
		deepSelectSection(section)
	}

	return
}

func SelectSections(course_id int64, deep bool) (sections []*common.Section) {
	query := `SELECT * FROM section WHERE course_id = $1 ORDER BY number`
	PrepareAndSelect(query, sections, course_id)
	if deep {
		for i := range sections {
			deepSelectSection(sections[i])
		}
	}

	return
}

func deepSelectSection(section *common.Section) {
	section.Meetings = SelectMeetings(section.Id)
	section.Books = SelectBooks(section.Id)
	section.Instructors = SelectInstructors(section.Id)
	section.Metadata = SelectMetadata(0, 0, 0, section.Id, 0)
}

func SelectMeetings(sectionId int64) (meetings []*common.Meeting) {
	query := `SELECT * FROM meetings WHERE section_id = $1 ORDER BY index`
	PrepareAndSelect(query, meetings, sectionId)

	for i := range meetings {
		meetings[i].Metadata = SelectMetadata(0, 0, 0, 0, meetings[i].Id)
	}
	return
}

func SelectInstructors(sectionId int64) (instructors []*common.Instructor) {
	query := `SELECT * FROM instructor WHERE section_id = $1`
	PrepareAndSelect(query, instructors, sectionId)
	return
}

func SelectBooks(sectionId int64) (books []*common.Book) {
	query := `SELECT * FROM book WHERE section_id = $1`
	PrepareAndSelect(query, books, sectionId)
	return
}

func SelectRegistrations(universityId int64) (registrations []*common.Registration) {
	query := `SELECT * FROM registrations WHERE university_id = $1`
	PrepareAndSelect(query, registrations, universityId)
	return
}

func SelectMetadata(universityId, subjectId, courseId, sectionId, meetingId int64) (metadata []*common.Metadata) {
	var err error
	var query string

	if universityId != 0 {
		query = `SELECT * FROM meeting WHERE university_id = $1`
		PrepareAndSelect(query, metadata, universityId)
	} else if subjectId != 0 {
		query = `SELECT * FROM meeting WHERE subject_id = $1`
		PrepareAndSelect(query, metadata, subjectId)
	} else if courseId != 0 {
		query = `SELECT * FROM meeting WHERE course_id = $1`
		PrepareAndSelect(query, metadata, courseId)
	} else if sectionId != 0 {
		query = `SELECT * FROM meeting WHERE section_id = $1`
		PrepareAndSelect(query, metadata, sectionId)

	} else if sectionId != 0 {
		query = `SELECT * FROM meeting WHERE meeting_id = $1`
		PrepareAndSelect(query, metadata, meetingId)
	} else {
		log.Panic("No valid metadata id was passed")
	}
	common.CheckError(err)
	return
}

func PrepareAndSelect(query string, data *interface{}, args ...interface{}) {
	if named, err := database.Preparex(query); err != nil {
		common.CheckError(err)
		err = named.Select(data, args)
		common.CheckError(err)
	}
}

func PrepareAndGet(query string, data *interface{}, args ...interface{}) {
	if named, err := database.Preparex(query); err != nil {
		common.CheckError(err)
		err = named.Get(data, args)
		common.CheckError(err)
	}
}

func ParseSeason(season string) (common.Season, error) {
	switch strings.ToLower(season) {
	case "w":
	case "winter":
		return common.WINTER
	case "s":
	case "spring":
		return common.SPRING
	case "u":
	case "summer":
		return common.SUMMER
	case "f":
	case "fall":
		return common.FALL
	}
	return errors.New("Could not parse season")
}
