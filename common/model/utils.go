package model

import (
	"bytes"
	"database/sql"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/pkg/errors"
	"uct/common/conf"
)

var trim = strings.TrimSpace

func init() {
	log.SetLevel(log.DebugLevel)
}

func CheckError(err error) {
	if err != nil {
		if conf.IsDebug() {
			log.Panicf("%+v\n", errors.Wrap(err, ""))
		}
		log.Fatalf("%+v\n", errors.Wrap(err, ""))
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

	// Remove NUL and Heading bytes from string, cannot be inserted into postgresql
	temp = bytes.Replace(temp, nullByte, emptyByteArray, -1)
	str = string(bytes.Replace(temp, headingBytes, emptyByteArray, -1))

	return trim(str)
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.WithFields(log.Fields{"elapsed": elapsed, "name": name}).Debug("")
}

func TimeTrackWithLog(start time.Time, logger *log.Logger, name string) {
	elapsed := time.Since(start)
	logger.WithFields(log.Fields{"elapsed": elapsed.Seconds()*1e3, "name": name}).Info()
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
	log.Info("Starting debug server on...", (*host).String())
	log.Info(http.ListenAndServe((*host).String(), nil))
}

func InitDB(connection string) (database *sqlx.DB, err error) {
	database, err = sqlx.Open("postgres", connection)
	if err != nil {
		err = fmt.Errorf("Failed to open postgres databse connection. %s", err)
	}
	return
}