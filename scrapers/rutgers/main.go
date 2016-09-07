package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	rs "uct/redis/sync"
	"time"
	uct "uct/common"
	"io/ioutil"
	"uct/redis"
	"github.com/satori/go.uuid"
	"errors"
)

var (
	host = "http://sis.rutgers.edu/soc"
)

var (
	app     = kingpin.New("rutgers", "A web scraper that retrives course information for Rutgers University's servers.")
	campus  = app.Flag("campus", "Choose campus code. NB=New Brunswick, CM=Camden, NK=Newark").HintOptions("CM", "NK", "NB").Short('u').PlaceHolder("[CM, NK, NB]").Required().String()
	format  = app.Flag("format", "Choose output format").Short('f').HintOptions(uct.PROTOBUF, uct.JSON).PlaceHolder("[protobuf, json]").Required().String()
	daemonInterval = app.Flag("daemon", "Run as a daemon with a refesh interval").Duration()
	daemonFile = app.Flag("daemon-dir", "If supplied the deamon will write files to this directory").ExistingDir()
	latest = app.Flag("latest", "Only output the current and next semester").Short('l').Bool()
	verbose = app.Flag("verbose", "Verbose log of object representations.").Short('v').Bool()
	configFile    = app.Flag("config", "configuration file for the application").Short('c').File()
	config uct.Config
	wrapper *v1.RedisWrapper
	rsync *rs.RedisSync
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
	campus := strings.ToLower(*campus)
	app.Name = app.Name + "-" + campus

	if *format != uct.JSON && *format != uct.PROTOBUF {
		log.Fatalln("Invalid format:", *format)
	}

	isDaemon := *daemonInterval > 0
	// Parse configuration file
	config = uct.NewConfig(*configFile)

	// Start profiling
	go uct.StartPprof(config.GetDebugSever(app.Name))

	// Channel to send scraped data on
	resultChan := make(chan uct.University)

	// Runs at regular intervals
	if isDaemon {
		go startDaemon(resultChan, make(chan bool))
	} else {
		go func() {
			entryPoint(resultChan)
			close(resultChan)
		}()
	}

	var school uct.University
	var reader *bytes.Reader

	for school = range resultChan {
		reader = uct.MarshalMessage(*format, school)

		// Push to redis
		if isDaemon {
			pushToRedis(reader)
		}

		if *daemonFile != "" {
			if data, err := ioutil.ReadAll(reader); err != nil {
				uct.CheckError(err)
			} else {
				fileName := *daemonFile +  "/" + app.Name + "-" + strconv.FormatInt(time.Now().Unix(), 10) + "." + *format
				log.Debugln("Writing file", fileName)
				if err = ioutil.WriteFile(fileName, data, 0644); err != nil {
					uct.CheckError(err)
				}
			}
		}
	}

	// Runs when the channel closes, the channel will not close in daemon mode
	io.Copy(os.Stdout, reader)
}

func pushToRedis(reader *bytes.Reader) {
	if data, err := ioutil.ReadAll(reader); err != nil {
		uct.CheckError(err)
	} else {
		if err := wrapper.Client.Set(wrapper.NameSpace + ":data:latest", data, 0).Err(); err != nil {
			log.Panicln(errors.New("failed to connect to redis server"))
		}
	}
}

func startDaemon(result chan uct.University, cancel chan bool) {
	// Override cli arg with environment variable
	if intervalFromEnv := config.Scrapers.Get(app.Name).Interval; intervalFromEnv != "" {
		if interval, err := time.ParseDuration(intervalFromEnv); err != nil {
			uct.CheckError(err)
		} else if interval > 0 {
			daemonInterval = &interval
		}
	}

	// Start redis client
	wrapper = v1.New(config, app.Name)
	rsync = rs.New(wrapper, *daemonInterval, uuid.NewV4().String())

	offsetChan := rsync.Sync(make(chan bool))

	var offset time.Duration = 0

	go func() {
		for {
			offset = <-offsetChan
		}
	}()

	delayScrape := func() {
		log.Println("Sleeping for", offset.String())
		time.Sleep(offset)
		entryPoint(result)
	}

	ticker := time.NewTicker(*daemonInterval)
	delayScrape()
	for range ticker.C {
		go delayScrape()
	}

}

