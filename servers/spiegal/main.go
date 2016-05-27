package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pquerna/ffjson/ffjson"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	uct "uct/common"
)

var (
	database            *sqlx.DB
	preparedStmts       = make(map[string]*sqlx.NamedStmt)
	jsonContentType     = []string{"application/json; charset=utf-8"}
	protobufContentType = "application/x-protobuf; charset=utf-8"
)

var (
	app = kingpin.New("spiegal", "A command-line application to serve university course information")
	//format         = app.Flag("format", "choose input format").Short('f').HintOptions("protobuf", "json").PlaceHolder("[protobuf, json]").Required().String()
	port   = app.Flag("port", "port to start server on").Short('o').Default("9876").Uint16()
	server = app.Flag("pprof", "host:port to start profiling on").Short('p').Default(uct.SPIEGAL_DEBUG_SERVER).TCP()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Start profiling
	go uct.StartPprof(*server)

	database = uct.InitDB(uct.GetUniversityDB())

	r := gin.Default()
	v1 := r.Group("/v1")
	v1.Use(jsonWriter())
	v2 := r.Group("/v2")
	v2.Use(protobufWriter())

	v1.GET("/university", universityHandler)
	v1.GET("/subject", subjectHandler)
	v1.GET("/course", courseHandler)
	v1.GET("/section", sectionHandler)

	v2.GET("/university", universityHandler)
	v2.GET("/subject", subjectHandler)
	v2.GET("/course", courseHandler)
	v2.GET("/section", sectionHandler)

	r.Run(":" + strconv.Itoa(int(*port)))
}

/*
Add to middleware

if s := c.Request.Header.Get("Accept"); s == "" || s != "application/json" {
c.Error(errors.New("Missing header, Accept: application/json"))
c.String(http.StatusBadRequest, "Missing header, Accept: application/json")
*/

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
		u := SelectUniversityQuery(id, deep)
		b, err := proto.Marshal(&u)
		uct.CheckError(err)
		c.Set("protobuf", b)
		c.Set("object", u)
	} else {
		u := uct.Universities{Universities: SelectUniversities(deep)}
		b, err := proto.Marshal(&u)
		uct.CheckError(err)
		c.Set("protobuf", b)
		c.Set("object", u)
	}
}

func subjectHandler(c *gin.Context) {
	dirtyDeep := c.DefaultQuery("deep", "false")
	dirtyUniversityId := c.Query("university_id")
	dirtyId := c.DefaultQuery("", "0")
	dirtySeason := c.Query("season")
	dirtyYear := c.Query("year")

	var season uct.Season
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

	if id, err = strconv.ParseInt(dirtyId, 10, 64); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: id=%s", dirtyId)
		return
	}

	if universityId, err = strconv.ParseInt(dirtyUniversityId, 10, 64); err != nil && id == 0 {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: university_id=%s", dirtyUniversityId)
		return
	}

	if season, err = ParseSeason(dirtySeason); err != nil && id == 0 {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: season=%s", dirtySeason)
		return
	}

	if _, err := strconv.ParseInt(dirtyYear, 10, 64); err != nil && id == 0 {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: year=%s", dirtyYear)
		return
	} else {
		year = dirtyYear
	}

	if id != 0 {
		sub, b := SelectSubject(id, deep)
		c.Set("protobuf", b)
		c.Set("object", sub)
	} else {
		s := uct.Subjects{Subjects: SelectSubjects(universityId, season, year, deep)}
		b, err := proto.Marshal(&s)
		uct.CheckError(err)
		c.Set("protobuf", b)
		c.Set("object", s)
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
		SelectCourse(id, deep)
	} else {
		SelectCourses(subjectId, deep)
	}
}

func sectionHandler(c *gin.Context) {
	dirtyDeep := c.DefaultQuery("deep", "true")
	dirtyTopicName := c.DefaultQuery("topic", "")
	dirtyId := c.DefaultQuery("id", "0")
	dirtyCourse := c.DefaultQuery("course_id", "0")

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
	} else if dirtyTopicName != "" {
		c.JSON(http.StatusOK, SelectSectionByTopic(dirtyTopicName, deep))
	} else {
		c.JSON(http.StatusOK, SelectSections(courseId, deep))
	}
}

func protobufWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		value, _ := c.Get("protobuf")
		b, ok := value.([]byte)
		if !ok {
			log.Fatal("Can't retrieve response data")
		}

		c.Header("Content-Length", strconv.Itoa(len(b)))
		c.Header("Content-Type", protobufContentType)
		c.Writer.Write(b)
	}
}

func jsonWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		value, _ := c.Get("object")
		err := ffjson.NewEncoder(c.Writer).Encode(value)
		uct.CheckError(err)
	}
}

func writeJson(w gin.ResponseWriter, b []byte) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = jsonContentType
	}
	w.Write(b)
}

var (
	SelectUniversityQuery      = `SELECT id, name, abbr, home_page, registration_page, main_color, accent_color, topic_name FROM university WHERE name = :university_name ORDER BY name`
	ListUniversitiesQuery      = `SELECT id, name, abbr, home_page, registration_page, main_color, accent_color, topic_name FROM university ORDER BY name`
	GetAvailableSemestersQuery = `SELECT season, year FROM subject WHERE university.name = :university_name GROUP BY season, year`
)

