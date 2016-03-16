package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
	uct "uct/common"
)

var (
	rutgersNB *uct.University
	host      = "http://sis.rutgers.edu/soc"
)

type (
	RSubject struct {
		Name    string `json:"description,omitempty"`
		Number  string `json:"code,omitempty"`
		Courses []RCourse `json:"courses,omitempty"`
		Season uct.Season
		Year 	int
	}

	RCourse struct {
		SubjectNotes      string `json:"subjectNotes"`
		CourseNumber      string `json:"courseNumber"`
		Subject           string `json:"subject"`
		CampusCode        string `json:"campusCode"`
		OpenSections      int    `json:"openSections"`
		SynopsisURL       string `json:"synopsisUrl"`
		SubjectGroupNotes string `json:"subjectGroupNotes"`
		OfferingUnitCode  string `json:"offeringUnitCode"`
		OfferingUnitTitle string `json:"offeringUnitTitle"`
		Title             string `json:"title"`
		CourseDescription string `json:"courseDescription"`
		PreReqNotes       string `json:"preReqNotes"`
		Sections          []RSection `json:"sections"`
		SupplementCode    string   `json:"supplementCode"`
		Credits           float64      `json:"credits"`
		UnitNotes         string   `json:"unitNotes"`
		CoreCodes         []interface{} `json:"coreCodes"`
		CourseNotes       string   `json:"courseNotes"`
		ExpandedTitle     string   `json:"expandedTitle"`
	}

	RSection          struct {
		SectionEligibility        string   `json:"sectionEligibility"`
		SessionDatePrintIndicator string   `json:"sessionDatePrintIndicator"`
		ExamCode                  string   `json:"examCode"`
		SpecialPermissionAddCode  string   `json:"specialPermissionAddCode"`
		CrossListedSections       []interface{} `json:"crossListedSections"`
		SectionNotes              string   `json:"sectionNotes"`
		SpecialPermissionDropCode            string   `json:"specialPermissionDropCode"`
		Instructor                           []RInstructor `json:"instructors"`
		Number                               string   `json:"number"`
		Majors                               []interface{} `json:"majors"`
		SessionDates                         string   `json:"sessionDates"`
		SpecialPermissionDropCodeDescription string   `json:"specialPermissionDropCodeDescription"`
		Subtopic                             string   `json:"subtopic"`
		OpenStatus                           bool     `json:"openStatus"`
		Comments                             []RComment `json:"comments"`
		Minors                               []interface{} `json:"minors"`
		CampusCode                           string   `json:"campusCode"`
		Index                                string   `json:"index"`
		UnitMajors                           []interface{} `json:"unitMajors"`
		Printed                              string   `json:"printed"`
		SpecialPermissionAddCodeDescription  string   `json:"specialPermissionAddCodeDescription"`
		Subtitle                             string   `json:"subtitle"`
		MeetingTimes                        []RMeetingTime `json:"meetingTimes"`
		LegendKey                           string   `json:"legendKey"`
		HonorPrograms                       []interface{} `json:"honorPrograms"`
	}

	RInstructor               struct {
		Name string `json:"name"`
	}

	RComment                             struct {
		Code        string `json:"code"`
		Description string `json:"description"`
	}

	RMeetingTime                        struct {
		CampusLocation  string `json:"campusLocation"`
		BaClassHours    string `json:"baClassHours"`
		RoomNumber      string `json:"roomNumber"`
		PmCode          string `json:"pmCode"`
		CampusAbbrev    string `json:"campusAbbrev"`
		CampusName      string `json:"campusName"`
		MeetingDay      string `json:"meetingDay"`
		BuildingCode    string `json:"buildingCode"`
		StartTime       string `json:"startTime"`
		EndTime         string `json:"endTime"`
		MeetingModeDesc string `json:"meetingModeDesc"`
		MeetingModeCode string `json:"meetingModeCode"`
	}
)

