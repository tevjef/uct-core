package common

import (
	"bytes"
	"database/sql"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var trim = strings.TrimSpace

func init() {
	log.SetLevel(log.DebugLevel)
}

func CheckError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func LogError(err error) {
	if err != nil {
		log.Errorln(err)
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
	log.Debugln(v...)
}

func LogVerbose(v interface{}) {
	log.Debugln(v)
}

var emptyByteArray = make([]byte, 0)

var nullByte = []byte("\x00")
var headingBytes = []byte("\x01")

func stripSpaces(str string) string {
	var lastRune rune
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) && unicode.IsSpace(lastRune) {
			// if the character is a space, drop it
			return -1
		}
		lastRune = r
		// else keep it in the string
		return r
	}, str)
}

func TrimAll(str string) string {
	str = stripSpaces(str)
	temp := []byte(str)

	// Remove NUL and Heading bytes from string
	temp = bytes.Replace(temp, nullByte, emptyByteArray, -1)
	str = string(bytes.Replace(temp, headingBytes, emptyByteArray, -1))

	return trim(str)
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.WithFields(log.Fields{"elapsed": elapsed, "name": name}).Debugln("Latency")
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
	log.Debug("Starting debug server on...", (*host).String())
	log.Debug(http.ListenAndServe((*host).String(), nil))
}

func InitDB(connection string) (database *sqlx.DB, err error) {
	database, err = sqlx.Open("postgres", connection)
	if err != nil {
		err = fmt.Errorf("Failed to open postgres databse connection. %s", err)
	}
	return
}