func entryPoint(result chan uct.University) {
	var school uct.University

	campus := strings.ToUpper(*campus)
	if campus == "CM" {
		school = getCampus("CM")
	} else if campus == "NK" {
		school = getCampus("NK")
	} else if campus == "NB" {
		school = getCampus("NB")
	} else {
		log.Fatalln("Invalid campus code:", campus)
	}

	result <- school
}

func getCampus(campus string) uct.University {
	var university uct.University

	university = uct.University{
		Name:             "Rutgers University–New Brunswick",
		Abbr:             "RU-NB",
		MainColor:        "F44336",
		AccentColor:      "607D8B",
		HomePage:         "http://newbrunswick.rutgers.edu/",
		RegistrationPage: "https://sims.rutgers.edu/webreg/",
		Registrations: []*uct.Registration{
			{
				Period:     uct.SEM_FALL.String(),
				PeriodDate: time.Date(2000, time.September, 6, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     uct.SEM_SPRING.String(),
				PeriodDate: time.Date(2000, time.January, 17, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     uct.SEM_SUMMER.String(),
				PeriodDate: time.Date(2000, time.May, 30, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     uct.SEM_WINTER.String(),
				PeriodDate: time.Date(2000, time.December, 23, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     uct.START_FALL.String(),
				PeriodDate: time.Date(2000, time.March, 20, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     uct.START_SPRING.String(),
				PeriodDate: time.Date(2000, time.October, 18, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     uct.START_SUMMER.String(),
				PeriodDate: time.Date(2000, time.January, 14, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     uct.START_WINTER.String(),
				PeriodDate: time.Date(2000, time.September, 21, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     uct.END_FALL.String(),
				PeriodDate: time.Date(2000, time.September, 13, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     uct.END_SPRING.String(),
				PeriodDate: time.Date(2000, time.January, 27, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     uct.END_SUMMER.String(),
				PeriodDate: time.Date(2000, time.June, 15, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     uct.END_WINTER.String(),
				PeriodDate: time.Date(2000, time.December, 22, 0, 0, 0, 0, time.UTC).Unix(),
			},
		},
		Metadata: []*uct.Metadata{
			{
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

	university.ResolvedSemesters = uct.ResolveSemesters(time.Now(), university.Registrations)

	Semesters := []*uct.Semester{
		university.ResolvedSemesters.Last,
		university.ResolvedSemesters.Current,
		university.ResolvedSemesters.Next}

	if *latest {
		Semesters = []*uct.Semester{
			university.ResolvedSemesters.Current,
			university.ResolvedSemesters.Next}
	}

	for _, ThisSemester := range Semesters {
		if ThisSemester.Season == uct.WINTER {
			ThisSemester.Year += 1
		}

		// Shadow copy variable
		ThisSemester := ThisSemester
		subjects := getSubjects(ThisSemester, campus)

		var wg sync.WaitGroup
		for i := range subjects {
			wg.Add(1)
			go func(sub *RSubject) {
				defer func() {
					wg.Done()
				}()
				courses := getCourses(sub.Number, campus, ThisSemester)
				for j := range courses {
					sub.Courses = append(sub.Courses, courses[j])
				}

			}(&subjects[i])

		}
		wg.Wait()
		for _, subject := range subjects {
			newSubject := &uct.Subject{
				Name:   subject.Name,
				Number: subject.Number,
				Season: subject.Season,
				Year:   strconv.Itoa(subject.Year)}
			for _, course := range subject.Courses {
				newCourse := &uct.Course{
					Name:     course.ExpandedTitle,
					Number:   course.CourseNumber,
					Synopsis: course.synopsis(),
					Metadata: course.metadata()}

				for _, section := range course.Sections {
					newSection := &uct.Section{
						Number:     section.Number,
						CallNumber: section.Index,
						Status:     section.status(),
						Credits:    uct.FloatToString("%.1f", course.Credits),
						Max:        0,
						Metadata:   section.metadata()}

					for _, instructor := range section.Instructor {
						newInstructor := &uct.Instructor{Name: instructor.Name}

						newSection.Instructors = append(newSection.Instructors, newInstructor)
					}

					for _, meeting := range section.MeetingTimes {
						newMeeting := &uct.Meeting{
							Room:      meeting.room(),
							Day:       meeting.dayPointer(),
							StartTime: meeting.PStartTime,
							EndTime:   meeting.PEndTime,
							ClassType: meeting.classType(),
							Metadata:  meeting.metadata()}

						newSection.Meetings = append(newSection.Meetings, newMeeting)
					}

					newCourse.Sections = append(newCourse.Sections, newSection)

				}
				newSubject.Courses = append(newSubject.Courses, newCourse)
			}
			university.Subjects = append(university.Subjects, newSubject)
		}
	}

	if campus == "NK" {
		university.Name = "Rutgers University–Newark"
		university.Abbr = "RU-NK"
		university.HomePage = "http://www.newark.rutgers.edu/"
		university.Metadata = []*uct.Metadata{
			{
				Title: "About", Content: `<p><b>Rutgers–Newark</b> is one of three regional campuses of <a href="/wiki/R
							utgers_University" title="Rutgers University">Rutgers University</a>, the <a href="/wiki/Public_universit
							y" title="Public university">public</a> research university of the <a href="/wiki/U.S._state" title="U.S
							. state">U.S. state</a> of <a href="/wiki/New_Jersey" title="New Jersey">New Jersey</a>, located in the
							 city of <a href="/wiki/Newark,_New_Jersey" title="Newark, New Jersey">Newark</a>. Rutgers, founded in 1
							 766 in <a href="/wiki/New_Brunswick,_New_Jersey" title="New Brunswick, New Jersey">New Brunswick</a>, i
							 s the <a href="/wiki/Colonial_colleges" title="Colonial colleges" class="mw-redirect">eighth oldest col
							 lege in the United States</a> and a member of the <a href="/wiki/Association_of_American_Universities"
							 title="Association of American Universities">Association of American Universities</a>. In 1945, the sta
							 te legislature voted to make Rutgers University, then a private <a href="/wiki/Liberal_arts_college" ti
							 tle="Liberal arts college">liberal arts college</a>, into the state university and the following year m
							 erged the school with the former <a href="/wiki/University_of_Newark" title="University of Newark" clas
							 s="mw-redirect">University of Newark</a> (1936–1946), which became the Rutgers–Newark campus. Rutgers a
							 lso incorporated the College of South Jersey and South Jersey Law School, in Camden, as a constituent c
							 ampus of the university and renamed it <a href="/wiki/Rutgers%E2%80%93Camden" title="Rutgers–Camden" cl
							 ass="mw-redirect">Rutgers–Camden</a> in 1950.</p> <p>Rutgers–Newark offers undergraduate (bachelors) an
							 d graduate (masters, doctoral) programs to more than 11,000 students. The campus is located on 38 acre
							 s in Newark's <a href="/wiki/University_Heights,_Newark,_New_Jersey" title="University Heights, Newark
							 , New Jersey" class="mw-redirect">University Heights</a> section. It consists of seven degree-granting
							  undergraduate, graduate, and professional schools, including the <a href="/wiki/Rutgers_Business_Schoo
							  l" title="Rutgers Business School" class="mw-redirect">Rutgers Business School</a> and <a href="/wiki/
							  Rutgers_School_of_Law_-_Newark" title="Rutgers School of Law - Newark" class="mw-redirect">Rutgers Sch
							  ool of Law - Newark</a>, and several research institutes including the <a href="/wiki/Institute_of_Ja
							  zz_Studies" title="Institute of Jazz Studies">Institute of Jazz Studies</a>. According to <i>U.S. News
							   &amp; World Report</i>, Rutgers–Newark is the most <a href="/wiki/Cultural_diversity" title="Cultural
								diversity">diverse</a> national university in the United States.</p>`,
			},
		}
	}
	if campus == "CM" {
		university.Name = "Rutgers University–Camden"
		university.Abbr = "RU-CAM"
		university.HomePage = "http://www.camden.rutgers.edu/"
		university.Metadata = []*uct.Metadata{
			{
				Title: "About", Content: `<p><b>Rutgers University–Camden</b> is one of three regional campuses of <a
							href="/wiki/Rutgers_University" title="Rutgers University">Rutgers University</a>, the <a href="/wiki/N
							ew_Jersey" title="New Jersey">New Jersey</a>'s <a href="/wiki/Public_university" title="Public universit
							y">public</a> <a href="/wiki/Research_university" title="Research university" class="mw-redirect">resear
							ch university</a>. It is located in <a href="/wiki/Camden,_New_Jersey" title="Camden, New Jersey">Camden
							</a>, New Jersey, <a href="/wiki/United_States" title="United States">United States</a>. Founded in the
							1920s, Rutgers–Camden began as an amalgam of the South Jersey Law School and the College of South Jersey
							. It is the southernmost of the three regional campuses of Rutgers—the others being located in <a href="
							/wiki/New_Brunswick,_New_Jersey" title="New Brunswick, New Jersey">New Brunswick</a> and <a href="/wiki/
							Newark,_New_Jersey" title="Newark, New Jersey">Newark</a>.<sup id="cite_ref-3" class="reference"><a href
							="#cite_note-3"><span>[</span>3<span>]</span></a></sup> The city of Camden is located on the <a href="/w
							iki/Delaware_River" title="Delaware River">Delaware River</a> east of <a href="/wiki/Philadelphia,_Penn
							sylvania" title="Philadelphia, Pennsylvania" class="mw-redirect">Philadelphia</a>.</p>`,
			},
		}
	}
	return university
}

func getSubjects(semester *uct.Semester, campus string) (subjects []RSubject) {
	var url = fmt.Sprintf("%s/subjects.json?semester=%s&campus=%s&level=U%sG", host, getRutgersSemester(semester), campus, "%2C")

	for i := 0; i < 3; i++ {
		log.WithFields(log.Fields{"season": semester.Season, "year": semester.Year, "campus": campus, "retry": i, "url": url}).Debug("Subject Request")
		resp, err := http.Get(url)
		if err != nil {
			log.Errorln(err)
			continue
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err := json.Unmarshal(data, &subjects); err != nil && err != io.EOF {
			log.Errorln(err)
			resp.Body.Close()
			continue
		}

		log.WithFields(log.Fields{"content-length": len(data), "status": resp.Status, "season": semester.Season,
			"year": semester.Year, "campus": campus, "url": url}).Debugln("Subject Reponse")

		resp.Body.Close()
		break
	}

	for i := range subjects {
		subjects[i].Name = strings.Title(strings.ToLower(subjects[i].Name))
		subjects[i].Season = semester.Season
		subjects[i].Year = int(semester.Year)
	}
	return
}

func getCourses(subject, campus string, semester *uct.Semester) (courses []RCourse) {
	var url = fmt.Sprintf("%s/courses.json?subject=%s&semester=%s&campus=%s&level=U%sG", host, subject, getRutgersSemester(semester), campus, "%2C")
	for i := 0; i < 3; i++ {
		log.WithFields(log.Fields{"subject" : subject, "season": semester.Season, "year": semester.Year, "campus": campus, "retry": i, "url": url}).Debug("Course Request")

		resp, err := http.Get(url)
		if err != nil {
			log.Errorf("Retrying %s after error: %s\n", i, err)
			continue
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err := json.Unmarshal(data, &courses); err != nil && err != io.EOF {
			resp.Body.Close()
			log.Errorf("Retrying %s after error: %s\n", i, err)
			continue
		}

		log.WithFields(log.Fields{"content-length": len(data), "subject" : subject, "status": resp.Status, "season": semester.Season,
			"year": semester.Year, "campus": campus,"url": url}).Debugln("Course Reponse")

		resp.Body.Close()
		break
	}

	for i := range courses {
		courses[i].clean()
		for j := range courses[i].Sections {
			courses[i].Sections[j].clean()
		}

		sort.Sort(sectionSorter{courses[i].Sections})
	}
	sort.Sort(courseSorter{courses})

	courses = FilterCourses(courses, func(course RCourse) bool {
		return len(course.Sections) > 0
	})

	return
}

func getRutgersSemester(semester *uct.Semester) string {
	if semester.Season == uct.FALL {
		return "9" + strconv.Itoa(int(semester.Year))
	} else if semester.Season == uct.SUMMER {
		return "7" + strconv.Itoa(int(semester.Year))
	} else if semester.Season == uct.SPRING {
		return "1" + strconv.Itoa(int(semester.Year))
	} else if semester.Season == uct.WINTER {
		return "0" + strconv.Itoa(int(semester.Year))
	}
	return ""
}

func (course *RCourse) clean() {
	course.Sections = FilterSections(course.Sections, func(section RSection) bool {
		return section.Printed == "Y"
	})

	m := map[string]int{}

	// Filter duplicate sections, yes it happens e.g Fall 2016 NB Biochem Engin SR DESIGN I PROJECTS
	course.Sections = FilterSections(course.Sections, func(section RSection) bool {
		key := section.Index + section.Number
		m[key]++
		return m[key] <= 1
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

func (section *RSection) clean() {
	section.Subtitle = uct.TrimAll(section.Subtitle)
	section.SectionNotes = uct.TrimAll(section.SectionNotes)
	section.CampusCode = uct.TrimAll(section.CampusCode)
	section.SpecialPermissionAddCodeDescription = uct.TrimAll(section.SpecialPermissionAddCodeDescription)

	for i := range section.MeetingTimes {
		section.MeetingTimes[i].clean()
	}

	sort.Sort(MeetingByClass(section.MeetingTimes))
}

func (meeting *RMeetingTime) clean() {
	meeting.StartTime = strings.TrimSpace(uct.TrimAll(meeting.StartTime))
	meeting.EndTime = strings.TrimSpace(uct.TrimAll(meeting.EndTime))

	meeting.MeetingDay = meeting.day()
	meeting.StartTime = meeting.getMeetingHourBegin()
	meeting.EndTime = meeting.getMeetingHourEnd()

	if meeting.StartTime != "" {
		t := meeting.StartTime
		meeting.PStartTime = &t
	} else {
		meeting.PStartTime = nil
	}
	if meeting.EndTime != "" {
		t := meeting.EndTime
		meeting.PEndTime = &t
	} else {
		meeting.PEndTime = nil
	}
}

func (section *RSection) status() string {
	if section.OpenStatus {
		return uct.OPEN.String()
	} else {
		return uct.CLOSED.String()
	}
}

func (section RSection) instructor() (instructors []*uct.Instructor) {
	for _, instructor := range section.Instructor {
		instructors = append(instructors, &uct.Instructor{Name: instructor.Name})
	}
	return
}

func (section RSection) metadata() (metadata []*uct.Metadata) {

	if len(section.CrossListedSections) > 0 {
		str := ""
		for _, cls := range section.CrossListedSections {
			str += cls.offeringUnitCode + ":" + cls.subjectCode + ":" + cls.courseNumber + ":" + cls.sectionNumber + ", "
		}
		if len(str) != 5 {
			metadata = append(metadata, &uct.Metadata{
				Title:   "Cross-listed Sections",
				Content: str,
			})
		}

	}

	if len(section.Comments) > 0 {
		sort.Sort(commentSorter{section.Comments})
		str := ""
		for _, comment := range section.Comments {
			str += (comment.Description + ", ")
		}
		str = str[:len(str)-2]
		metadata = append(metadata, &uct.Metadata{
			Title:   "Comments",
			Content: str,
		})
	}

	if len(section.Majors) > 0 {
		isMajorHeaderSet := false
		isUnitHeaderSet := false
		var buffer bytes.Buffer
		for _, unit := range section.Majors {
			if unit.isMajorCode {
				if !isMajorHeaderSet {
					isMajorHeaderSet = true
					buffer.WriteString("Majors: ")
				}
				buffer.WriteString(unit.code)
				buffer.WriteString(", ")
			} else if unit.isUnitCode {
				if !isUnitHeaderSet {
					isUnitHeaderSet = true
					buffer.WriteString("Schools: ")
				}
				buffer.WriteString(unit.code)
				buffer.WriteString(", ")
			}
		}

		openTo := buffer.String()
		if len(openTo) > len("Majors: ") {
			metadata = append(metadata, &uct.Metadata{
				Title:   "Open To",
				Content: openTo,
			})
		}
	}

	if len(section.SectionNotes) > 0 {
		metadata = append(metadata, &uct.Metadata{
			Title:   "Section Notes",
			Content: section.SectionNotes,
		})
	}

	if len(section.SynopsisUrl) > 0 {
		metadata = append(metadata, &uct.Metadata{
			Title:   "Synopsis Url",
			Content: section.SynopsisUrl,
		})
	}

	if len(section.ExamCode) > 0 {
		metadata = append(metadata, &uct.Metadata{
			Title:   "Exam Code",
			Content: getExamCode(section.ExamCode),
		})
	}

	if len(section.SpecialPermissionAddCodeDescription) > 0 {
		metadata = append(metadata, &uct.Metadata{
			Title:   "Special Permission",
			Content: "Code: " + section.SpecialPermissionAddCode + "\n" + section.SpecialPermissionAddCodeDescription,
		})
	}

	if len(section.Subtitle) > 0 {
		metadata = append(metadata, &uct.Metadata{
			Title:   "Subtitle",
			Content: section.Subtitle,
		})
	}

	return
}

type sectionSorter struct {
	sections []RSection
}

func (a sectionSorter) Len() int {
	return len(a.sections)
}

func (a sectionSorter) Swap(i, j int) {
	a.sections[i], a.sections[j] = a.sections[j], a.sections[i]
}

func (a sectionSorter) Less(i, j int) bool {
	return a.sections[i].Index < a.sections[j].Index
}

type courseSorter struct {
	courses []RCourse
}

func (a courseSorter) Len() int {
	return len(a.courses)
}

func (a courseSorter) Swap(i, j int) {
	a.courses[i], a.courses[j] = a.courses[j], a.courses[i]
}

func (a courseSorter) Less(i, j int) bool {
	c1 := a.courses[i]
	c2 := a.courses[j]
	var hash = func(s []RSection) string {
		var buffer bytes.Buffer
		for i := range s {
			buffer.WriteString(s[i].Index)
			buffer.WriteString(s[i].SectionNotes)
			buffer.WriteString(s[i].Subtitle)
		}
		return buffer.String()
	}
	return (c1.Title + c1.CourseNumber + hash(c1.Sections) + strconv.Itoa(int(c1.Credits))) < (c2.Title + c2.CourseNumber + hash(c2.Sections) + strconv.Itoa(int(c2.Credits)))
}

type commentSorter struct {
	comments []RComment
}

func (a commentSorter) Len() int {
	return len(a.comments)
}

func (a commentSorter) Swap(i, j int) {
	a.comments[i], a.comments[j] = a.comments[j], a.comments[i]
}

func (a commentSorter) Less(i, j int) bool {
	return a.comments[i].Code < a.comments[j].Code
}

func (meeting MeetingByClass) Len() int {
	return len(meeting)
}

func (meeting MeetingByClass) Swap(i, j int) {
	meeting[i], meeting[j] = meeting[j], meeting[i]
}

func (meeting MeetingByClass) Less(i, j int) bool {
	if meeting[i].isByArrangement() {
		return false
	}
	if meeting[j].isByArrangement() {
		return true
	}
	if meeting[i].isRecitation() {
		return false
	}
	if meeting[j].isRecitation() {
		return true
	}
	day1 := meeting[i].dayRank()
	day2 := meeting[j].dayRank()

	if day1 <= day2 {
		return true
	}
	return IsAfter(meeting[i].StartTime, meeting[j].StartTime)
}

func (meeting RMeetingTime) classRank() int {
	if meeting.isLecture() {
		return 1
	} else if meeting.isStudio() {
		return 2
	} else if meeting.isRecitation() {
		return 3
	} else if meeting.isByArrangement() {
		return 4
	} else if meeting.isLab() {
		return 5
	}
	return 99
}

func (meeting RMeetingTime) dayRank() int {
	switch meeting.MeetingDay {
	case "Monday":
		return 1
	case "Tuesday":
		return 2
	case "Wednesday":
		return 3
	case "Thurdsday":
		return 4
	case "Friday":
		return 5
	case "Saturday":
		return 6
	case "Sunday":
		return 7
	}
	return 8
}

func (meeting RMeetingTime) room() *string {
	if meeting.BuildingCode != "" {
		room := meeting.BuildingCode + "-" + meeting.RoomNumber
		return &room
	}
	return nil
}

func (meetingTime RMeetingTime) getMeetingHourBegin() string {
	if len(meetingTime.StartTime) > 1 || len(meetingTime.EndTime) > 1 {

		meridian := ""

		if meetingTime.PmCode != "" {
			if meetingTime.PmCode == "A" {
				meridian = "AM"
			} else {
				meridian = "PM"
			}
		}
		return formatMeetingHours(meetingTime.StartTime) + " " + meridian
	}
	return ""
}

func (meetingTime RMeetingTime) getMeetingHourEnd() string {
	if len(meetingTime.StartTime) > 1 || len(meetingTime.EndTime) > 1 {
		var meridian string
		starttime := meetingTime.StartTime
		endtime := meetingTime.EndTime
		pmcode := meetingTime.PmCode

		end, _ := strconv.Atoi(endtime[:2])
		start, _ := strconv.Atoi(starttime[:2])

		if pmcode != "A" {
			meridian = "PM"
		} else if end < start {
			meridian = "PM"
		} else if endtime[:2] == "12" {
			meridian = "PM"
		} else {
			meridian = "AM"
		}

		return formatMeetingHours(meetingTime.EndTime) + " " + meridian
	}
	return ""
}

func (meetingTime RMeetingTime) getMeetingHourBeginTime() time.Time {
	if len(uct.TrimAll(meetingTime.StartTime)) > 1 || len(uct.TrimAll(meetingTime.EndTime)) > 1 {

		meridian := ""

		if meetingTime.PmCode != "" {
			if meetingTime.PmCode == "A" {
				meridian = "AM"
			} else {
				meridian = "PM"
			}
		}

		kitchenTime := uct.TrimAll(formatMeetingHours(meetingTime.StartTime) + meridian)
		time, err := time.Parse(time.Kitchen, kitchenTime)
		uct.CheckError(err)
		return time
	}
	return time.Unix(0, 0)
}

func (meeting RMeetingTime) metadata() (metadata []*uct.Metadata) {

	return
}

func (course RCourse) synopsis() *string {
	if course.CourseDescription == "" {
		return nil
	} else {
		return &course.CourseDescription
	}
}

func (course RCourse) metadata() (metadata []*uct.Metadata) {

	if course.UnitNotes != "" {
		metadata = append(metadata, &uct.Metadata{
			Title:   "School Notes",
			Content: course.UnitNotes,
		})
	}

	if course.SubjectNotes != "" {
		metadata = append(metadata, &uct.Metadata{
			Title:   "Subject Notes",
			Content: course.SubjectNotes,
		})
	}
	if course.PreReqNotes != "" {
		metadata = append(metadata, &uct.Metadata{
			Title:   "Prequisites",
			Content: course.PreReqNotes,
		})
	}
	if course.SynopsisURL != "" {
		metadata = append(metadata, &uct.Metadata{
			Title:   "Synopsis Url",
			Content: course.SynopsisURL,
		})
	}

	return metadata
}

func FilterSubjects(vs []RSubject, f func(RSubject) bool) []RSubject {
	vsf := make([]RSubject, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
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

type (
	MeetingByClass []RMeetingTime

	RSubject struct {
		Name    string    `json:"description,omitempty"`
		Number  string    `json:"code,omitempty"`
		Courses []RCourse `json:"courses,omitempty"`
		Season  string
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
		CampusLocation  string  `json:"campusLocation"`
		BaClassHours    string  `json:"baClassHours"`
		RoomNumber      string  `json:"roomNumber"`
		PmCode          string  `json:"pmCode"`
		CampusAbbrev    string  `json:"campusAbbrev"`
		CampusName      string  `json:"campusName"`
		MeetingDay      string  `json:"meetingDay"`
		BuildingCode    string  `json:"buildingCode"`
		StartTime       string  `json:"startTime"`
		EndTime         string  `json:"endTime"`
		PStartTime      *string `json:"-"`
		PEndTime        *string `json:"-"`
		MeetingModeDesc string  `json:"meetingModeDesc"`
		MeetingModeCode string  `json:"meetingModeCode"`
	}
)
