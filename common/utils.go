package common

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gogo/protobuf/proto"
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

const (
	Empty         = ""
	IndexNotFound = -1
)

var trim = strings.TrimSpace

func substringAfter(str, separator string) string {
	if isEmpty(str) {
		return str
	}
	pos := strings.Index(str, separator)
	if pos == IndexNotFound {
		return Empty
	}
	return str[pos+len(separator):]
}

func substringAfterLast(str, separator string) string {
	if isEmpty(str) {
		return str
	}
	if isEmpty(separator) {
		return Empty
	}
	pos := strings.LastIndex(str, separator)
	if pos == IndexNotFound || pos == len(str)-len(separator) {
		return Empty
	}
	return str[pos+len(separator):]
}

func substringBefore(str, separator string) string {
	if isEmpty(str) {
		return str
	}
	pos := strings.Index(str, separator)
	if pos == IndexNotFound {
		return str
	}
	return str[:pos]
}

func substringBeforeLast(str, separator string) string {
	if isEmpty(str) || isEmpty(separator) {
		return str
	}
	pos := strings.LastIndex(str, separator)
	if pos == IndexNotFound {
		return Empty
	}
	return str[:pos]
}

func CheckError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func isEmpty(str string) bool {
	return len(str) == 0
}

//ToNullString invalidates a sql.NullString if empty, validates if not empty
func ToNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func ToNullFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid: f != 0}
}

func ToNullBool(b bool) sql.NullBool {
	return sql.NullBool{Bool: b, Valid: true}
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

func getDummyDoc(filename string) *goquery.Document {
	file, _ := ioutil.ReadFile(filename)
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(file))
	return doc
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
	log.Printf("%s took %s", name, elapsed)
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
		out, err = proto.Marshal(&m)
		if err != nil {
			log.Fatalln("Failed to encode message:", err)
		}
	}
	return bytes.NewReader(out)
}

func UnmarshallMessage(format string, r io.Reader, m *University) {
	if format == "json" {
		dec := ffjson.NewDecoder()
		if err := dec.DecodeReader(r, &*m); err != nil {
			log.Fatalln("Failed to unmarshal message:", err)
		}
	} else if format == "protobuf" {
		data, err := ioutil.ReadAll(r)
		if err = proto.Unmarshal(data, &*m); err != nil {
			log.Fatalln("Failed to unmarshal message:", err)
		}
	}
}
