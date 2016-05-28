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
	protobufContentType = "application/x-protobuf"
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

	PrepareAllStmts()

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

func universityHandler(c *gin.Context) {
	topicName := c.DefaultQuery("university_topic_name", "")

	if topicName != "" {
		u := SelectUniversity(topicName)
		b, err := proto.Marshal(&u)
		uct.CheckError(err)
		c.Set("protobuf", b)
		c.Set("object", u)
	} else {
		u := uct.Universities{Universities: SelectUniversities()}
		b, err := proto.Marshal(&u)
		uct.CheckError(err)
		c.Set("protobuf", b)
		c.Set("object", u)
	}
}

func subjectHandler(c *gin.Context) {
	season := c.Query("season")
	year := c.Query("year")
	uniTopicName := c.Query("university_topic_name")
	subjectTopicName := c.Query("subject_topic_name")

	if _, err := strconv.ParseInt(year, 10, 64); err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, "Could not parse parameter: year=%s", year)
		return
	}

	if season == "" {
		c.String(http.StatusBadRequest, "Must provide the season parameter")
		return
	}

	if uniTopicName == "" {
		c.String(http.StatusBadRequest, "Must provide a university_topic_name parameter")
		return
	}

	if subjectTopicName != "" {
		sub, b := SelectSubject(subjectTopicName)
		c.Set("protobuf", b)
		c.Set("object", sub)
	} else {
		s := uct.Subjects{Subjects: SelectSubjects(uniTopicName, season, year)}
		b, err := proto.Marshal(&s)
		uct.CheckError(err)
		c.Set("protobuf", b)
		c.Set("object", s)
	}
}

func courseHandler(c *gin.Context) {
	subjectTopicName := c.Query("subject_topic_name")
	courseTopicName := c.Query("course_topic_name")

	var err error

	if subjectTopicName == "" {
		c.Error(err)
		c.String(http.StatusBadRequest, "Must provide the subject_topic_name parameter")
		return
	}

	if courseTopicName != "" {
		course, b := SelectCourse(courseTopicName)
		c.Set("protobuf", b)
		c.Set("object", course)
	} else {
		courses := uct.Courses{Courses:SelectCourses(subjectTopicName)}
		b, err := proto.Marshal(&courses)
		uct.CheckError(err)
		c.Set("protobuf", b)
		c.Set("object", courses)
	}
}

func sectionHandler(c *gin.Context) {
	sectionTopicName := c.Query("section_topic_name")

	var err error

	if sectionTopicName == "" {
		c.Error(err)
		c.String(http.StatusBadRequest, "Must provide the section_topic_name parameter")
		return
	}

	s := SelectSection(sectionTopicName)
	b, err := proto.Marshal(&s)
	uct.CheckError(err)
	c.Set("protobuf", b)
	c.Set("object", s)
}

func protobufWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if value, exists := c.Get("protobuf"); exists {
			b, ok := value.([]byte)
			if !ok {
				log.Fatal("Can't retrieve response data")
			}

			c.Header("Content-Length", strconv.Itoa(len(b)))
			c.Header("Content-Type", protobufContentType)
			c.Writer.Write(b)
		}

	}
}

func jsonWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if value, exists := c.Get("object"); exists {
			writeJsonHeaders(c.Writer)
			err := ffjson.NewEncoder(c.Writer).Encode(value)
			uct.CheckError(err)
		}

	}
}

func writeJsonHeaders(w gin.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = jsonContentType
	}
}

var ()

func SelectUniversity(universityName string) (university uct.University) {
	if err := Get(SelectUniversityQuery, &university, uct.Unique{UniversityName: universityName}); err != nil {
		uct.CheckError(err)
	}
	s := GetAvailableSemesters(universityName)
	university.AvailableSemesters = s
	university.Metadata = SelectMetadata(university.Id, 0, 0, 0, 0)

	return
}

