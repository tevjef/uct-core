package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"uct/common"
)

var (
	database      *sqlx.DB
	preparedStmts = make(map[string]*sqlx.Stmt)
)

func init() {
	database = initDB(common.GetUniversityDB())
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
	r.Run(":9876")
}

/*
Add to middleware

if s := c.Request.Header.Get("Accept"); s == "" || s != "application/json" {
c.Error(errors.New("Missing header, Accept: application/json"))
c.String(http.StatusBadRequest, "Missing header, Accept: application/json")
}*/
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

	if id != 0 {
		c.JSON(http.StatusOK, SelectUniversity(id, deep))
	} else {
		c.JSON(http.StatusOK, SelectUniversities(deep))
	}

}

func subjectHandler(c *gin.Context) {
	dirtyDeep := c.DefaultQuery("deep", "false")
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
		return
	}

	if universityId, err = strconv.ParseInt(dirtyUniversityId, 10, 64); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: university_id=%s", dirtyUniversityId)
		return
	}

	if id, err = strconv.ParseInt(dirtyId, 10, 64); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: id=%s", dirtyId)
		return
	}

	if season, err = ParseSeason(dirtySeason); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: season=%s", dirtySeason)
		return
	}

	if _, err := strconv.ParseInt(dirtyYear, 10, 64); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: year=%s", dirtyYear)
		return
	} else {
		year = dirtyYear
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
	dirtySubjectId := c.Query("subject_id")

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

	if subjectId, err = strconv.ParseInt(dirtySubjectId, 10, 64); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: subject_id=%s", dirtySubjectId)
	}

	if id != 0 {
		c.JSON(http.StatusOK, SelectCourse(id, deep))
	} else {
		c.JSON(http.StatusOK, SelectCourses(subjectId, deep))
	}
}

func sectionHandler(c *gin.Context) {
	dirtyDeep := c.DefaultQuery("deep", "true")
	dirtyId := c.DefaultQuery("id", "0")
	dirtyCourse := c.Query("course_id")

	var deep bool
	var id int64
	var err error
	var courseId int64

	if deep, err = strconv.ParseBool(dirtyDeep); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: deep=%s", dirtyDeep)
	}

	if id, err = strconv.ParseInt(dirtyId, 10, 64); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: id=%s", dirtyId)
	}

	if courseId, err = strconv.ParseInt(dirtyCourse, 10, 64); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: course_id=%s", dirtyCourse)
	}

	if id != 0 {
		c.JSON(http.StatusOK, SelectSection(id, deep))
	} else {
		c.JSON(http.StatusOK, SelectSections(courseId, deep))
	}
}

func SelectUniversity(university_id int64, deep bool) (university common.University) {
	key := "university"
	query := `SELECT * FROM university WHERE id = $1 ORDER BY name`
	if err := Get(GetCachedStmt(key, query), &university, university_id); err != nil {
		common.CheckError(err)
	}
	if deep && &university != nil {
		deepSelectUniversities(&university)
	}
	return
}

func SelectUniversities(deep bool) (universities []common.University) {
	key := "universities"
	query := `SELECT * FROM university ORDER BY name`
	if err := Select(GetCachedStmt(key, query), &universities); err != nil {
		common.CheckError(err)
	}
	if deep {
		for i, _ := range universities {
			deepSelectUniversities(&universities[i])
		}
	}
	return
}

func deepSelectUniversities(university *common.University) {
	university.Registrations = SelectRegistrations(university.Id)
	university.Metadata = SelectMetadata(university.Id, 0, 0, 0, 0)
}

func SelectSubject(subject_id int64, deep bool) (subject common.Subject) {
	defer common.TimeTrack(time.Now(), "SelectSubject deep:"+fmt.Sprint(deep))
	key := "subject"
	query := `SELECT * FROM subject WHERE id = $1 ORDER BY number`
	if err := Get(GetCachedStmt(key, query), &subject, subject_id); err != nil {
		common.CheckError(err)
	}
	if deep && &subject != nil {
		deepSelectSubject(&subject)
	}
	return
}

func SelectSubjects(university_id int64, season common.Season, year string, deep bool) (subjects []common.Subject) {
	defer common.TimeTrack(time.Now(), "SelectSubjects deep:"+fmt.Sprint(deep))
	key := "subjects"
	query := `SELECT * FROM subject WHERE university_id = $1 AND season = $2 AND year = $3 ORDER BY number`
	if err := Select(GetCachedStmt(key, query), &subjects, university_id, season.String(), year); err != nil {
		common.CheckError(err)
	}
	if deep {
		for i := range subjects {
			deepSelectSubject(&subjects[i])
		}
	}
	return
}

func deepSelectSubject(subject *common.Subject) {
	subject.Courses = SelectCourses(subject.Id, true)
	subject.Metadata = SelectMetadata(0, subject.Id, 0, 0, 0)
}

func SelectCourse(course_id int64, deep bool) (course common.Course) {
	defer common.TimeTrack(time.Now(), "SelectCourse deep:"+fmt.Sprint(deep))
	key := "course"
	query := `SELECT * FROM course WHERE id = $1 ORDER BY name`
	if err := Get(GetCachedStmt(key, query), &course, course_id); err != nil {
		common.CheckError(err)
	}
	if deep && &course != nil {
		deepSelectCourse(&course)
	}
	return
}

