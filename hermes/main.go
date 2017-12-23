package main

import (
	_ "net/http/pprof"
	"os"
	"time"

	"strconv"

	log "github.com/Sirupsen/logrus"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tevjef/go-fcm"
	"github.com/tevjef/uct-core/common/conf"
	"github.com/tevjef/uct-core/common/database"
	_ "github.com/tevjef/uct-core/common/metrics"
	"github.com/tevjef/uct-core/common/model"
	"github.com/tevjef/uct-core/common/notification"
	"github.com/tevjef/uct-core/common/redis"
	"github.com/tevjef/uct-core/common/try"
	"golang.org/x/net/context"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	notificationsIn = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hermes_notifications_in_count",
		Help: "Number notifications received by Hermes",
	}, []string{"university_name", "status"})

	notificationsOut = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hermes_notifications_out_count",
		Help: "Number notifications processed by Heremes",
	}, []string{"university_name", "status"})
	fcmElapsed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hermes_fcm_elapsed_second",
		Help: "Time taken to send notification",
	}, []string{"university_name", "status"})

	fcmElapsedHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "hermes_histogram_fcm_elapsed_second",
		Help: "Time taken to send notification",
	}, []string{"university_name", "status"})
)

type hermes struct {
	app       *kingpin.ApplicationModel
	config    *hermesConfig
	fcmClient *fcm.Client
	redis     *redis.Helper
	postgres  database.Handler
	ctx       context.Context
}

type hermesConfig struct {
	service           conf.Config
	dryRun            bool
	firebaseProjectID string
	tokenServerAddr   string
	tokenServerPort   int16
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	prometheus.MustRegister(notificationsIn, notificationsOut, fcmElapsed, fcmElapsedHistogram)
}

func main() {
	hconf := &hermesConfig{}

	app := kingpin.New("hermes", "A server that listens to a database for events and publishes notifications to Firebase Cloud Messaging")

	app.Flag("dry-run", "enable dry-run").
		Short('d').
		Default("true").
		Envar("HERMES_DRY_RUN").
		BoolVar(&hconf.dryRun)

	app.Flag("firebase-project-id", "Firebase project Id").
		Default("vt").
		Envar("FIREBASE_PROJECT_ID").
		StringVar(&hconf.firebaseProjectID)

	app.Flag("token-server-addr", "Token server address").
		Default("vt").
		Envar("TOKEN_SERVER_ADDR").
		StringVar(&hconf.tokenServerAddr)

	app.Flag("token-server-port", "Token server port").
		Default("9875").
		Envar("TOKEN_SERVER_PORT").
		Int16Var(&hconf.tokenServerPort)

	configFile := app.Flag("config", "configuration file for the application").
		Short('c').
		Envar("HERMES_CONFIG").
		File()

	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Parse configuration file
	hconf.service = conf.OpenConfigWithName(*configFile, app.Name)

	if hconf.dryRun {
		log.Infoln("Enabling FCM in dry run mode")
	} else {
		log.Infoln("Enabling FCM in production mode")
	}

	// Open database connection
	pgDatabase, err := model.OpenPostgres(hconf.service.DatabaseConfig(app.Name))
	if err != nil {
		log.WithError(err).Fatalln("failed to open database connection")
	}

	provider := &tokenProvider{hconf.tokenServerAddr, strconv.Itoa(int(hconf.tokenServerPort))}
	fcmClient, err := fcm.NewClient(hconf.firebaseProjectID, provider)

	// Start profiling
	go model.StartPprof(hconf.service.DebugSever(app.Name))

	(&hermes{
		app:       app.Model(),
		config:    hconf,
		fcmClient: fcmClient,
		redis:     redis.NewHelper(hconf.service, app.Name),
		postgres:  database.NewHandler(app.Name, pgDatabase, queries),
	}).init()
}

func (hermes *hermes) init() {
	resultChan := hermes.waitForPop()

	for {
		select {
		case pair := <-resultChan:
			go hermes.recvNotification(pair)
		}
	}
}

