package common

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/jmoiron/sqlx"
	"github.com/pquerna/ffjson/ffjson"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var trim = strings.TrimSpace

func CheckError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func ToNullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: true}
}

func Atoi64(str string) sql.NullInt64 {
	if val, err := strconv.Atoi(str); err == nil {
		return ToNullInt64(int64(val))
	}
	return sql.NullInt64{Int64: 0, Valid: true}
}

func ToFloat64(str string) float64 {
	s, err := strconv.ParseFloat(str, 64)
	CheckError(err)
	return s
}

func FloatToString(format string, num float64) string {
	return fmt.Sprintf(format, num)
}

func Log(v ...interface{}) {
	s := fmt.Sprint(v...)
	log.Printf("%s[%s] %s\n", os.Args[0], strconv.Itoa(os.Getpid()), s)
}

func LogVerbose(v interface{}) {
	log.Printf("%s[%s] %#v\n", os.Args[0], strconv.Itoa(os.Getpid()), v)
}

func TrimAll(str string) string {
	regex, err := regexp.Compile("\\s+")
	CheckError(err)
	str = regex.ReplaceAllString(str, " ")

	// Remove NUL and Heading bytes from string
	str = string(bytes.Replace([]byte(str), []byte("\x00"), []byte(""), -1))
	str = string(bytes.Replace([]byte(str), []byte("\x01"), []byte(""), -1))
	return trim(str)
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s\n", name, elapsed)
}

// stack returns a nicely formated stack frame, skipping skip frames
func Stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}

func StartPprof(host *net.TCPAddr) {
	log.Println("**Starting debug server on...", (*host).String())
	log.Println(http.ListenAndServe((*host).String(), nil))
}

func InitDB(connection string) *sqlx.DB {
	database, err := sqlx.Open("postgres", connection)
	if err != nil {
		log.Fatalln(err)
	}
	return database
}

func InitTnfluxServer() client.Client {
	influxClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     INFLUX_HOST,
		Username: INFLUX_USER,
		Password: INFLUX_PASS,
	})
	CheckError(err)
	return influxClient
}

func MarshalMessage(format string, m University) *bytes.Reader {
	var out []byte
	var err error
	if format == "json" {
		out, err = json.MarshalIndent(m, "", "   ")
		if err != nil {
			log.Fatalln("Failed to encode message:", err)
		}
	} else if format == "protobuf" {
		out, err = m.Marshal()
		if err != nil {
			log.Fatalln("Failed to encode message:", err)
		}
	}
	return bytes.NewReader(out)
}

func UnmarshallMessage(format string, r io.Reader, m *University) error {
	if format == "json" {
		dec := ffjson.NewDecoder()
		if err := dec.DecodeReader(r, &*m); err != nil {
			log.Fatalln("Failed to unmarshal message:", err)
		}
	} else if format == "protobuf" {
		data, err := ioutil.ReadAll(r)
		if err = m.Unmarshal(data); err != nil {
			log.Fatalln("Failed to unmarshal message:", err)
		}
	}
	if m.Equal(University{}) {
		return fmt.Errorf("%s Reason %s", "Failed to unmarshal message:", "empty data")
	}
	return nil
}

func CheckUniqueSubject(subjects []*Subject) {
	m := make(map[string]int)
	for subjectIndex := range subjects {
		subject := subjects[subjectIndex]
		key := subject.Season + subject.Year + subject.Name
		m[key]++
		if m[key] > 1 {
			Log("Duplicate subject found:", key, " c:", m[key])
			subject.Name = subject.Name + "_" + strconv.Itoa(m[key])
		}
	}
}

func CheckUniqueCourse(subject *Subject, courses []*Course) {
	m := map[string]int{}
	for courseIndex := range courses {
		course := courses[courseIndex]
		key := course.Name + course.Number
		m[key]++
		if m[key] > 1 {
			Log("Subject ", subject.Name, " in ", subject.Season)
			Log("Duplicate course found: ", key, " c:", m[key])
			course.Name = course.Name + "_" + strconv.Itoa(m[key])
		}
	}
}

func ValidateAll(uni *University) {
	uni.Validate()
	CheckUniqueSubject(uni.Subjects)
	for subjectIndex := range uni.Subjects {
		subject := uni.Subjects[subjectIndex]
		subject.Validate(uni)

		courses := subject.Courses
		CheckUniqueCourse(subject, courses)
		for courseIndex := range courses {
			course := courses[courseIndex]
			course.Validate(subject)

			sections := course.Sections
			for sectionIndex := range sections {
				section := sections[sectionIndex]
				section.Validate(course)

				//[]Instructors
				instructors := section.Instructors
				for instructorIndex := range instructors {
					instructor := instructors[instructorIndex]
					instructor.Index = int32(instructorIndex)
					instructor.Validate()
				}

				//[]Meeting
				meetings := section.Meetings
				for meetingIndex := range meetings {
					meeting := meetings[meetingIndex]
					meeting.Index = int32(meetingIndex)
					meeting.Validate()

					// Meeting []Metadata
					metadatas := meeting.Metadata
					for metadataIndex := range metadatas {
						metadata := metadatas[metadataIndex]
						metadata.Validate()
					}
				}

				//[]Books
				books := section.Books
				for bookIndex := range books {
					book := books[bookIndex]
					book.Validate()
				}

				// Section []Metadata
				metadatas := section.Metadata
				for metadataIndex := range metadatas {
					metadata := metadatas[metadataIndex]
					metadata.Validate()
				}
			}

			// Course []Metadata
			metadatas := course.Metadata
			for metadataIndex := range metadatas {
				metadata := metadatas[metadataIndex]
				metadata.Validate()
			}
		}
	}

	for registrations := range uni.Registrations {
		_ = uni.Registrations[registrations]

	}

	// university []Metadata
	metadatas := uni.Metadata
	for metadataIndex := range metadatas {
		metadata := metadatas[metadataIndex]
		metadata.Validate()

	}
}
