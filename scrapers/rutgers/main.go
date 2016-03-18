package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	uct "uct/common"
	"sync"
)

var (
	rutgersNB *uct.University
	host      = "http://sis.rutgers.edu/soc"
)

type (
	MeetingByClass []RMeetingTime

	RSubject struct {
		Name    string    `json:"description,omitempty"`
		Number  string    `json:"code,omitempty"`
		Courses []RCourse `json:"courses,omitempty"`
		Season  uct.Season
		Year    int
	}

	RCourse struct {
		SubjectNotes      string        `json:"subjectNotes"`
		CourseNumber      string        `json:"courseNumber"`
		Subject           string        `json:"subject"`
		CampusCode        string        `json:"campusCode"`
		OpenSections      int           `json:"openSections"`
		SynopsisURL       string        `json:"synopsisUrl"`
		SubjectGroupNotes string        `json:"subjectGroupNotes"`
		OfferingUnitCode  string        `json:"offeringUnitCode"`
		OfferingUnitTitle string        `json:"offeringUnitTitle"`
		Title             string        `json:"title"`
		CourseDescription string        `json:"courseDescription"`
		PreReqNotes       string        `json:"preReqNotes"`
		Sections          []RSection    `json:"sections"`
		SupplementCode    string        `json:"supplementCode"`
		Credits           float64       `json:"credits"`
		UnitNotes         string        `json:"unitNotes"`
		CoreCodes         []interface{} `json:"coreCodes"`
		CourseNotes       string        `json:"courseNotes"`
		ExpandedTitle     string        `json:"expandedTitle"`
	}

	RSection struct {
		SectionEligibility                   string                 `json:"sectionEligibility"`
		SessionDatePrintIndicator            string                 `json:"sessionDatePrintIndicator"`
		ExamCode                             string                 `json:"examCode"`
		SpecialPermissionAddCode             string                 `json:"specialPermissionAddCode"`
		CrossListedSections                  []RCrossListedSections `json:"crossListedSections"`
		SectionNotes                         string                 `json:"sectionNotes"`
		SpecialPermissionDropCode            string                 `json:"specialPermissionDropCode"`
		Instructor                           []RInstructor          `json:"instructors"`
		Number                               string                 `json:"number"`
		Majors                               []RMajor               `json:"majors"`
		SessionDates                         string                 `json:"sessionDates"`
		SpecialPermissionDropCodeDescription string                 `json:"specialPermissionDropCodeDescription"`
		Subtopic                             string                 `json:"subtopic"`
		SynopsisUrl                          string                 `json:"synopsisUrl"`
		OpenStatus                           bool                   `json:"openStatus"`
		Comments                             []RComment             `json:"comments"`
		Minors                               []interface{}          `json:"minors"`
		CampusCode                           string                 `json:"campusCode"`
		Index                                string                 `json:"index"`
		UnitMajors                           []interface{}          `json:"unitMajors"`
		Printed                              string                 `json:"printed"`
		SpecialPermissionAddCodeDescription  string                 `json:"specialPermissionAddCodeDescription"`
		Subtitle                             string                 `json:"subtitle"`
		MeetingTimes                         []RMeetingTime         `json:"meetingTimes"`
		LegendKey                            string                 `json:"legendKey"`
		HonorPrograms                        []interface{}          `json:"honorPrograms"`
	}

	RInstructor struct {
		Name string `json:"name"`
	}

	RMajor struct {
		isMajorCode bool   `json:"isMajorCode"`
		isUnitCode  bool   `json:"isUnitCode"`
		code        string `json:"code"`
	}

	RComment struct {
		Code        string `json:"code"`
		Description string `json:"description"`
	}

	RCrossListedSections struct {
		sectionNumber    string `json:"sectionNumber"`
		offeringUnitCode string `json:"offeringUnitCode"`
		courseNumber     string `json:"courseNumber"`
		subjectCode      string `json:"subjectCode"`
	}

	RMeetingTime struct {
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


	//NextSemester := Semesters[2]

	ThisSemester := Semesters.Current
	if ThisSemester.Season == uct.WINTER {
		ThisSemester.Year += 1
	}

	subjects := getSubjects(ThisSemester)
	var wg sync.WaitGroup
	wg.Add(len(subjects))
	for i, _ := range subjects {

		go func(sub *RSubject) {
			courses := getCourses(sub.Number, ThisSemester)

			for j, _ := range courses {
				sub.Courses = append(sub.Courses, courses[j])
			}
			wg.Done()
		}(&subjects[i])

	}
	wg.Wait()

	var dbSubjects []uct.Subject
	for _, subject := range subjects {
		newSubject := uct.Subject{
			Name:   subject.Name,
			Abbr:   subject.Name,
			Season: subject.Season,
			Year:   subject.Year}
		for _, course := range subject.Courses {
			newCourse := uct.Course{
				Name:     course.ExpandedTitle,
				Number:   course.CourseNumber,
				Synopsis: uct.ToNullString(course.CourseDescription),
				Metadata: course.metadata()}

			for _, section := range course.Sections {
				newSection := uct.Section{
					Number:     section.Number,
					CallNumber: section.Index,
					Status:     section.status(),
					Credits:    uct.FloatToString("%.1f", course.Credits),
					Max:        0,
					Metadata:   section.metadata()}

				for _, meeting := range section.MeetingTimes {
					newMeeting := uct.Meeting{
						Room:      meeting.room(),
						Day:       meeting.day(),
						StartTime: meeting.getMeetingHourBegin(),
						EndTime:   meeting.getMeetingHourEnd(),
						Metadata:  meeting.metadata()}

					newMeeting.VetAndBuild()
					newSection.Meetings = append(newSection.Meetings, newMeeting)
				}

				newSection.VetAndBuild()
				newCourse.Sections = append(newCourse.Sections, newSection)

			}
			newCourse.VetAndBuild()
			newSubject.Courses = append(newSubject.Courses, newCourse)
		}
		newSubject.VetAndBuild()
		dbSubjects = append(dbSubjects, newSubject)
	}

	bolB, _ := json.Marshal(dbSubjects)
	fmt.Printf("%s\n", bolB)

}

func getSubjects(semester uct.Semester) (subjects []RSubject) {
	var url = fmt.Sprintf("%s/subjects.json?semester=%s&campus=NB&level=U%sG", host, getRutgersSemester(semester), "%2C")
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

	for i, _ := range subjects {
		subjects[i].Name = strings.Title(strings.ToLower(subjects[i].Name))
		subjects[i].Season = semester.Season
		subjects[i].Year = semester.Year
	}

	defer resp.Body.Close()
	return
}

func getCourses(subject string, semester uct.Semester) (courses []RCourse) {
	var url = fmt.Sprintf("%s/courses.json?subject=%s&semester=%s&campus=NB&level=U%sG", host, subject, getRutgersSemester(semester), "%2C")
	uct.Log("GET  ", url)
	resp, err := http.Get(url)
	if err != nil {
		uct.Log(err)
		return
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&courses); err == io.EOF {
	} else if err != nil {
		uct.LogVerbose(err)
	}

	for i, _ := range courses {
		courses[i].clean()
		for j, _ := range courses[i].Sections {
			courses[i].Sections[j].clean()
			for k, _:= range courses[i].Sections[j].MeetingTimes {
				courses[i].Sections[j].MeetingTimes[k].clean()
			}
		}
	}

	courses = FilterCourses(courses, func(course RCourse) bool {
		return len(course.Sections) > 0
	})

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
		if !(val.Printed == "Y") && val.OpenStatus {
			num++
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

	if course.ExpandedTitle == "" {
		log.Fatal("WTF IS GOING ON!")
	}

	course.CourseNumber = uct.TrimAll(course.CourseNumber)

	course.CourseDescription = uct.TrimAll(course.CourseDescription)

	course.CourseNotes = uct.TrimAll(course.CourseNotes)

	course.SubjectNotes = uct.TrimAll(course.SubjectNotes)

	course.SynopsisURL = uct.TrimAll(course.SynopsisURL)

	course.PreReqNotes = uct.TrimAll(course.PreReqNotes)

}

func (section *RSection) clean() {
	section.Subtitle = uct.TrimAll(section.Subtitle)
	section.SectionNotes = uct.TrimAll(section.SectionNotes)
	section.CampusCode = uct.TrimAll(section.CampusCode)
	section.SpecialPermissionAddCodeDescription = uct.TrimAll(section.SpecialPermissionAddCodeDescription)

	sort.Stable(MeetingByClass(section.MeetingTimes))
}

func (meeting *RMeetingTime) clean() {
	meeting.StartTime = uct.TrimAll(meeting.StartTime)
	meeting.EndTime = uct.TrimAll(meeting.EndTime)

}

func (section *RSection) status() uct.Status {
	if section.OpenStatus {
		return uct.OPEN
	} else {
		return uct.CLOSED
	}
}

func (section RSection) instructor() (instructors []uct.Instructor) {
	for _, instructor := range section.Instructor {
		instructors = append(instructors, uct.Instructor{Name: instructor.Name})
	}
	return
}

func (section RSection) metadata() (metadata []uct.Metadata) {

	if len(section.CrossListedSections) > 0 {
		str := ""
		for _, cls := range section.CrossListedSections {
			str += cls.offeringUnitCode + ":" + cls.subjectCode + ":" + cls.courseNumber + ":" + cls.sectionNumber + ", "
		}
		metadata = append(metadata, uct.Metadata{
			Title:   "Cross-listed Sections",
			Content: str,
		})
	}

	if len(section.Comments) > 0 {
		str := ""
		for _, comment := range section.Comments {
			str += comment.Description + ", \n"
		}
		metadata = append(metadata, uct.Metadata{
			Title:   "Comments",
			Content: str,
		})
	}

	if len(section.Majors) > 0 {
		isMajorHeaderSet := false
		isUnitHeaderSet := false
		var openTo string
		for _, unit := range section.Majors {
			if unit.isMajorCode {
				if !isMajorHeaderSet {
					isMajorHeaderSet = true
					openTo = openTo + "Majors: "
				}
			} else if unit.isUnitCode {
				if !isUnitHeaderSet {
					isUnitHeaderSet = true
					openTo += "Schools: "
				}
				openTo += unit.code
				openTo += ", "
			}
			openTo += unit.code + ", "
		}
		metadata = append(metadata, uct.Metadata{
			Title:   "Open To",
			Content: openTo,
		})
	}

	if len(section.SectionNotes) > 0 {
		metadata = append(metadata, uct.Metadata{
			Title:   "Section Notes",
			Content: section.SectionNotes,
		})
	}

	if len(section.SynopsisUrl) > 0 {
		metadata = append(metadata, uct.Metadata{
			Title:   "Synopsis Url",
			Content: section.SynopsisUrl,
		})
	}

	if len(section.ExamCode) > 0 {
		metadata = append(metadata, uct.Metadata{
			Title:   "Exam Code",
			Content: section.ExamCode,
		})
	}

	if len(section.SpecialPermissionAddCodeDescription) > 0 {
		metadata = append(metadata, uct.Metadata{
			Title:   "Special Permission",
			Content: "Code: " + section.SpecialPermissionAddCode + "\n" + section.SpecialPermissionAddCodeDescription,
		})
	}

	if len(section.Subtitle) > 0 {
		metadata = append(metadata, uct.Metadata{
			Title:   "Subtitle",
			Content: section.Subtitle,
		})
	}

	if len(section.CampusCode) > 0 {
		metadata = append(metadata, uct.Metadata{
			Title:   "Campus Code",
			Content: section.CampusCode,
		})
	}

	return
}

func (meeting MeetingByClass) Len() int {
	return len(meeting)
}

func (meeting MeetingByClass) Swap(i, j int) {
	meeting[i], meeting[j] = meeting[j], meeting[i]
}

func (meeting MeetingByClass) Less(i, j int) bool {
	return meeting[i].classRank() < meeting[j].classRank()
}

func (meeting RMeetingTime) classRank() int {
	if meeting.isLecture() {
		return 1
	} else if meeting.isRecitation() {
		return 2
	} else if meeting.isByArrangement() {
		return 3
	} else if meeting.isLab() {
		return 4
	} else if meeting.isStudio() {
		return 5
	}
	return 99
}

func (meeting RMeetingTime) timeRank() int {
	switch meeting.MeetingDay {
	case "Monday":
		return 10
	case "Tuesday":
		return 9
	case "Wednesday":
		return 8
	case "Thurdsday":
		return 7
	case "Friday":
		return 6
	case "Saturday":
		return 5
	case "Sunday":
		return 4
	}
	return -1
}

func (meeting RMeetingTime) room() string {
	if meeting.BuildingCode != "" {
		return meeting.BuildingCode + " - " + meeting.RoomNumber
	}
	return ""
}

func (time RMeetingTime) getMeetingHourBegin() string {
	meridian := ""

	if time.PmCode != "" {
		if time.PmCode == "A" {
			meridian = "AM"
		} else {
			meridian = "PM"
		}
	}
	return formatMeetingHours(time.StartTime) + " " + meridian
}

func (time RMeetingTime) getMeetingHourEnd() string {
	if len(time.StartTime) > 1 || len(time.EndTime) > 1 {
		var meridian string
		starttime := time.StartTime
		endtime := time.EndTime
		pmcode := time.PmCode

		end, _ := strconv.Atoi(endtime[:2])
		start, _ := strconv.Atoi(starttime[:2])

		if !(pmcode == "A") {
			meridian = "PM"
		} else if end < start {
			meridian = "PM"
		} else if endtime[:2] == "12" {
			meridian = "AM"
		} else {
			meridian = "AM"
		}

		return formatMeetingHours(time.EndTime) + " " + meridian
	}
	return ""
}

func (meeting RMeetingTime) day() string {
	switch meeting.MeetingDay {
	case "M":
		return "Monday"
	case "T":
		return "Tuesday"
	case "W":
		return "Wednesday"
	case "TH":
		return "Thursday"
	case "F":
		return "Friday"
	case "S":
		return "Saturday"
	case "U":
		return "Sunday"
	}
	return ""
}

func (meeting RMeetingTime) metadata() (metadata []uct.Metadata) {
	if meeting.CampusAbbrev != "" {
		campus := ""
		switch meeting.CampusAbbrev {
		case "NWK":
			campus = "Newark"
			break
		case "CAM":
			campus = "Camden"
			break
		default:
			campus = "New Brunswick"
		}
		metadata = append(metadata, uct.Metadata{
			Title:   "Campus",
			Content: meeting.CampusAbbrev + " - " + campus,
		})
	}

	if meeting.MeetingModeCode != "" {
		metadata = append(metadata, uct.Metadata{
			Title:   "Class Type",
			Content: meeting.classType(),
		})
	}

	return
}

func (course RCourse) metadata() (metadata []uct.Metadata) {

	if course.SubjectNotes != "" {
		metadata = append(metadata, uct.Metadata{
			Title:   "Subject Notes",
			Content: course.SubjectNotes,
		})
	}
	if course.PreReqNotes != "" {
		metadata = append(metadata, uct.Metadata{
			Title:   "Prequisites",
			Content: course.PreReqNotes,
		})
	}
	if course.SynopsisURL != "" {
		metadata = append(metadata, uct.Metadata{
			Title:   "Synopsis Url",
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
