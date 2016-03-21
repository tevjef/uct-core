package common

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
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
