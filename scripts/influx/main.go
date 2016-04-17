package main

import (
	"github.com/influxdata/influxdb/client/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	uct "uct/common"
	"time"
	"fmt"
)

var (
	database     *sqlx.DB
	influxClient client.Client
	sem      = make(chan int, 50)
	connection = uct.GetUniversityDB()
)

func init() {
	database = initDB(connection)

	var err error
	influxClient, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     uct.INFLUX_HOST,
		Username: uct.INFLUX_USER,
		Password: uct.INFLUX_PASS,
	})
	uct.CheckError(err)
}

func initDB(connection string) *sqlx.DB {
	database, err := sqlx.Open("postgres", connection)
	if err != nil {
		log.Fatalln(err)
	}
	return database
}

type SectionsStatus struct {
	Count      int64  `db:"count"`
	Status     string `db:"status"`
	Subject    string `db:"subject"`
	Course     string `db:"course"`
	University string `db:"university"`
	Season     string `db:"season"`
	Year       string `db:"year"`
}

func main() {
	sections := getSectionStatus()

	time := time.Now()

	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "universityct",
		Precision: "s",
	})

	for _, val := range sections {
		val := val
		sem <- 1
		go func() {

			tags := map[string]string{
				"university": val.University,
				"subject":    val.Subject,
				"course":     val.Course,
				"season":   val.Season + val.Year,
				"status": val.Status,
			}

			fields := map[string]interface{}{
				"value":    val.Count,
			}

			point, err := client.NewPoint(
				"section_count",
				tags,
				fields,
				time,
			)

			uct.CheckError(err)

			bp.AddPoint(point)

			fmt.Println(val)

			fmt.Println("InfluxDB logging: ", tags, fields)
			<- sem
		}()
	}

	err := influxClient.Write(bp)
	uct.CheckError(err)
}

func getSectionStatus() (sectionsStatuses []SectionsStatus) {
	query := `SELECT count(*), section.status, subject.season, subject.year, university.topic_name university, subject.topic_name subject, course.topic_name course FROM section
				JOIN course ON section.course_id = course.id
				JOIN subject ON course.subject_id = subject.id
				JOIN university ON subject.university_id = university.id
			  GROUP BY subject.season, subject.year, university, subject, section.status, course
			  ORDER BY subject, university, season, year;`

	err := database.Select(&sectionsStatuses, query)
	uct.CheckError(err)
	return
}
