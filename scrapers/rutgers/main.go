package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"uct/common/model"
	"uct/common/conf"
	rutgers "uct/scrapers/rutgers/model"

	"uct/redis"
	"uct/redis/harmony"

	log "github.com/Sirupsen/logrus"
	"github.com/pquerna/ffjson/ffjson"
	"gopkg.in/alecthomas/kingpin.v2"
	"crypto/md5"
)

var (
	host = "http://sis.rutgers.edu/soc"
)

var (
	app            = kingpin.New("rutgers", "A web scraper that retrives course information for Rutgers University's servers.")
	campus         = app.Flag("campus", "Choose campus code. NB=New Brunswick, CM=Camden, NK=Newark").HintOptions("CM", "NK", "NB").Short('u').PlaceHolder("[CM, NK, NB]").Required().String()
	format         = app.Flag("format", "Choose output format").Short('f').HintOptions(model.PROTOBUF, model.JSON).PlaceHolder("[protobuf, json]").Required().String()
	daemonInterval = app.Flag("daemon", "Run as a daemon with a refesh interval").Duration()
	daemonFile     = app.Flag("daemon-dir", "If supplied the deamon will write files to this directory").ExistingDir()
	latest         = app.Flag("latest", "Only output the current and next semester").Short('l').Bool()
	configFile     = app.Flag("config", "configuration file for the application").Short('c').File()
	config         conf.Config
	redisWrapper   *redishelper.RedisWrapper
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
	campus := strings.ToLower(*campus)
	app.Name = app.Name + "-" + campus

	if *format != model.JSON && *format != model.PROTOBUF {
		log.Fatalln("Invalid format:", *format)
	}

	isDaemon := *daemonInterval > 0
	// Parse configuration file
	config = conf.OpenConfig(*configFile)
	config.AppName = app.Name

	// Start profiling
	go model.StartPprof(config.GetDebugSever(app.Name))

	// Channel to send scraped data on
	resultChan := make(chan model.University)

	// Runs at regular intervals
	if isDaemon {
		// Override cli arg with environment variable
		if intervalFromEnv := config.Scrapers.Get(app.Name).Interval; intervalFromEnv != "" {
			if interval, err := time.ParseDuration(intervalFromEnv); err != nil {
				model.CheckError(err)
			} else if interval > 0 {
				daemonInterval = &interval
			}
		}

		redisWrapper = redishelper.New(config, app.Name)

		harmony.DaemonScraper(redisWrapper, *daemonInterval, func(cancel chan bool) {
			entryPoint(resultChan)
		})

	} else {
		go func() {
			entryPoint(resultChan)
			close(resultChan)
		}()
	}

	// block as it waits for results to come in
	for school := range resultChan {
		reader := model.MarshalMessage(*format, school)

		// Write to redis
		if isDaemon {
			pushToRedis(reader)
			continue
		}

		// Write to file
		if *daemonFile != "" {
			if data, err := ioutil.ReadAll(reader); err != nil {
				model.CheckError(err)
			} else {
				fileName := *daemonFile + "/" + app.Name + "-" + strconv.FormatInt(time.Now().Unix(), 10) + "." + *format
				log.Debugln("Writing file", fileName)
				if err = ioutil.WriteFile(fileName, data, 0644); err != nil {
					model.CheckError(err)
				}
			}
			continue
		}

		// Write to stdout
		io.Copy(os.Stdout, reader)
	}
}

func pushToRedis(reader *bytes.Reader) {
	if data, err := ioutil.ReadAll(reader); err != nil {
		model.CheckError(err)
	} else {
		log.WithFields(log.Fields{"scraper_name": app.Name, "bytes": len(data), "hash": md5.New().Sum(data)[:63]}).Info()
		if err := redisWrapper.Client.Set(redisWrapper.NameSpace+":data:latest", data, 0).Err(); err != nil {
			log.Panicln(errors.New("failed to connect to redis server"))
		}

		if _, err := redisWrapper.LPushNotExist(redishelper.BaseNamespace+":queue", redisWrapper.NameSpace); err != nil {
			log.Panicln(errors.New("failed to queue univeristiy for upload"))
		}
	}
}

