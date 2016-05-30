package main

import (
	"bufio"
	"flag"
	_ "github.com/lib/pq"
	"github.com/pquerna/ffjson/ffjson"
	"log"
	_ "net/http/pprof"
	"os"
	"testing"
	uct "uct/common"
)

func setup() {
	var err error

	if err != nil {
		log.Fatalln(err)
	}
}

var university uct.University
var testApp app

func TestMain(m *testing.M) {
	flag.Parse()

	file, err := os.Open("C:\\Users\\Tevin\\Desktop\\Development\\Go\\src\\uct\\scrapers\\rutgers\\json.out")
	uct.CheckError(err)

	input := bufio.NewReader(file)

	dec := ffjson.NewDecoder()
	if err := dec.DecodeReader(input, &university); err != nil {
		log.Fatal(err)
	}

	go audit()

	testApp = app{dbHandler: MockDatabaseHandler{}}

	os.Exit(m.Run())
}

func TestInsertUniversity(t *testing.T) {
	testApp.insertUniversity(university)
}

func BenchmarkInsertUniversity(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		testApp.insertUniversity(university)
	}
}

type MockDatabaseHandler struct {
}

func (dbHandler MockDatabaseHandler) insert(query string, data interface{}) (id int64) {
	return
}

func (dbHandler MockDatabaseHandler) update(query string, data interface{}) (id int64) {
	return
}

func (dbHandler MockDatabaseHandler) upsert(insertQuery, updateQuery string, data interface{}) (id int64) {
	return
}

func (dbHandler MockDatabaseHandler) exists(query string, data interface{}) (id int64) {
	return
}
