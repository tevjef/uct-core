package rutgers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/common/publishing"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
}

func RutgersScraper(w http.ResponseWriter, r *http.Request) {
	MainFunc()

	fmt.Fprint(w, "Complete")
}

func MainFunc() {
	rconf := &rutgersConfig{}

	app := kingpin.New("rutgers", "A web scraper that retrives course information for Rutgers University's servers.")

	app.Flag("output-http-url", "Choose endpoint to send results to.").
		Default("").
		Envar("OUTPUT_HTTP_URL").
		StringVar(&rconf.outputHttpUrl)

	app.Flag("campus", "Choose campus code. NB=New Brunswick, CM=Camden, NK=Newark").
		Short('u').
		PlaceHolder("[CM, NK, NB]").
		Required().
		Envar("RUTGERS_CAMPUS").
		EnumVar(&rconf.campus, "CM", "NK", "NB")

	app.Flag("format", "choose output format").
		Short('f').
		HintOptions(model.Json, model.Protobuf).
		PlaceHolder("[protobuf, json]").
		Default("protobuf").
		Envar("RUTGERS_OUTPUT_FORMAT").
		EnumVar(&rconf.outputFormat, "protobuf", "json")

	app.Flag("latest", "Only output the current and next semester").
		Short('l').
		Envar("RUTGERS_LATEST").
		BoolVar(&rconf.latest)

	kingpin.MustParse(app.Parse([]string{}))
	app.Name = app.Name + "-" + rconf.campus

	(&rutgers{
		app:    app.Model(),
		config: rconf,
		ctx:    context.TODO(),
	}).init()
}

func (rutgers *rutgers) init() {
	uni := rutgers.getCampus(rutgers.config.campus)
	if reader, err := model.MarshalMessage(rutgers.config.outputFormat, uni); err != nil {
		log.WithError(err).Fatal()
	} else {
		if rutgers.config.outputHttpUrl != "" {
			err := publishing.PublishToHttp(rutgers.config.outputHttpUrl, reader)
			if err != nil {
				log.Fatal(err)
			}
			log.WithField("topic_name", uni.TopicName).Infof("scraping complete")
		} else {
			io.Copy(os.Stdout, reader)
		}
	}
}