func SelectUniversity(universityName string, deep bool) (university uct.University) {
	uct.Unique{UniversityName: universityName}
	if err := Get(SelectUniversityQuery, &university, uct.Unique{UniversityName: universityName}); err != nil {
		uct.CheckError(err)
	}
	s := GetAvailableSemesters(universityName)
	university.AvailableSemesters = s
	university.Metadata = SelectMetadata(university.Id, 0, 0, 0, 0)

	return
}

func SelectUniversities(deep bool) (universities []uct.University) {
	if err := Select(ListUniversitiesQuery, &universities, uct.Unique{}); err != nil {
		uct.CheckError(err)
	}

	for i, _ := range universities {
		s := GetAvailableSemesters(universities[i].Name)
		universities[i].AvailableSemesters = s
		universities[i].Metadata = SelectMetadata(universities[i].Id, 0, 0, 0, 0)
	}
	return
}

func GetAvailableSemesters(universityName string) (semesters []uct.Semester) {
	defer uct.TimeTrack(time.Now(), "GetAvailableSemesters: ")
	s := []uct.Semester{}
	if err := Select(GetAvailableSemestersQuery, &s, uct.Unique{UniversityName: universityName}); err != nil {
		uct.CheckError(err)
	}
	return
}

/*func SelectSubject(subject_id int64, deep bool) (subject uct.Subject) {
	defer uct.TimeTrack(time.Now(), "SelectSubject deep:"+fmt.Sprint(deep))
	key := "subject"
	query := `SELECT id, name, number, season, year, hash, topic_name FROM subject WHERE id = $1 ORDER BY number`
	if err := Get(GetCachedStmt(key, query), &subject, subject_id); err != nil {
		uct.CheckError(err)
	}
	if deep && &subject != nil {
		deepSelectSubject(&subject)
	}
	return
}*/

type Data struct {
	Data []byte `db:"data"`
}

var (
	SelectProtoSubjectQuery = `SELECT data FROM subject JOIN university ON university.id = subject.university_id
									AND university.name = :university_name
									AND season = :subject_season
									AND year =:subject_year
									AND subject.name = :subject_name
									AND subject.number = :subject_number`

	ListSubjectQuery = `SELECT subject.id, subject.name, subject.number, subject.season, subject.year, subject.topic_name
								FROM subject JOIN university ON university.id = subject.university_id
									AND university.name = :university_name
									AND season = :subject_season
									AND year =:subject_year`
)

func SelectSubject(universityName, season, year, name, number string) (subject uct.Subject, b []byte) {
	defer uct.TimeTrack(time.Now(), "SelectProtoSubject")
	m := map[string]interface{}{"university_name": universityName, "subject_season": season, "subject_year": year, "subject_name": name,
		"subject_number": number,
	}
	d := Data{}
	if err := Get(SelectProtoSubjectQuery, &d, m); err != nil {
		uct.CheckError(err)
	}
	b = d.Data
	err := proto.Unmarshal(d.Data, &subject)
	uct.CheckError(err)
	return
}

func SelectSubjects(universityName, season, year string) (subjects []uct.Subject) {
	defer uct.TimeTrack(time.Now(), "SelectSubjects ")
	m := map[string]interface{}{"university_name": universityName, "subject_season": season, "subject_year": year}
	if err := Select(ListSubjectQuery, &subjects, m); err != nil {
		uct.CheckError(err)
	}
	return
}

var (
	SelectCourseQuery = `SELECT data FROM course JOIN subject ON subejct.id = course.subject_id
							JOIN university ON university.id = subject.university_id
							WHERE university.name = :university_name
							AND subject.season = :subject_season
							AND subject.year = :subject_year
							AND subject.name = :subject_name
							AND subject.number = :subject_number
							AND course.name = :course_name
							AND course.number = :course_number`
)

func SelectCourse(universityName, season, year, subjectName, subjectNumber, courseName, courseNumber string) (course uct.Course, b []byte) {
	defer uct.TimeTrack(time.Now(), "SelectCourse")
	d := Data{}
	m := map[string]interface{}{"university_name": universityName, "subject_season": season, "subject_year": year,
		"subject_name": subjectName, "subject_number": subjectNumber, "course_name":courseName, "course_number":courseNumber}
	if err := Get(SelectCourseQuery, &d, m); err != nil {
		uct.CheckError(err)
	}
	b = d.Data
	err := proto.Unmarshal(b, &course)
	uct.CheckError(err)
	return
}

func SelectCourses(subjectId int64, deep bool) (courses []uct.Course) {
	defer uct.TimeTrack(time.Now(), "SelectCourses deep:"+fmt.Sprint(deep))
	key := "courses"
	query := `SELECT * FROM course WHERE subject_id = $1 ORDER BY number`
	if err := Select(GetCachedStmt(key, query), &courses, subjectId); err != nil {
		uct.CheckError(err)
	}
	if deep {
		for i := range courses {
			deepSelectCourse(&courses[i])
		}
	}
	return
}

