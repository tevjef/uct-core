package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"
	"uct/common/conf"
	"uct/common/model"

	"uct/redis"
	"uct/redis/harmony"

	"os/exec"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"hash/fnv"
)

var (
	app            = kingpin.New("jet", "A program the wraps a uct scraper and collect it's output")
	outputFormat   = app.Flag("output-format", "Choose output format").Short('f').HintOptions(model.PROTOBUF, model.JSON).PlaceHolder("[protobuf, json]").Required().String()
	inputFormat    = app.Flag("input-format", "Choose input format").HintOptions(model.PROTOBUF, model.JSON).PlaceHolder("[protobuf, json]").Required().String()
	daemonInterval = app.Flag("daemon", "Run as a daemon with a refesh interval").Duration()
	daemonFile     = app.Flag("daemon-dir", "If supplied the deamon will write files to this directory").ExistingDir()
	configFile     = app.Flag("config", "configuration file for the application").Short('c').File()
	scraperName    = app.Flag("scraper-name", "The scraper name, used in logging").Required().String()
	command        = app.Flag("scraper", "The scraper this program wraps, the name of the executable").Required().String()
	config         conf.Config
	helper         *redis.Helper
)

func main() {
	kingpin.MustParse(app.Parse(deleteArgs(os.Args[1:])))
	app.Name = *scraperName

	log.SetLevel(log.InfoLevel)

	if *outputFormat != model.JSON && *outputFormat != model.PROTOBUF {
		log.Fatalln("Invalid format:", *outputFormat)
	}

	isDaemon := *daemonInterval > 0
	// Parse configuration file
	config = conf.OpenConfig(*configFile)
	config.AppName = app.Name

	// Channel to send scraped data on
	resultChan := make(chan model.University)

	// Runs at regular intervals
	if isDaemon {
		// Override cli arg with environment variable
		if intervalFromEnv := config.Scrapers.Get(app.Name).Interval; intervalFromEnv != "" {
			if interval, err := time.ParseDuration(intervalFromEnv); err != nil {
				model.CheckError(err)
			} else if interval > 0 {
				daemonInterval = &interval
			}
		}

		helper = redis.NewHelper(config, app.Name)

		harmony.DaemonScraper(helper, *daemonInterval, func(cancel chan struct{}) {
			entryPoint(resultChan)
		})

	} else {
		go func() {
			entryPoint(resultChan)
			close(resultChan)
		}()
	}

	// block as it waits for results to come in
	for school := range resultChan {
		if school.Name == "" {
			continue
		}

		reader, err := model.MarshalMessage(*outputFormat, school)
		if err != nil {
			log.WithError(err).Fatal()
		}

		// Write to file
		if *daemonFile != "" {
			if data, err := ioutil.ReadAll(reader); err != nil {
				model.CheckError(err)
			} else {
				fileName := *daemonFile + "/" + app.Name + "-" + strconv.FormatInt(time.Now().Unix(), 10) + "." + *outputFormat
				log.Debugln("Writing file", fileName)
				if err = ioutil.WriteFile(fileName, data, 0644); err != nil {
					model.CheckError(err)
				}
			}
			continue
		}

		// Write to redis
		if isDaemon {
			pushToRedis(reader)
			continue
		}

		// Write to stdout
		io.Copy(os.Stdout, reader)
	}
}

func pushToRedis(reader *bytes.Reader) {
	if data, err := ioutil.ReadAll(reader); err != nil {
		model.CheckError(err)
	} else {
		log.WithFields(log.Fields{"scraper_name": app.Name, "bytes": len(data), "hash": hash(data)}).Info()
		if err := helper.Client.Set(helper.NameSpace+":data:latest", data, 0).Err(); err != nil {
			log.Fatalln(errors.New("failed to connect to redis server"))
		}

		if _, err := helper.LPushNotExist(redis.ScraperQueue, helper.NameSpace); err != nil {
			log.Fatalln(errors.New("failed to queue univeristiy for upload"))
		}
	}
}

func hash(s []byte) string {
	h := fnv.New32a()
	h.Write(s)
	return strconv.Itoa(int(h.Sum32()))
}

func entryPoint(result chan model.University) {
	starTime := time.Now()

	var school model.University

	cmd := exec.Command(*command, parseArgs(os.Args)...)
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

	if err = model.UnmarshalMessage(*inputFormat, stdout, &school); err != nil {
		school = model.University{}
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	if school.Name == "" {
		return
	} else {
		log.WithFields(log.Fields{"scraper_name": app.Name, "elapsed": time.Since(starTime).Seconds()}).Info()
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