func SelectUniversities() (universities []*uct.University) {
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

func GetAvailableSemesters(universityName string) (semesters []*uct.Semester) {
	defer uct.TimeTrack(time.Now(), "GetAvailableSemesters: ")
	s := []uct.Semester{}
	if err := Select(GetAvailableSemestersQuery, &s, uct.Unique{UniversityName: universityName}); err != nil {
		uct.CheckError(err)
	}
	return
}

type Data struct {
	Data []byte `db:"data"`
}

func SelectSubject(subjectTopicName string) (subject uct.Subject, b []byte) {
	defer uct.TimeTrack(time.Now(), "SelectProtoSubject")
	m := map[string]interface{}{"topic_name": subjectTopicName}
	d := Data{}
	if err := Get(SelectProtoSubjectQuery, &d, m); err != nil {
		uct.CheckError(err)
	}
	b = d.Data
	err := proto.Unmarshal(d.Data, &subject)
	uct.CheckError(err)
	return
}

func SelectSubjects(uniTopicName, season, year string) (subjects []*uct.Subject) {
	defer uct.TimeTrack(time.Now(), "SelectSubjects")
	m := map[string]interface{}{"topic_name": uniTopicName, "subject_season": season, "subject_year": year}
	if err := Select(ListSubjectQuery, &subjects, m); err != nil {
		uct.CheckError(err)
	}
	return
}

func SelectCourse(courseTopicName string) (course uct.Course, b []byte) {
	defer uct.TimeTrack(time.Now(), "SelectCourse")
	d := Data{}
	m := map[string]interface{}{"topic_name": courseTopicName}
	if err := Get(SelectCourseQuery, &d, m); err != nil {
		uct.CheckError(err)
	}
	b = d.Data
	err := proto.Unmarshal(b, &course)
	uct.CheckError(err)
	return
}

func SelectCourses(subjectTopicName string) (courses []*uct.Course) {
	defer uct.TimeTrack(time.Now(), "SelectCourses deep")

	d := []Data{}
	m := map[string]interface{}{"topic_name": subjectTopicName}

	if err := Select(ListCoursesQuery, &d, m); err != nil {
		uct.CheckError(err)
	}

	for i := range d {
		c := uct.Course{}
		err := proto.Unmarshal(d[i].Data, &c)
		uct.CheckError(err)
		courses = append(courses, &c)
	}

	return
}

func SelectSection(sectionTopicName string) (section uct.Section) {
	defer uct.TimeTrack(time.Now(), "SelectSection")

	m := map[string]interface{}{"topic_name": sectionTopicName}

	if err := Get(SelectSectionQuery, &section, m); err != nil {
		uct.CheckError(err)
	}
	deepSelectSection(&section)
	return
}

func deepSelectSection(section *uct.Section) {
	section.Meetings = SelectMeetings(section.Id)
	section.Books = SelectBooks(section.Id)
	section.Instructors = SelectInstructors(section.Id)
	section.Metadata = SelectMetadata(0, 0, 0, section.Id, 0)
}

func SelectMeetings(sectionId int64) (meetings []*uct.Meeting) {
	defer uct.TimeTrack(time.Now(), "SelectMeetings")
	m := map[string]interface{}{"section_id": sectionId}
	if err := Select(SelectMeeting, &meetings, m); err != nil {
		uct.CheckError(err)
	}
	for i := range meetings {
		meetings[i].Metadata = SelectMetadata(0, 0, 0, 0, meetings[i].Id)
	}
	return
}

func SelectInstructors(sectionId int64) (instructors []*uct.Instructor) {
	defer uct.TimeTrack(time.Now(), "SelectInstructors")
	m := map[string]interface{}{"section_id": sectionId}
	if err := Select(SelectInstructor, &instructors, m); err != nil {
		uct.CheckError(err)
	}
	return
}

func SelectBooks(sectionId int64) (books []*uct.Book) {
	defer uct.TimeTrack(time.Now(), "SelectInstructors")
	m := map[string]interface{}{"section_id": sectionId}
	if err := Select(SelectBook, &books, m); err != nil {
		uct.CheckError(err)
	}
	return
}

func SelectMetadata(universityId, subjectId, courseId, sectionId, meetingId int64) (metadata []*uct.Metadata) {
	defer uct.TimeTrack(time.Now(), "SelectMetadata")
	var err error
	if universityId != 0 {
		err = Select(UniversityMetadataQuery, &metadata, universityId)
	} else if subjectId != 0 {
		err = Select(SubjectMetadataQuery, &metadata, subjectId)
	} else if courseId != 0 {
		err = Select(CourseMetadataQuery, &metadata, courseId)
	} else if sectionId != 0 {
		err = Select(SectionMetadataQuery, &metadata, sectionId)
	} else if meetingId != 0 {
		err = Select(MeetingMetadataQuery, &metadata, meetingId)
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
		panic(fmt.Errorf("Error: %s Query: %s", query, err))
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
	SelectUniversityQuery      = `SELECT id, name, abbr, home_page, registration_page, main_color, accent_color, topic_name FROM university WHERE topic_name = :topic_name ORDER BY name`
	ListUniversitiesQuery      = `SELECT id, name, abbr, home_page, registration_page, main_color, accent_color, topic_name FROM university ORDER BY name`
	GetAvailableSemestersQuery = `SELECT season, year FROM subject JOIN university ON university.id = subject.university_id
									WHERE university.name = :university_name GROUP BY season, year`

	SelectProtoSubjectQuery = `SELECT data FROM subject WHERE topic_name = :topic_name`

	ListSubjectQuery = `SELECT subject.id, university_id, subject.name, subject.number, subject.season, subject.year, subject.topic_name FROM subject JOIN university ON university.id = subject.university_id
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