func (hermes *hermes) recvNotification(pair notificationPair) {
	label := prometheus.Labels{
		"university_name": pair.n.University.TopicName,
		"status":          pair.n.Status,
	}

	notificationsIn.With(label).Inc()
	log.WithFields(log.Fields{"university_name": pair.n.University.TopicName,
		"notification_id": pair.n.NotificationId, "status": pair.n.Status,
		"topic": pair.n.TopicName}).Info("postgres_notification")

	defer func(start time.Time) {
		fcmElapsed.With(label).Set(time.Since(start).Seconds())
		fcmElapsedHistogram.With(label).Observe(time.Since(start).Seconds())
		log.WithFields(log.Fields{"elapsed": time.Since(start).Seconds() * 1e3,
			"university_name": pair.n.University.TopicName,
			"name":            "send_notification"}).Infoln()
	}(time.Now())

	// Retry in case of SSL/TLS timeout errors. FCM itself should be rock solid
	err := try.Do(func(attempt int) (retry bool, err error) {
		if err = hermes.sendFcmNotification(pair); err != nil {
			return true, err
		}
		return false, nil
	})

	if err != nil {
		log.WithError(err).Errorln()
	}

	notificationsOut.With(label).Inc()
}

func (hermes *hermes) waitForPop() chan notificationPair {
	c := make(chan notificationPair)
	go func() {
		for {
			if pair, err := hermes.popNotification(); err == nil {
				c <- *pair
			} else {
				log.WithError(err).Warningln()
			}
		}

	}()

	go func() {
		time.Sleep(time.Second * 10)

		var rawN = `
{
    "university": {"id": 3, "abbr": "RU-NB", "name": "Rutgers University–New Brunswick", "subjects": [{"id": 1048, "name": "Sociology", "year": "2016", "number": "920", "season": "fall", "courses": [{"id": 14785, "name": "Soc Mental Illness", "number": "307", "sections": [{"id": 33259, "max": 1, "now": 0, "number": "02", "status": "Closed", "credits": "3.0", "topic_id": "17834289047972563637", "course_id": 14785, "created_at": "2016-05-30T07:05:20.93539", "topic_name": "rutgers.universitynew.brunswick.920.sociology.fall.2016.307.soc.mental.illness.02.11593", "updated_at": "2016-08-25T05:01:52.299825", "call_number": "11593"}], "topic_id": "3437346514840870658", "subject_id": 1048, "topic_name": "rutgers.universitynew.brunswick.920.sociology.fall.2016.307.soc.mental.illness"}], "topic_id": "473453796608589362", "topic_name": "rutgers.universitynew.brunswick.920.sociology.fall.2016", "university_id": 3}], "topic_id": "13265320283999417559", "home_page": "http://newbrunswick.rutgers.edu/", "main_color": "F44336", "topic_name": "rutgers.universitynew.brunswick", "registration_page": "https://sims.rutgers.edu/webreg/"},
    "topic_name": "rutgers.universitynewark.050.american.studies.summer.2017.200.intro.amer.studies.bq.04605",
    "notification_id": 1234567,
    "status": "Open"
}
`

		var noti model.UCTNotification
		err := noti.UnmarshalJSON([]byte(rawN))
		if err != nil {
			panic(err)
		}

		pair := notificationPair{
			n:   &noti,
			raw: rawN,
		}

		c <- pair
	}()

	return c
}

func (hermes *hermes) popNotification() (*notificationPair, error) {
	if topic, err := hermes.redis.Client.BRPopLPush(notification.MainQueue, notification.DoneQueue, 0).Result(); err == nil {
		if b, err := hermes.redis.Client.Get(notification.MainQueueData + topic).Bytes(); err != nil {
			return nil, errors.Wrap(err, "error getting notification data")
		} else {
			uctNotification := &model.UCTNotification{}
			if err := uctNotification.UnmarshalJSON(b); err != nil {
				return nil, err
			} else if jsonBytes, err := ffjson.Marshal(uctNotification); err != nil {
				return nil, err
			} else if _, err := hermes.redis.Client.Del(topic).Result(); err != nil {
				log.WithError(err).Warningln("failed to del topic data")
				return &notificationPair{n: uctNotification, raw: string(jsonBytes)}, nil
			} else {
				return &notificationPair{n: uctNotification, raw: string(jsonBytes)}, nil
			}
		}
	} else {
		return nil, err
	}
}

type notificationPair struct {
	n   *model.UCTNotification
	raw string
}

func (hermes *hermes) acknowledgeNotification(notificationId, messageId int64) int64 {
	args := map[string]interface{}{"notification_id": notificationId, "message_id": messageId}
	return hermes.postgres.Update(AckNotificationQuery, args)
}

var queries = []string{
	AckNotificationQuery,
}

const AckNotificationQuery = `UPDATE notification SET (ack_at, message_id) = (now(), :message_id) WHERE id = :notification_id RETURNING notification.id`
