package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pquerna/ffjson/ffjson"

	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"testing"
	uct "uct/common"
)

var (
	db     *sqlx.DB
	course = []byte(`[{"name":"Colloquium In Art, Culture, \u0026 Media","number":"301","synopsis":{"String":"","Valid":false},"hash":"51976a63b95239f1a7f887c7351c6353cae3020d","topic_name":"Colloquium.In.Art..Culture....Media","sections":[{"number":"01","call_number":"11392","status":"Closed","credits":"3.0","meeting":[{"room":{"String":"BRD - 312","Valid":true},"day":{"String":"Friday","Valid":true},"start_time":"2:30 PM","end_time":"5:20 PM","metadata":[{"title":"Type","content":"LEC"}]}],"instructors":[{"name":"GILBERT"}],"metadata":[{"title":"Exam Code","content":"C"},{"title":"Campus Code","content":"NK"}]},{"number":"02","call_number":"11393","status":"Closed","credits":"3.0","meeting":[{"room":{"String":"BRD - 313","Valid":true},"day":{"String":"Friday","Valid":true},"start_time":"2:30 PM","end_time":"5:20 PM","metadata":[{"title":"Type","content":"LEC"}]}],"instructors":[{"name":"ENGLOT"}],"metadata":[{"title":"Exam Code","content":"C"},{"title":"Campus Code","content":"NK"}]},{"number":"03","call_number":"16163","status":"Closed","credits":"3.0","meeting":[{"room":{"String":"BRD - 207","Valid":true},"day":{"String":"Thursday","Valid":true},"start_time":"2:30 PM","end_time":"5:20 PM","metadata":[{"title":"Type","content":"LEC"}]}],"instructors":[{"name":"WATSON"}],"metadata":[{"title":"Exam Code","content":"C"},{"title":"Special Permission","content":"Code: 04\nInstructor"},{"title":"Campus Code","content":"NK"}]}],"metadata":[{"title":"Subject Notes","content":"21:083:301:01 (Colloquium in Arts, Culture \u0026 Media) is required for all students declaring a major in the Department of Arts, Culture \u0026 Media after January 1, 2010. Students majoring in departments other than Arts, Culture \u0026 Media should contact the course professor for permission to enroll."},{"title":"Prequisites","content":"(21:083:101 )"}]},
	{"name":"Tevin Colloquium In Art, Culture, \u0026 Media","number":"301","synopsis":{"String":"","Valid":false},"hash":"51976a63b95239f1a7f887c7351c6353cae3020d","topic_name":"Colloquium.In.Art..Culture....Media","sections":[{"number":"01","call_number":"11392","status":"Closed","credits":"3.0","meeting":[{"room":{"String":"BRD - 312","Valid":true},"day":{"String":"Friday","Valid":true},"start_time":"2:30 PM","end_time":"5:20 PM","metadata":[{"title":"Type","content":"LEC"}]}],"instructors":[{"name":"GILBERT"}],"metadata":[{"title":"Exam Code","content":"C"},{"title":"Campus Code","content":"NK"}]},{"number":"02","call_number":"11393","status":"Closed","credits":"3.0","meeting":[{"room":{"String":"BRD - 313","Valid":true},"day":{"String":"Friday","Valid":true},"start_time":"2:30 PM","end_time":"5:20 PM","metadata":[{"title":"Type","content":"LEC"}]}],"instructors":[{"name":"ENGLOT"}],"metadata":[{"title":"Exam Code","content":"C"},{"title":"Campus Code","content":"NK"}]},{"number":"03","call_number":"16163","status":"Closed","credits":"3.0","meeting":[{"room":{"String":"BRD - 207","Valid":true},"day":{"String":"Thursday","Valid":true},"start_time":"2:30 PM","end_time":"5:20 PM","metadata":[{"title":"Type","content":"LEC"}]}],"instructors":[{"name":"WATSON"}],"metadata":[{"title":"Exam Code","content":"C"},{"title":"Special Permission","content":"Code: 04\nInstructor"},{"title":"Campus Code","content":"NK"}]}],"metadata":[{"title":"Subject Notes","content":"21:083:301:01 (Colloquium in Arts, Culture \u0026 Media) is required for all students declaring a major in the Department of Arts, Culture \u0026 Media after January 1, 2010. Students majoring in departments other than Arts, Culture \u0026 Media should contact the course professor for permission to enroll."},{"title":"Prequisites","content":"(21:083:101 )"}]},
	{"name":"Jeffrey Colloquium In Art, Culture, \u0026 Media","number":"301","synopsis":{"String":"","Valid":false},"hash":"51976a63b95239f1a7f887c7351c6353cae3020d","topic_name":"Colloquium.In.Art..Culture....Media","sections":[{"number":"01","call_number":"11392","status":"Closed","credits":"3.0","meeting":[{"room":{"String":"BRD - 312","Valid":true},"day":{"String":"Friday","Valid":true},"start_time":"2:30 PM","end_time":"5:20 PM","metadata":[{"title":"Type","content":"LEC"}]}],"instructors":[{"name":"GILBERT"}],"metadata":[{"title":"Exam Code","content":"C"},{"title":"Campus Code","content":"NK"}]},{"number":"02","call_number":"11393","status":"Closed","credits":"3.0","meeting":[{"room":{"String":"BRD - 313","Valid":true},"day":{"String":"Friday","Valid":true},"start_time":"2:30 PM","end_time":"5:20 PM","metadata":[{"title":"Type","content":"LEC"}]}],"instructors":[{"name":"ENGLOT"}],"metadata":[{"title":"Exam Code","content":"C"},{"title":"Campus Code","content":"NK"}]},{"number":"03","call_number":"16163","status":"Closed","credits":"3.0","meeting":[{"room":{"String":"BRD - 207","Valid":true},"day":{"String":"Thursday","Valid":true},"start_time":"2:30 PM","end_time":"5:20 PM","metadata":[{"title":"Type","content":"LEC"}]}],"instructors":[{"name":"WATSON"}],"metadata":[{"title":"Exam Code","content":"C"},{"title":"Special Permission","content":"Code: 04\nInstructor"},{"title":"Campus Code","content":"NK"}]}],"metadata":[{"title":"Subject Notes","content":"21:083:301:01 (Colloquium in Arts, Culture \u0026 Media) is required for all students declaring a major in the Department of Arts, Culture \u0026 Media after January 1, 2010. Students majoring in departments other than Arts, Culture \u0026 Media should contact the course professor for permission to enroll."},{"title":"Prequisites","content":"(21:083:101 )"}]}]`)
)

func setup() {
	var err error
	db, err = sqlx.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s:5432/%s",
			uct.DbUser, uct.DbPassword, uct.DbHost, uct.DbName))
	if err != nil {
		log.Fatalln(err)
	}
}

func TestInsertUniversity(t *testing.T) {
	setup()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	f, err := os.Open("C:\\Users\\Tevin\\Desktop\\Development\\Go\\src\\uct\\short.txt")
	uct.CheckError(err)

	/*	f := bytes.NewReader(course)
	 */

	dec := ffjson.NewDecoder()

	var universities []uct.University

	if err := dec.DecodeReader(f, &universities); err != nil {
		log.Panicln(err)
	}
	for _, university := range universities {
		fmt.Println(university)
	}
	//insertUniversity(db, university)

}

/*

func BenchmarkCourseInsert(b *testing.B) {
	setup()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	b.ReportAllocs()

	var newCourse uct.Course

	err := json.Unmarshal(course, &newCourse)
	uct.CheckError(err)

	for i := 0; i < b.N; i++ {
		insertCourse(database, newCourse)
	}
}*/
