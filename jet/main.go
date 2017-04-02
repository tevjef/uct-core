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

	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/uct-core/common/conf"
	"github.com/tevjef/uct-core/common/model"
	"github.com/tevjef/uct-core/common/redis"
	"github.com/tevjef/uct-core/common/redis/harmony"
	"golang.org/x/net/context"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type jet struct {
	app    *kingpin.ApplicationModel
	config *jetConfig
	redis  *redis.Helper
	ctx    context.Context
}

type jetConfig struct {
	service        conf.Config
	inputFormat    string
	outputFormat   string
	daemonInterval time.Duration
	daemonFile     string
	scraperName    string
	scraperCommand string
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
}

func main() {
	jconf := jetConfig{}
	app := kingpin.New("jet", "A program the wraps a uct scraper and collect it's output")

	app.Flag("output-format", "Choose output format").Short('f').
		HintOptions(model.Protobuf, model.Json).
		PlaceHolder("[protobuf, json]").
		Required().
		Envar("JET_OUTPUT_FORMAT").
		EnumVar(&jconf.outputFormat, "protobuf", "json")

	app.Flag("input-format", "Choose input format").
		HintOptions(model.Protobuf, model.Json).
		PlaceHolder("[protobuf, json]").
		Required().
		Envar("JET_INPUT_FORMAT").
		EnumVar(&jconf.inputFormat, "protobuf", "json")

	app.Flag("daemon", "Run as a daemon with a refesh interval").
		Envar("JET_DEAMON").
		DurationVar(&jconf.daemonInterval)

	app.Flag("daemon-dir", "If supplied the deamon will write files to this directory").
		Envar("JET_DAEMON_DIR").
		ExistingDirVar(&jconf.daemonFile)

	app.Flag("scraper-name", "The scraper name, used in logging").
		Required().
		Envar("JET_SCRAPER_NAME").
		StringVar(&jconf.scraperName)

	app.Flag("scraper", "The scraper this program wraps, the name of the executable").
		Required().
		Envar("JET_SCRAPER_PATH").
		StringVar(&jconf.scraperCommand)

	configFile := app.Flag("config", "configuration file for the application").
		Short('c').
		Envar("JET_CONFIG").
		File()

	kingpin.MustParse(app.Parse(deleteArgs(os.Args[1:])))
	app.Name = jconf.scraperName

	// Parse configuration file
	jconf.service = conf.OpenConfigWithName(*configFile, app.Name)

	// Start profiling
	go model.StartPprof(jconf.service.DebugSever(app.Name))

	(&jet{
		app: app.Model(),
		config: &jconf,
		redis: redis.NewHelper(jconf.service, app.Name),
	}).init()
}

func (jet *jet) init() {
	// Channel to send scraped data on
	resultChan := make(chan model.University)

	// Runs at regular intervals
	if jet.config.daemonInterval > 0 {
		// Override cli arg with environment variable
		if intervalFromEnv := jet.config.service.Scrapers.Get(jet.app.Name).Interval; intervalFromEnv != "" {
			if interval, err := time.ParseDuration(intervalFromEnv); err != nil {
				log.WithError(err).Fatalln("failed to parse duration")
			} else if interval > 0 {
				jet.config.daemonInterval = interval
			}
		}

		harmony.DaemonScraper(jet.redis, jet.config.daemonInterval, func(ctx context.Context) {
			jet.entryPoint(resultChan)
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

	var school model.University

	cmd := exec.Command(jet.config.scraperCommand, parseArgs(os.Args)...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// DATA RACE!!!
	go io.Copy(os.Stderr, stderr)

	if err = model.UnmarshalMessage(jet.config.inputFormat, stdout, &school); err != nil {
		school = model.University{}
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	if school.Name == "" {
		return
	} else {
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