func entryPoint(result chan model.University) {
	var school model.University

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

func getCampus(campus string) model.University {
	var university model.University

	university = getRutgers(campus)

	// Rutgers servers go down for maintenance between 3 and 5 AM UTC every day.
	// Data scraped from this time would be inaccurate and may lead to unforeseen errors
	if currentHour := time.Now().UTC().Hour(); currentHour >= 7 && currentHour < 9 {
		return university
	}

	university.ResolvedSemesters = model.ResolveSemesters(time.Now(), university.Registrations)

	Semesters := []*model.Semester{
		university.ResolvedSemesters.Last,
		university.ResolvedSemesters.Current,
		university.ResolvedSemesters.Next}

	if *latest {
		Semesters = []*model.Semester{
			university.ResolvedSemesters.Current,
			university.ResolvedSemesters.Next}
	}

	for _, ThisSemester := range Semesters {
		if ThisSemester.Season == model.WINTER {
			ThisSemester.Year += 1
		}

		// Shadow copy variable
		ThisSemester := ThisSemester
		subjects := getSubjects(ThisSemester, campus)

		var wg sync.WaitGroup

		sem := make(chan int, 10)
		for i := range subjects {
			wg.Add(1)
			go func(sub *rutgers.RSubject) {
				defer func() {
					wg.Done()
				}()
				sem <- 1
				courses := getCourses(sub.Number, campus, ThisSemester)
				<-sem
				for j := range courses {
					sub.Courses = append(sub.Courses, courses[j])
				}

			}(&subjects[i])

		}
		wg.Wait()

		// Filter subjects that don't have a course
		subjects = rutgers.FilterSubjects(subjects, func(subject rutgers.RSubject) bool {
			return len(subject.Courses) > 0
		})

		for _, subject := range subjects {
			newSubject := &model.Subject{
				Name:   subject.Name,
				Number: subject.Number,
				Season: subject.Season,
				Year:   strconv.Itoa(subject.Year)}
			for _, course := range subject.Courses {
				newCourse := &model.Course{
					Name:     course.ExpandedTitle,
					Number:   course.CourseNumber,
					Synopsis: course.Synopsis(),
					Metadata: course.Metadata()}

				for _, section := range course.Sections {
					newSection := &model.Section{
						Number:     section.Number,
						CallNumber: section.Index,
						Status:     section.Status(),
						Credits:    model.FloatToString("%.1f", course.Credits),
						Max:        0,
						Metadata:   section.Metadata()}

					for _, instructor := range section.Instructor {
						newInstructor := &model.Instructor{Name: instructor.Name}

						newSection.Instructors = append(newSection.Instructors, newInstructor)
					}

					for _, meeting := range section.MeetingTimes {
						newMeeting := &model.Meeting{
							Room:      meeting.Room(),
							Day:       meeting.DayPointer(),
							StartTime: meeting.PStartTime,
							EndTime:   meeting.PEndTime,
							ClassType: meeting.ClassType(),
							Metadata:  meeting.Metadata()}

						newSection.Meetings = append(newSection.Meetings, newMeeting)
					}

					newCourse.Sections = append(newCourse.Sections, newSection)

				}
				newSubject.Courses = append(newSubject.Courses, newCourse)
			}
			university.Subjects = append(university.Subjects, newSubject)
		}
	}

	return university
}

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

func getSubjects(semester *model.Semester, campus string) (subjects []rutgers.RSubject) {
	var url = fmt.Sprintf("%s/subjects.json?semester=%s&campus=%s&level=U%sG", host, getRutgersSemester(semester), campus, "%2C")

	for i := 0; i < 3; i++ {
		log.WithFields(log.Fields{"season": semester.Season, "year": semester.Year, "campus": campus, "retry": i, "url": url}).Debug("Subject Request")
		resp, err := httpClient.Get(url)
		if err != nil {
			log.Errorf("Retrying %d after error: %s\n", i, err)
			continue
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err := ffjson.NewDecoder().Decode(data, &subjects); err != nil && err != io.EOF {
			log.Errorf("Retrying %d after error: %s\n", i, err)
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

func getCourses(subject, campus string, semester *model.Semester) (courses []rutgers.RCourse) {
	var url = fmt.Sprintf("%s/courses.json?subject=%s&semester=%s&campus=%s&level=U%sG", host, subject, getRutgersSemester(semester), campus, "%2C")
	for i := 0; i < 3; i++ {
		log.WithFields(log.Fields{"subject": subject, "season": semester.Season, "year": semester.Year, "campus": campus, "retry": i, "url": url}).Debug("Course Request")

		resp, err := httpClient.Get(url)
		if err != nil {
			log.Errorf("Retrying %d after error: %s\n", i, err)
			continue
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err := ffjson.NewDecoder().Decode(data, &courses); err != nil {
			resp.Body.Close()
			log.Errorf("Retrying %d after error: %s\n", i, err)
			continue
		}

		log.WithFields(log.Fields{"content-length": len(data), "subject": subject, "status": resp.Status, "season": semester.Season,
			"year": semester.Year, "campus": campus, "url": url}).Debugln("Course Response")

		resp.Body.Close()
		break
	}

	for i := range courses {
		courses[i].Clean()
		for j := range courses[i].Sections {
			courses[i].Sections[j].Clean()
		}

		sort.Sort(rutgers.SectionSorter{courses[i].Sections})
	}
	sort.Sort(rutgers.CourseSorter{courses})

	courses = rutgers.FilterCourses(courses, func(course rutgers.RCourse) bool {
		return len(course.Sections) > 0
	})

	return
}

func getRutgersSemester(semester *model.Semester) string {
	if semester.Season == model.FALL {
		return "9" + strconv.Itoa(int(semester.Year))
	} else if semester.Season == model.SUMMER {
		return "7" + strconv.Itoa(int(semester.Year))
	} else if semester.Season == model.SPRING {
		return "1" + strconv.Itoa(int(semester.Year))
	} else if semester.Season == model.WINTER {
		return "0" + strconv.Itoa(int(semester.Year))
	}
	return ""
}
