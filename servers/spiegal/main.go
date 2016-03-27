package main

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"net/http"
"log"
	"github.com/jmoiron/sqlx"
	"uct/common"
	"github.com/alecthomas/template/parse"
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
	var deep bool
	var err error
	var id int64

	if deep, err = strconv.ParseBool(c.DefaultQuery("deep", "true")); err != nil {
		c.String(http.StatusBadRequest, "Could not parse parameter: deep")
	}

	if id, err = strconv.ParseInt(c.Query("id"), 10, 64); err != nil {
		c.String(http.StatusBadRequest, "Could not parse parameter: id")
	}

	if s := c.Request.Header.Get("Accept"); s == "" || s != "application/json" {
		c.String(http.StatusBadRequest, "Missing header, Accept: application/json")
	}


}

func subjectHandler(c *gin.Context) {

}

func courseHandler(c *gin.Context) {


}

func sectionHandler(c *gin.Context) {


}

func SelectUniversity(deep bool) (universities []common.University) {
	err := database.Select(&universities,
		`SELECT id, name, abbr, home_page, registrations_page, main_color, accent_color, topic_name
		 FROM university ORDER BY name`)
	common.CheckError(err)


}

func SelectCourses(subjectId int64, deep bool) (courses []*common.Course) {
	err := database.Select(courses, `SELECT * FROM course WHERE subject_id = $1`,subjectId)
	common.CheckError(err)

	if deep {
		for i := range courses {
			courses[i].Sections = SelectSections(courses[i].Id, deep)
			courses[i].Metadata = SelectMetadata(0,0,courses[i].Id, 0,0)
		}
	}
	return
}

func SelectSection(section_id int64, deep bool) (section *common.Section) {
	err := database.Select(section, `SELECT * FROM section WHERE id = $1`, section_id)
	common.CheckError(err)

	if deep && section != nil {
		deepSelectSection(section)
	}
	
	return
}

func SelectSections(course_id int64, deep bool) (sections []*common.Section) {
	err := database.Select(sections, `SELECT * FROM section WHERE course_id = $1`, course_id)
	common.CheckError(err)

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
	section.Metadata = SelectMetadata(0,0,0,section.Id,0)
}

func SelectMeetings(sectionId int64) (meetings []*common.Meeting) {
	err := database.Select(meetings, `SELECT * FROM meetings WHERE section_id = $1 ORDER BY index`, sectionId)
	common.CheckError(err)

	for i := range meetings {
		meetings[i].Metadata = SelectMetadata(0,0,0,0,meetings[i].Id)
	}
	return
}

func SelectInstructors(sectionId int64) (instructors []*common.Instructor) {
	err := database.Select(instructors, `SELECT * FROM instructor WHERE section_id = $1`, sectionId)
	common.CheckError(err)
	return
}

func SelectBooks(sectionId int64) (books []*common.Book) {
	err := database.Select(books,`SELECT * FROM book WHERE section_id = $1`, sectionId)
	common.CheckError(err)
	return
}

func SelectMetadata(universityId, subjectId, courseId, sectionId, meetingId int64) (metadata []*common.Metadata) {
	var err error
	if universityId != 0 {
		err = database.Select(metadata, `SELECT * FROM meeting WHERE university_id = $1`, universityId)
	} else if subjectId != 0 {
		err = database.Select(metadata, `SELECT * FROM meeting WHERE subject_id = $1`, subjectId)
	} else if courseId != 0 {
		err = database.Select(metadata, `SELECT * FROM meeting WHERE course_id = $1`, courseId)
	} else if sectionId != 0 {
		err = database.Select(metadata, `SELECT * FROM meeting WHERE section_id = $1`, sectionId)
	} else if sectionId != 0 {
		err = database.Select(metadata, `SELECT * FROM meeting WHERE meeting_id = $1`, meetingId)
	} else {
		log.Panic("No valid metadata id was passed")
	}
	common.CheckError(err)
	return
}