func main() {

	rutgersNB = &uct.University{
		Name:             "Rutgers University–New Brunswick",
		Abbr:             "RU-NB",
		MainColor:        "F44336",
		AccentColor:      "607D8B",
		HomePage:         "http://newbrunswick.rutgers.edu/",
		RegistrationPage: "https://sims.rutgers.edu/webreg/",
		Registrations: []uct.Registration{
			uct.Registration{
				Period:     uct.SEM_FALL,
				PeriodDate: time.Date(0000, time.September, 6, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.SEM_SPRING,
				PeriodDate: time.Date(0000, time.January, 17, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.SEM_SUMMER,
				PeriodDate: time.Date(0000, time.May, 30, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.SEM_WINTER,
				PeriodDate: time.Date(0000, time.December, 23, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.START_FALL,
				PeriodDate: time.Date(0000, time.March, 20, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.START_SPRING,
				PeriodDate: time.Date(0000, time.October, 18, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.START_SUMMER,
				PeriodDate: time.Date(0000, time.January, 14, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.START_WINTER,
				PeriodDate: time.Date(0000, time.September, 21, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.END_FALL,
				PeriodDate: time.Date(0000, time.September, 13, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.END_SPRING,
				PeriodDate: time.Date(0000, time.January, 27, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.END_SUMMER,
				PeriodDate: time.Date(0000, time.August, 15, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.END_WINTER,
				PeriodDate: time.Date(0000, time.December, 22, 0, 0, 0, 0, time.UTC),
			},
		},
		Metadata: []uct.Metadata{
			uct.Metadata{
				Title: "About", Content: `<p><b>Rutgers University–New Brunswick</b> is the oldest campus of <a href="/wiki/Rutgers_Uni
				versity" title="Rutgers University">Rutgers University</a>, the others being in <a href="/wiki/Rutgers%
				E2%80%93Camden" title="Rutgers–Camden" class="mw-redirect">Camden</a> and <a href="/wiki/Rutgers%E2%80%
				93Newark" title="Rutgers–Newark" class="mw-redirect">Newark</a>. It is primarily located in the <a href
				="/wiki/New_Brunswick,_New_Jersey" title="New Brunswick, New Jersey">City of New Brunswick</a> and <a h
				ref="/wiki/Piscataway_Township,_New_Jersey" title="Piscataway Township, New Jersey" class="mw-redirect"
				>Piscataway Township</a>. The campus is composed of several smaller campuses: <i>College Avenue</i>, <i
				><a href="/wiki/Busch_Campus_(Rutgers_University)" title="Busch Campus (Rutgers University)" class="mw-
				redirect">Busch</a></i>, <i>Livingston,</i> <i>Cook</i>, and <i>Douglass</i>, the latter two sometimes
				referred to as "Cook/Douglass," as they are adjacent to each other. Rutgers–New Brunswick also includes
				 several buildings in downtown New Brunswick. The New Brunswick campuses include 19 undergraduate, grad
				 uate and professional schools, including the School of Arts and Sciences, School of Environmental and
				 Biological Sciences, School of Communication, Information and Library Studies, the <a href="/wiki/Edwa
				 rd_J._Bloustein_School_of_Planning_and_Public_Policy" title="Edward J. Bloustein School of Planning an
				 d Public Policy">Edward J. Bloustein School of Planning and Public Policy</a>, <a href="/wiki/School_o
				 f_Engineering_(Rutgers_University)" title="School of Engineering (Rutgers University)" class="mw-redir
				 ect">School of Engineering</a>, the <a href="/wiki/Ernest_Mario_School_of_Pharmacy" title="Ernest Mari
				 o School of Pharmacy">Ernest Mario School of Pharmacy</a>, the Graduate School, the Graduate School of
				  Applied and Professional Psychology, the Graduate School of Education, <a href="/wiki/School_of_Manag
				  ement_and_Labor_Relations" title="School of Management and Labor Relations" class="mw-redirect">School
				   of Management and Labor Relations</a>, the <a href="/wiki/Mason_Gross_School_of_the_Arts" title="Maso
				   n Gross School of the Arts">Mason Gross School of the Arts</a>, the College of Nursing, the <a href="
				   /wiki/Rutgers_Business_School" title="Rutgers Business School" class="mw-redirect">Rutgers Business
				   School</a> and the <a href="/wiki/School_of_Social_Work_(Rutgers_University)" title="School of Social
				    Work (Rutgers University)" class="mw-redirect">School of Social Work</a>.</p>`,
			},
		},
	}

	Semesters := uct.ResolveSemesters(time.Now(), rutgersNB.Registrations)

	//CurrentSemester := Semesters[1]
	//NextSemester := Semesters[2]

	subjects := getSubjects(Semesters.Last)
	for _, val := range subjects {
		for _,v := range getCourses(val.Number, Semesters.Last) {
			val.Courses = append(val.Courses, v)
		}
	}

	var dbSubjects []uct.Subject
	for _, val := range subjects {
		newSubject := uct.Subject{Name:val.Name,Abbr:val.Name, Season: val.Season, Year:time.Date(val.Year, 0,0,0,0,0,0, time.UTC)}
		for _, course := range val.Courses {
			newCourse := uct.Course{Name:course.ExpandedTitle, Number:course.CourseNumber, Synopsis:uct.ToNullString(course.CourseDescription), Metadata:course.metadata()}
			newSubject.Courses = append(newSubject.Courses, newCourse)
		}

		dbSubjects = append(dbSubjects, newSubject)
	}

	bolB, _ := json.MarshalIndent(dbSubjects, "", "    ")
	fmt.Printf("%s\n", bolB)

}

func getSubjects(semester uct.Semester) (subjects []RSubject) {
	var url = fmt.Sprintf("%s/subjects.json?semester=%s&campus=NB&level=U,G", host, getRutgersSemester(semester))
	uct.Log("GET  ", url)
	resp, err := http.Get(url)
	if err != nil {
		uct.Log(err)
		return
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&subjects); err == io.EOF {
	} else if err != nil {
		log.Fatal(err)
	}

	for _, val := range subjects {
		val.Season = semester.Season
		val.Year = semester.Year
	}

	defer resp.Body.Close()
	return
}

func getCourses(subject string, semester uct.Semester) (courses []RCourse) {
	var url = fmt.Sprintf("%s/courses.json?subject=%s&semester=%s&campus=NB&level=U,G", host, subject, getRutgersSemester(semester))
	uct.Log("GET  ", url)
	resp, err := http.Get(url)
	if err != nil {
		uct.Log(err)
		return
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&courses); err == io.EOF {
	} else if err != nil {
		uct.Log(err)
	}

	for _, val := range courses {
		val.clean()
	}

	courses = FilterCourses(courses, func(course RCourse) bool {
		return len(course.Sections) > 0
	} )

	defer resp.Body.Close()
	return
}

func getRutgersSemester(semester uct.Semester) string {
	if semester.Season == uct.FALL {
		return "9" + strconv.Itoa(semester.Year)
	} else if semester.Season == uct.SUMMER {
		return "7" + strconv.Itoa(semester.Year)
	} else if semester.Season == uct.SPRING {
		return "1" + strconv.Itoa(semester.Year)
	} else if semester.Season == uct.WINTER {
		return "0" + strconv.Itoa(semester.Year)
	}
	return ""
}

func (course *RCourse) clean() {
	// Open Sections
	num := 0
	for _, val := range course.Sections {
		if (!(val.Printed=="Y") && val.OpenStatus) {
			num++;
		}
	}
	course.OpenSections = course.OpenSections - num

	course.Sections = FilterSections(course.Sections, func(section RSection) bool {
		return section.Printed == "Y"
	})

	course.ExpandedTitle = uct.TrimAll(course.ExpandedTitle)
	if len(course.ExpandedTitle) == 0 {
		course.ExpandedTitle = course.Title
	}

	course.CourseNumber = uct.TrimAll(course.CourseNumber)

	course.CourseDescription = uct.TrimAll(course.CourseDescription)

	course.CourseNotes = uct.TrimAll(course.CourseNotes)

	course.SubjectNotes = uct.TrimAll(course.SubjectNotes)

	course.SynopsisURL = uct.TrimAll(course.SynopsisURL)

	course.PreReqNotes = uct.TrimAll(course.PreReqNotes)

}

func (course RCourse) metadata() (metadata []uct.Metadata) {

	if course.SubjectNotes != "" {
		metadata = append(metadata, uct.Metadata{
			Title: "Subject Notes",
			Content: course.SubjectNotes,
		})
	}
	if course.PreReqNotes != "" {
		metadata = append(metadata, uct.Metadata{
			Title: "Prequisites",
			Content: course.PreReqNotes,
		})
	}
	if course.SynopsisURL != "" {
		metadata = append(metadata, uct.Metadata{
			Title: "Synopsis",
			Content: course.SynopsisURL,
		})
	}

	return metadata
}

func FilterCourses(vs []RCourse, f func(RCourse) bool) []RCourse {
	vsf := make([]RCourse, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func FilterSections(vs []RSection, f func(RSection) bool) []RSection {
	vsf := make([]RSection, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func AppendRSubjects(subjects []RSubject, toAppend []RSubject) []RSubject {
	for _, val := range toAppend {
		subjects = append(subjects, val)
	}
	return subjects
}
