package main

import (
	"bufio"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"regexp"
	"strings"
	"sync"
	"time"
	"strconv"
	uct "uct/common"
	"os"
	"github.com/pquerna/ffjson/ffjson"
"net/http"
"log"
)

func main() {
	enc := ffjson.NewEncoder(os.Stdout)

	var schools []uct.University
	schools = append(schools, getUniversity())
	err := enc.Encode(schools)
	uct.CheckError(err)

}


func getUniversity() (university uct.University) {
	university = uct.University{
		Name:             "New Jersey Institute of Technology",
		Abbr:             "NJIT",
		MainColor:        "F44336",
		AccentColor:      "607D8B",
		HomePage:         "http://njit.edu/",
		RegistrationPage: "https://www.njit.edu/cp/login.php",
		Registrations: []uct.Registration{
			uct.Registration{
				Period:     uct.SEM_FALL.String(),
				PeriodDate: time.Date(2000, time.September, 6, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.SEM_SPRING.String(),
				PeriodDate: time.Date(2000, time.January, 17, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.SEM_SUMMER.String(),
				PeriodDate: time.Date(2000, time.May, 30, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.SEM_WINTER.String(),
				PeriodDate: time.Date(2000, time.December, 23, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.START_FALL.String(),
				PeriodDate: time.Date(2000, time.March, 20, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.START_SPRING.String(),
				PeriodDate: time.Date(2000, time.October, 18, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.START_SUMMER.String(),
				PeriodDate: time.Date(2000, time.January, 14, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.START_WINTER.String(),
				PeriodDate: time.Date(2000, time.September, 21, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.END_FALL.String(),
				PeriodDate: time.Date(2000, time.September, 13, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.END_SPRING.String(),
				PeriodDate: time.Date(2000, time.January, 27, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.END_SUMMER.String(),
				PeriodDate: time.Date(2000, time.August, 15, 0, 0, 0, 0, time.UTC),
			},
			uct.Registration{
				Period:     uct.END_WINTER.String(),
				PeriodDate: time.Date(2000, time.December, 22, 0, 0, 0, 0, time.UTC),
			},
		},
		Metadata: []uct.Metadata{
			uct.Metadata{
				Title: "About", Content: ``,
			},
		},
	}

	res := uct.ResolveSemesters(time.Now(), university.Registrations)
	Semesters := [3]uct.Semester{res.Last, res.Current, res.Next}
	for _, ThisSemester := range Semesters {
		if ThisSemester.Season == uct.WINTER {
			ThisSemester.Year += 1
		}
		subjects := getSubjectList(uctToNjitSeason(ThisSemester))
		var wg sync.WaitGroup
		wg.Add(len(subjects))
		for i, _ := range subjects {
			go func(sub *NSubject) {
				courses := getCourses(*sub)

				for j, _ := range courses {
					sub.Courses = append(sub.Courses, courses[j])
				}

				sub.Courses =courses
				wg.Done()
			}(&subjects[i])

		}
		wg.Wait()
		for _, subject := range subjects {
			newSubject := uct.Subject{
				Name:   subject.SubjectName,
				Number: subject.SubjectId,
				Season: subject.Semester.Season.FullString(),
				Year:   strconv.Itoa(subject.Semester.Year)}
			for _, course := range subject.Courses {
				newCourse := uct.Course{
					Name:     course.CourseName,
					Number:   course.CourseNum,
					Synopsis: &course.CourseDescription,}

				for _, section := range course.Section {
					newSection := uct.Section{
						Number:     section.SectionNum,
						CallNumber: section.CallNum,
						Status:     section.Status,
						Credits:    section.Credits,
						Max:        section.Max,
						Now:        section.Now, }

					if section.Instructor != "" {
						newInstructor := uct.Instructor{Name: section.Instructor}

						newInstructor.VetAndBuild()
						newSection.Instructors = append(newSection.Instructors, newInstructor)


					}
					
					for _, meeting := range section.MeetingTimes {
						newMeeting := uct.Meeting{
							Room:      &meeting.Room,
							Day:       &meeting.Day,
							StartTime: meeting.StartTime,
							EndTime:   meeting.EndTime,
						}

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
			university.Subjects = append(university.Subjects, newSubject)
		}
	}
	university.VetAndBuild()
	return
}

func getSubjectList(semester NSemester) []NSubject {
	log.Println("GETTING SUBJECT LIST")
	url := fmt.Sprintf("https://courseschedules.njit.edu/index.aspx?semester=%s%s",
		strconv.Itoa(semester.Year), semester.Season.String())
	log.Println("Getting: ",url)

	doc, err := goquery.NewDocument(url)
	checkError(err)

	return extractSubjectList(doc, semester)
}

func extractSubjectList(doc *goquery.Document, semester NSemester) (subjectList []NSubject) {
	doc.Find(".dashed_wrapper a").Each(func(i int, s *goquery.Selection) {

		num := trim(substringBefore(s.Text(), "-"))
		name := trim(substringAfter(s.Text(), "-"))
		if num[:1] == "R" && name == "" {
			name = "Offered by Rutgers"
		} else if num == "MIT" {
			name = "Offered by Electrical Tech"
		} else if name == "" {
			name = "UNKNOWN NAME"
		}

		subject := NSubject{
			SubjectId:  num,
			SubjectName:name,
			Semester:semester}
		log.Printf("FOUND SUBJECT in Semester %#v :: %#v", semester, subject)
		subjectList = append(subjectList, subject)
	})
	return
}


func getCourses(subject NSubject) (courses []NCourse) {
	var url = fmt.Sprintf("https://courseschedules.njit.edu/index.aspx?semester=%s%s&subjectID=%s",
	strconv.Itoa(subject.Semester.Year), subject.Semester.Season.String(), subject.SubjectId)
	log.Println("Url: ", url)
	doc, err := goquery.NewDocument(url)
	checkError(err)

	return extractCourseList(doc)
}

func extractCourseList(doc *goquery.Document) (courses []NCourse) {
	doc.Find(".subject_wrapper").Each(func(i int, s *goquery.Selection) {
		course := NCourse{
			CourseNum:         extractCourseNum(s),
			CourseName:        extractCourseName(s),
			CourseDescription: extractCourseDescription(s),
			Section:           getSections(s),
		}
		log.Printf("FOUND Course in document :: %#v", course)

		if course.CourseNum != "" {
			courses = append(courses, course)
		}

	})
	return
}

func getSections(s *goquery.Selection) (sections []NSection) {
	s.Find(".sectionRow").Each(func(i int, s *goquery.Selection) {
		section := NSection{
			SectionNum:   extractSectionNum(s),
			CallNum:      extractCallNum(s),
			MeetingTimes: extractTimes(s),
			Status:       extractStatus(s),
			Max:          extractMaxSize(s),
			Now:          extractNowSize(s),
			Instructor:   extractInstructor(s),
			BookUrl:      extractBookUrl(s),
			Credits:      extractCredits(s),
		}

		log.Printf("FOUND Section in document :: %#v", section)

		sections = append(sections, section)
	})
	return
}

func extractCourseName(selection *goquery.Selection) string {
	return trim(substringAfter(selection.Find(".catalogdescription").Text(), "-"))
}

func extractCourseNum(selection *goquery.Selection) string {
	return trim(substringAfterLast(trim(substringBefore(selection.Find(".catalogdescription").Text(), "-")), " "))
}

func extractCourseDescription(selection *goquery.Selection) string {
	url := trim(fmt.Sprintln(selection.Find(".catalogdescription a").AttrOr("href", "")))
	log.Println("LOGGING URL", url)
	client := http.Client{}
	req, _ := http.NewRequest("GET", "http://catalog.njit.edu/ribbit/index.cgi?format=html&page=fsinjector.rjs&fullpage=true", nil)
	req.Header.Add("Referer",url)
	resp, err := client.Do(req)
	if err != nil {

		return ""
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	body, _ := ioutil.ReadAll(resp.Body)
	//checkError(err)
	result := substringAfter(string(body), "courseblockdesc")
	if len(result) < 4 {
		return ""
	}
	result = substringBefore(result[3:], "<b")
	if string(result[0]) == "<" || strings.Contains(result, "at SISConnxService"){
		return ""
	}
	result = strings.Replace(result, "\\\"", "\"", -1)
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(result))

	return trim(doc.Text())
}

func extractSectionNum(selection *goquery.Selection) string {
	return trim(selection.Find(".section").Text())
}

func extractCallNum(selection *goquery.Selection) string {
	return trim(selection.Find(".call span").Text())
}

func extractBookUrl(selection *goquery.Selection) string {
	return strings.Replace(trim(selection.Find(".call a").AttrOr("href", "")), " ", "%20", -1)
}

func extractRoomNum(selection *goquery.Selection) string {
	s, _ := selection.Find(".room").Html()
	s = strings.Replace(s, "<br/>", "\n", -1)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err != nil {
		log.Print(err)
	}
	return trim(doc.Text())
}

func extractTimes(selection *goquery.Selection) (meetingTimes []NMeetingTime) {
	s, _ := selection.Find(".call").Next().Html()
	s = strings.Replace(s, "<br/>", "\n", -1)
	s = strings.Replace(s, "\t", "", -1)
	regex, _ := regexp.Compile(`\s\s+`)
	s = regex.ReplaceAllString(s, "")
	s = substringBefore(s, "Sec")
	rawroom := extractRoomNum(selection)
	rooms := strings.Split(rawroom, "\n")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err != nil {
		log.Print(err)
	}
	result := trim(doc.Text())
	scanner := bufio.NewScanner(strings.NewReader(result))
	j := 0
	for scanner.Scan() {
		var room string
		if len(rooms) > j {
			room = rooms[j]
		}
		j++
		text := scanner.Text()
		i := 0
		for {
			day := text[i]
			if day == 'M' || day == 'T' || day == 'W' || day == 'R' || day == 'F' || day == 'S' {
				start, end := getStartAndEnd(text)
				meetingTimes = append(meetingTimes, NMeetingTime{
					Day:       getDay(string(day)),
					StartTime: start,
					EndTime:   end,
					Room:      room,
				})
				getStartAndEnd(text)
			} else {
				break
			}
			i++
		}

	}
	return
}

func getDay(day string) string {
	switch day {
	case "M":
		return "Monday"
	case "T":
		return "Tuesday"
	case "W":
		return "Wednesday"
	case "R":
		return "Thursday"
	case "F":
		return "Friday"
	case "S":
		return "Saturday"
	}

	return ""
}

func getStartAndEnd(time string) (string, string) {
	r, _ := regexp.Compile("\\d{3,4}(AM|PM)")
	result := r.FindAllString(time, -1)
	if result != nil {
		return formatTime(result[0]), formatTime(result[1])
	}
	return "", ""
}

func formatTime(time string) string {
	if len(time) == 6 {
		return time[:2] + ":" + time[2:]
	}
	if len(time) == 5 {
		return time[:1] + ":" + time[1:]
	}
	return time
}

func extractStatus(selection *goquery.Selection) string {
	return trim(selection.Find(".status").Text())
}

func extractMaxSize(selection *goquery.Selection) float64 {
	return ToFloat64(selection.Find(".max").Text())
}

func extractNowSize(selection *goquery.Selection) float64 {
	return ToFloat64(trim(selection.Find(".now").Text()))
}

func extractInstructor(selection *goquery.Selection) string {
	return trim(selection.Find(".instructor").Text())
}

func extractCredits(selection *goquery.Selection) string {
	if result :=trim (selection.Find(".credits").Text()); strings.Contains(result, "#") {
		return "0"
	} else {
		return result
	}
}

type (
	NSubject struct {
		SubjectId   string `json:"sid,omitempty"`
		SubjectName string `json:"name,omitempty"`
		Semester    NSemester
		Courses     []NCourse `json:"sections,omitempty"`

	}

	NCourse struct {
		CourseName        string    `json:"name,omitempty"`
		CourseNum         string    `json:"number,omitempty"`
		CourseDescription string    `json:"description,omitempty"`
		Section           []NSection `json:"sections,omitempty"`
	}

	NSection struct {
		SectionNum   string        `json:"section_number,omitempty"`
		CallNum      string        `json:"call_number,omitempty"`
		MeetingTimes []NMeetingTime `json:"meeting_time,omitempty"`
		Status       string        `json:"status,omitempty"`
		Max          float64        `json:"max,omitempty"`
		Now          float64        `json:"now,omitempty"`
		Instructor   string        `json:"instructor,omitempty"`
		BookUrl      string        `json:"book_url,omitempty"`
		Credits      string        `json:"credits,omitempty"`
	}

	NMeetingTime struct {
		Day       string `json:"day,omitempty"`
		StartTime string `json:"start_time,omitempty"`
		EndTime   string `json:"end_time,omitempty"`
		Room      string `json:"room,omitempty"`
	}

	NSeason int

	NSemester struct {
		Year   int
		Season NSeason
	}

	ResolvedSemester struct {
		Last    NSemester
		Current NSemester
		Next    NSemester
	}
)

const (
	FALL NSeason = iota
	SPRING
	SUMMER
	WINTER
)

var seasons = [...]string{
	"f",
	"s",
	"u",
	"w",
}

var seasonsFull = [...]string{
	"fall",
	"spring",
	"summer",
	"winter",
}

func (s NSemester) String() string {
	if (s.Season == FALL) {
		return "Sep-01-"+ strconv.Itoa(s.Year)
	} else if (s.Season == WINTER) {
		return "Dec-01-"+ strconv.Itoa(s.Year)
	} else if (s.Season == SPRING) {
		return "Feb-01-"+ strconv.Itoa(s.Year)
	} else {
		return "Jun-01-"+ strconv.Itoa(s.Year)
	}
}

func (s NSeason) String() string {
	return seasons[s]
}

func (s NSeason) FullString() string {
	return seasonsFull[s]
}

func uctToNjitSeason(sem uct.Semester) NSemester {
	var season NSeason
	switch sem.Season {
	case uct.FALL:
		season = FALL
	case uct.SPRING:
		season = SPRING
	case uct.SUMMER:
		season = SUMMER
	case uct.WINTER:
		season = WINTER
	}

	return NSemester{Year:sem.Year, Season:season}
}