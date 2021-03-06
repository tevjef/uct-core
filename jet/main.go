package main

import (
	"bytes"
	"errors"
	"hash/fnv"
	"io"
	"io/ioutil"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"context"

	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/conf"
	_ "github.com/tevjef/uct-backend/common/metrics"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/common/redis"
	"github.com/tevjef/uct-backend/common/redis/harmony"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type jet struct {
	app     *kingpin.ApplicationModel
	config  *jetConfig
	redis   *redis.Helper
	metrics metrics
	ctx     context.Context
}

type metrics struct {
	scraperBytes    *prometheus.GaugeVec
	scraperDuration *prometheus.GaugeVec
	scrapeCount     *prometheus.CounterVec
}

type jetConfig struct {
	service        conf.Config
	inputFormat    string
	outputFormat   string
	daemonInterval time.Duration
	daemonJitter   int
	daemonFile     string
	scraperName    string
	scraperCommand string
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
}

func main() {
	jconf := jetConfig{}
	app := kingpin.New("jet", "A program the wraps a uct scraper and collect it's output").DefaultEnvars()

	app.Flag("output-format", "Choose output format").Short('f').
		HintOptions(model.Protobuf, model.Json).
		PlaceHolder("[protobuf, json]").
		Default("protobuf").
		EnumVar(&jconf.outputFormat, "protobuf", "json")

	app.Flag("input-format", "Choose input format").
		HintOptions(model.Protobuf, model.Json).
		PlaceHolder("[protobuf, json]").
		Default("protobuf").
		EnumVar(&jconf.inputFormat, "protobuf", "json")

	app.Flag("daemon", "Run as a daemon with a refesh interval. -1 to disable").
		DurationVar(&jconf.daemonInterval)

	app.Flag("daemon-jitter", "Jitter to add to the daemon interval").
		IntVar(&jconf.daemonJitter)

	app.Flag("daemon-dir", "If supplied the deamon will write files to this directory").
		ExistingDirVar(&jconf.daemonFile)

	app.Flag("scraper-name", "The scraper name, used in logging").
		Required().
		StringVar(&jconf.scraperName)

	app.Flag("scraper", "The scraper this program wraps, the name of the executable").
		Required().
		StringVar(&jconf.scraperCommand)

	configFile := app.Flag("config", "configuration file for the application").
		Short('c').
		File()

	kingpin.MustParse(app.Parse(deleteArgs(os.Args[1:])))
	app.Name = jconf.scraperName

	// Parse configuration file
	jconf.service = conf.OpenConfigWithName(*configFile, app.Name)

	// Start profiling
	go model.StartPprof(jconf.service.DebugSever(app.Name))

	appMetrics := metrics{
		scraperBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scraper_payload_bytes",
			Help: "Bytes scraped by the scraper",
		}, []string{"scraper_name"}),

		scraperDuration: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scraper_duration_seconds",
			Help: "Time taken for the scraper to scrape.",
		}, []string{"scraper_name"}),

		scrapeCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "scraper_trigger_count",
			Help: "Times the scraper has been triggered",
		}, []string{"scraper_name"}),
	}

	prometheus.MustRegister(
		appMetrics.scrapeCount,
		appMetrics.scraperBytes,
		appMetrics.scraperDuration,
	)

	(&jet{
		app:     app.Model(),
		config:  &jconf,
		redis:   redis.NewHelper(jconf.service, app.Name),
		metrics: appMetrics,
	}).init()
}

func (jet *jet) init() {
	// Channel to send scraped data on
	resultChan := make(chan model.University)

	// Runs at regular intervals
	if jet.config.daemonInterval > 0 {
		harmony.DaemonScraper(&harmony.Config{
			Redis:    jet.redis,
			Interval: jet.config.daemonInterval,
			Action: func(ctx context.Context) {
				jet.entryPoint(resultChan)
			},
			Jitter: jet.config.daemonJitter,
			Ctx:    context.Background(),
		})
	} else {
		go func() {
			jet.entryPoint(resultChan)
			close(resultChan)
		}()
	}

	// block as it waits for results to come in
	for school := range resultChan {
		if school.Name == "" {
			continue
		}

		reader, err := model.MarshalMessage(jet.config.outputFormat, school)
		if err != nil {
			log.WithError(err).Fatal()
		}

		// Write to file
		if jet.config.daemonFile != "" {
			if data, err := ioutil.ReadAll(reader); err != nil {
				log.WithError(err).Fatalln("failed to read all data")
			} else {
				fileName := jet.config.daemonFile + "/" + jet.app.Name + "-" + strconv.FormatInt(time.Now().Unix(), 10) + "." + jet.config.outputFormat
				log.Debugln("Writing file", fileName)
				if err = ioutil.WriteFile(fileName, data, 0644); err != nil {
					log.WithError(err).Fatalln("failed to write file")
				}
			}
			continue
		}

		// Write to redis
		if jet.config.daemonInterval > 0 {
			jet.pushToRedis(reader)
			continue
		}

		// Write to stdout
		io.Copy(os.Stdout, reader)
	}
}

func (jet *jet) pushToRedis(reader *bytes.Reader) {
	if data, err := ioutil.ReadAll(reader); err != nil {
		log.WithError(err).Fatalln("failed to read all data")
	} else {
		jet.metrics.scraperBytes.WithLabelValues(jet.app.Name).Set(float64(len(data)))
		log.WithFields(log.Fields{"scraper_name": jet.app.Name, "bytes": len(data), "hash": hash(data)}).Info()
		if err := jet.redis.Client.Set(jet.redis.NameSpace+":data:latest", data, 0).Err(); err != nil {
			log.Fatalln(errors.New("failed to connect to redis server"))
		}

		if _, err := jet.redis.LPushNotExist(redis.ScraperQueue, jet.redis.NameSpace); err != nil {
			log.Fatalln(errors.New("failed to queue univeristiy for upload"))
		}
	}
}

func hash(s []byte) string {
	h := fnv.New32a()
	h.Write(s)
	return strconv.Itoa(int(h.Sum32()))
}

func (jet *jet) entryPoint(result chan model.University) {
	starTime := time.Now()
	jet.metrics.scrapeCount.WithLabelValues(jet.app.Name).Inc()

	var school model.University

	cmd := exec.Command(jet.config.scraperCommand, parseArgs(os.Args)...)
	cmd.Stderr = os.Stdout

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	if err = model.UnmarshalMessage(jet.config.inputFormat, stdout, &school); err != nil {
		school = model.University{}
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	if school.Name == "" {
		log.Info("no school data returned:", err)
		return
	} else {
		jet.metrics.scraperDuration.WithLabelValues(jet.app.Name).Set(time.Since(starTime).Seconds())
		log.WithFields(log.Fields{"scraper_name": jet.app.Name, "elapsed": time.Since(starTime).Seconds()}).Info()
		result <- school
	}
}

func parseArgs(str []string) []string {
	for i, val := range str {
		if val == "--scraper" {
			return str[i+2:]
		}
	}
	return str
}

func deleteArgs(str []string) []string {
	for i, val := range str {
		if val == "--scraper" {
			return str[:i+2]
		}
	}
	return str
}
