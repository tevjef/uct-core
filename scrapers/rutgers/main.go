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
	"io/ioutil"
)

var (
	rutgersNB *uct.University
	host      = "http://sis.rutgers.edu/soc"
)

type (
	RSubject struct {
		Name   string `json:"description,omitempty"`
		Number string `json:"code,omitempty"`
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
		Sections          []struct {
			SectionEligibility        string   `json:"sectionEligibility"`
			SessionDatePrintIndicator string   `json:"sessionDatePrintIndicator"`
			ExamCode                  string   `json:"examCode"`
			SpecialPermissionAddCode  string   `json:"specialPermissionAddCode"`
			CrossListedSections       []string `json:"crossListedSections"`
			SectionNotes              string   `json:"sectionNotes"`
			SpecialPermissionDropCode string   `json:"specialPermissionDropCode"`
			Instructors               []struct {
				Name string `json:"name"`
			} `json:"instructors"`
			Number                               string   `json:"number"`
			Majors                               []string `json:"majors"`
			SessionDates                         string   `json:"sessionDates"`
			SpecialPermissionDropCodeDescription string   `json:"specialPermissionDropCodeDescription"`
			Subtopic                             string   `json:"subtopic"`
			OpenStatus                           bool     `json:"openStatus"`
			Comments                             []struct {
				Code        string `json:"code"`
				Description string `json:"description"`
			} `json:"comments"`
			Minors                              []string `json:"minors"`
			CampusCode                          string   `json:"campusCode"`
			Index                               string   `json:"index"`
			UnitMajors                          []string `json:"unitMajors"`
			Printed                             string   `json:"printed"`
			SpecialPermissionAddCodeDescription string   `json:"specialPermissionAddCodeDescription"`
			Subtitle                            string   `json:"subtitle"`
			MeetingTimes                        []struct {
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
			} `json:"meetingTimes"`
			LegendKey     string   `json:"legendKey"`
			HonorPrograms []string `json:"honorPrograms"`
		} `json:"sections"`
		SupplementCode string   `json:"supplementCode"`
		Credits        int      `json:"credits"`
		UnitNotes      string   `json:"unitNotes"`
		CoreCodes      []string `json:"coreCodes"`
		CourseNotes    string   `json:"courseNotes"`
		ExpandedTitle  string   `json:"expandedTitle"`
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
		getCourses(val.Number, Semesters.Last)
	}
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

	b, err := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(b, &courses)
	/*dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&courses); err == io.EOF {
	} else if err != nil {
		log.Fatal(err)
	}*/

	defer resp.Body.Close()

	bolB, _ := json.Marshal(courses)
	fmt.Printf("%s", bolB)
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

func AppendRSubjects(subjects []RSubject, toAppend []RSubject) []RSubject {
	for _, val := range toAppend {
		subjects = append(subjects, val)
	}
	return subjects
}
