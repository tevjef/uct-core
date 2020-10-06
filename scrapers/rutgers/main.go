package rutgers

import (
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/common/publishing"
	_ "github.com/tevjef/uct-backend/common/trace"
	"go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/trace"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyLevel: "severity",
			log.FieldKeyMsg:   "message",
		},
	})
	log.SetLevel(log.DebugLevel)
}

func RutgersScraper(w http.ResponseWriter, r *http.Request) {
	MainFunc(r)

	w.WriteHeader(http.StatusOK)
}

func MainFunc(r *http.Request) {
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

	sc, _ := (&propagation.HTTPFormat{}).SpanContextFromRequest(r)
	ctx, span := trace.StartSpanWithRemoteParent(r.Context(), "/func.RutgersScraper", sc, trace.WithSpanKind(trace.SpanKindServer))
	span.AddAttributes(trace.StringAttribute("campus", rconf.campus))
	defer span.End()

	(&rutgers{
		app:    app.Model(),
		config: rconf,
		ctx:    ctx,
	}).init()
}

func (rutgers *rutgers) init() {
	uni := rutgers.getCampus(rutgers.config.campus)
	if reader, err := model.MarshalMessage(rutgers.config.outputFormat, uni); err != nil {
		log.WithError(err).Warningf("%s: failed to parse university", rutgers.config.campus)
	} else if len(uni.Subjects) == 0 {
		log.WithError(err).Warningf("%s: incomplete university, exiting...", rutgers.config.campus)
	} else {
		if rutgers.config.outputHttpUrl != "" {
			err := publishing.PublishToHttp(rutgers.ctx, rutgers.config.outputHttpUrl, reader)
			if err != nil {
				log.Fatal(err)
			}
			log.WithField("topic_name", uni.TopicName).Infof("%s: sent scraping result to: %s", rutgers.config.campus, rutgers.config.outputHttpUrl)
		} else {
			io.Copy(os.Stdout, reader)
		}
	}
}