func deepSelectCourse(course *uct.Course) {
	course.Sections = SelectSections(course.Id, true)
	course.Metadata = SelectMetadata(0, 0, course.Id, 0, 0)
}

func SelectSection(section_id int64, deep bool) (section uct.Section) {
	defer uct.TimeTrack(time.Now(), "SelectSection deep:"+fmt.Sprint(deep))

	key := "section"
	query := `SELECT * FROM section WHERE id = $1 ORDER BY number`
	if err := Get(GetCachedStmt(key, query), &section, section_id); err != nil {
		uct.CheckError(err)
	}
	if deep && &section != nil {
		deepSelectSection(&section)
	}

	return
}

func SelectSectionByTopic(topicName string, deep bool) (section uct.Section) {
	defer uct.TimeTrack(time.Now(), "SelectSection deep:"+fmt.Sprint(deep))

	key := "section"
	query := `SELECT * FROM section WHERE topic_name = $1 ORDER BY number`
	if err := Get(GetCachedStmt(key, query), &section, topicName); err != nil {
		uct.CheckError(err)
	}
	if deep && &section != nil {
		deepSelectSection(&section)
	}

	return
}

func SelectSections(course_id int64, deep bool) (sections []uct.Section) {
	defer uct.TimeTrack(time.Now(), "SelectSections deep:"+fmt.Sprint(deep))

	key := "sections"
	query := `SELECT * FROM section WHERE course_id = $1 ORDER BY number`
	if err := Select(GetCachedStmt(key, query), &sections, course_id); err != nil {
		uct.CheckError(err)
	}
	if deep {
		for i := range sections {
			deepSelectSection(&sections[i])
		}
	}

	return
}

func deepSelectSection(section *uct.Section) {
	section.Meetings = SelectMeetings(section.Id)
	section.Books = SelectBooks(section.Id)
	section.Instructors = SelectInstructors(section.Id)
	section.Metadata = SelectMetadata(0, 0, 0, section.Id, 0)
}

func SelectMeetings(sectionId int64) (meetings []uct.Meeting) {
	defer uct.TimeTrack(time.Now(), "SelectMeetings")

	key := "meetings"
	query := `SELECT * FROM meeting WHERE section_id = $1 ORDER BY index`
	if err := Select(GetCachedStmt(key, query), &meetings, sectionId); err != nil {
		uct.CheckError(err)
	}
	for i := range meetings {
		meetings[i].Metadata = SelectMetadata(0, 0, 0, 0, meetings[i].Id)
	}
	return
}

func SelectInstructors(sectionId int64) (instructors []uct.Instructor) {
	defer uct.TimeTrack(time.Now(), "SelectInstructors")

	key := "instructors"
	query := `SELECT * FROM instructor WHERE section_id = $1`
	if err := Select(GetCachedStmt(key, query), &instructors, sectionId); err != nil {
		uct.CheckError(err)
	}
	return
}

func SelectBooks(sectionId int64) (books []uct.Book) {
	defer uct.TimeTrack(time.Now(), "SelectInstructors")

	key := "books"
	query := `SELECT * FROM book WHERE section_id = $1`
	if err := Select(GetCachedStmt(key, query), &books, sectionId); err != nil {
		uct.CheckError(err)
	}
	return
}

func SelectMetadata(universityId, subjectId, courseId, sectionId, meetingId int64) (metadata []uct.Metadata) {
	defer uct.TimeTrack(time.Now(), "SelectMetadata")

	var err error
	var query string

	if universityId != 0 {
		key := "university_metatdata"
		query = `SELECT title, content FROM metadata WHERE university_id = $1`
		err = Select(GetCachedStmt(key, query), &metadata, universityId)
	} else if subjectId != 0 {
		key := "subject_metatdata"
		query = `SELECT title, content FROM metadata WHERE subject_id = $1`
		err = Select(GetCachedStmt(key, query), &metadata, subjectId)
	} else if courseId != 0 {
		key := "course_metatdata"
		query = `SELECT title, content FROM metadata WHERE course_id = $1`
		err = Select(GetCachedStmt(key, query), &metadata, courseId)
	} else if sectionId != 0 {
		key := "section_metatdata"
		query = `SELECT title, content FROM metadata WHERE section_id = $1`
		err = Select(GetCachedStmt(key, query), &metadata, sectionId)
	} else if meetingId != 0 {
		key := "meeting_metatdata"
		query = `SELECT title, content FROM metadata WHERE meeting_id = $1`
		err = Select(GetCachedStmt(key, query), &metadata, meetingId)
	} else {
		log.Panic("No valid metadata id was passed")
	}
	uct.CheckError(err)
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
		panic(fmt.Errorf("Error %s Query: %s", query, err))
		return nil
	} else {
		return named
	}
}

func ParseSeason(s string) (season string, err error) {
	switch strings.ToLower(s) {
	case "winter", "w":
		return uct.WINTER, err
	case "spring", "s":
		return uct.SPRING, err
	case "summer", "u":
		return uct.SUMMER, err
	case "fall", "f":
		return uct.FALL, err
	}
	err = errors.New("Could not parse season")
	return season, err
}