func SelectCourses(subjectId int64, deep bool) (courses []common.Course) {
	defer common.TimeTrack(time.Now(), "SelectCourses deep:"+fmt.Sprint(deep))
	key := "courses"
	query := `SELECT * FROM course WHERE subject_id = $1 ORDER BY name`
	if err := Select(GetCachedStmt(key, query), &courses, subjectId); err != nil {
		common.CheckError(err)
	}
	if deep {
		for i := range courses {
			deepSelectCourse(&courses[i])
		}
	}
	return
}

func deepSelectCourse(course *common.Course) {
	course.Sections = SelectSections(course.Id, true)
	course.Metadata = SelectMetadata(0, 0, course.Id, 0, 0)
}

func SelectSection(section_id int64, deep bool) (section common.Section) {
	defer common.TimeTrack(time.Now(), "SelectSection deep:"+fmt.Sprint(deep))

	key := "section"
	query := `SELECT * FROM section WHERE id = $1 ORDER BY number`
	if err := Get(GetCachedStmt(key, query), &section, section_id); err != nil {
		common.CheckError(err)
	}
	if deep && &section != nil {
		deepSelectSection(&section)
	}

	return
}

func SelectSections(course_id int64, deep bool) (sections []common.Section) {
	defer common.TimeTrack(time.Now(), "SelectSections deep:"+fmt.Sprint(deep))

	key := "sections"
	query := `SELECT * FROM section WHERE course_id = $1 ORDER BY number`
	if err := Select(GetCachedStmt(key, query), &sections, course_id); err != nil {
		common.CheckError(err)
	}
	if deep {
		for i := range sections {
			deepSelectSection(&sections[i])
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

func SelectMeetings(sectionId int64) (meetings []common.Meeting) {
	defer common.TimeTrack(time.Now(), "SelectMeetings")

	key := "meetings"
	query := `SELECT * FROM meeting WHERE section_id = $1 ORDER BY index`
	if err := Select(GetCachedStmt(key, query), &meetings, sectionId); err != nil {
		common.CheckError(err)
	}
	for i := range meetings {
		meetings[i].Metadata = SelectMetadata(0, 0, 0, 0, meetings[i].Id)
	}
	return
}

func SelectInstructors(sectionId int64) (instructors []common.Instructor) {
	defer common.TimeTrack(time.Now(), "SelectInstructors")

	key := "instructors"
	query := `SELECT * FROM instructor WHERE section_id = $1`
	if err := Select(GetCachedStmt(key, query), &instructors, sectionId); err != nil {
		common.CheckError(err)
	}
	return
}

func SelectBooks(sectionId int64) (books []common.Book) {
	defer common.TimeTrack(time.Now(), "SelectInstructors")

	key := "books"
	query := `SELECT * FROM book WHERE section_id = $1`
	if err := Select(GetCachedStmt(key, query), &books, sectionId); err != nil {
		common.CheckError(err)
	}
	return
}

func SelectRegistrations(universityId int64) (registrations []common.Registration) {
	key := "registrations"
	query := `SELECT * FROM registrations WHERE university_id = $1`
	if err := Select(GetCachedStmt(key, query), &registrations, universityId); err != nil {
		common.CheckError(err)
	}
	return
}

func SelectMetadata(universityId, subjectId, courseId, sectionId, meetingId int64) (metadata []common.Metadata) {
	defer common.TimeTrack(time.Now(), "SelectMetadata")

	var err error
	var query string

	if universityId != 0 {
		key := "university_metatdata"
		query = `SELECT * FROM metadata WHERE university_id = $1`
		err = Select(GetCachedStmt(key, query), &metadata, universityId)
	} else if subjectId != 0 {
		key := "subject_metatdata"
		query = `SELECT * FROM metadata WHERE subject_id = $1`
		err = Select(GetCachedStmt(key, query), &metadata, subjectId)
	} else if courseId != 0 {
		key := "course_metatdata"
		query = `SELECT * FROM metadata WHERE course_id = $1`
		err = Select(GetCachedStmt(key, query), &metadata, courseId)
	} else if sectionId != 0 {
		key := "section_metatdata"
		query = `SELECT * FROM metadata WHERE section_id = $1`
		err = Select(GetCachedStmt(key, query), &metadata, sectionId)
	} else if meetingId != 0 {
		key := "meeting_metatdata"
		query = `SELECT * FROM metadata WHERE meeting_id = $1`
		err = Select(GetCachedStmt(key, query), &metadata, meetingId)
	} else {
		log.Panic("No valid metadata id was passed")
	}
	common.CheckError(err)
	return
}

func Select(named *sqlx.Stmt, data interface{}, args ...interface{}) error {
	if err := named.Select(data, args...); err != nil {
		return err
	}
	return nil
}

func Get(named *sqlx.Stmt, data interface{}, args ...interface{}) error {
	if err := named.Get(data, args...); err != nil {
		return err
	}
	return nil
}

func GetCachedStmt(key, query string) *sqlx.Stmt {
	if stmt := preparedStmts[key]; stmt == nil {
		preparedStmts[key] = Prepare(query)
	}
	return preparedStmts[key]
}

func Prepare(query string) *sqlx.Stmt {
	if named, err := database.Preparex(query); err != nil {
		common.CheckError(err)
		return nil
	} else {
		return named
	}
}

func ParseSeason(s string) (season common.Season, err error) {
	switch strings.ToLower(s) {
	case "winter", "w":
		return common.WINTER, err
	case "spring", "s":
		return common.SPRING, err
	case "summer", "u":
		return common.SUMMER, err
	case "fall", "f":
		return common.FALL, err
	}
	err = errors.New("Could not parse season")
	return season, err
}
