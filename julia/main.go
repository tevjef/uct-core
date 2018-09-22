package main

import (
	_ "net/http/pprof"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/lib/pq"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tevjef/uct-backend/common/conf"
	_ "github.com/tevjef/uct-backend/common/metrics"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/common/notification"
	"github.com/tevjef/uct-backend/common/redis"
	"github.com/tevjef/uct-backend/julia/notifier"
	"golang.org/x/net/context"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	notificationsIn = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "julia_notifications_in_count",
		Help: "Number notifications received by Julia",
	}, []string{"university_name", "status"})

	notificationsOut = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "julia_notifications_out_count",
		Help: "Number notifications processed by Julia",
	}, []string{"university_name", "status"})
)

type julia struct {
	app      *kingpin.ApplicationModel
	config   *juliaConfig
	redis    *redis.Helper
	notifier notifier.Notifier
	process  *Process
	ctx      context.Context
}

type juliaConfig struct {
	service        conf.Config
	inputFormat    string
	outputFormat   string
	daemonInterval time.Duration
	daemonFile     string
	scraperName    string
	scraperCommand string
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	prometheus.MustRegister(notificationsIn, notificationsOut)
}

func main() {
	jconf := &juliaConfig{}

	app := kingpin.New("julia", "An application that queue messages from the database")

	configFile := app.Flag("config", "configuration file for the application").
		Short('c').
		Envar("JULIA_CONFIG").
		File()

	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Parse configuration file
	jconf.service = conf.OpenConfigWithName(*configFile, app.Name)

	// Start profiling
	go model.StartPprof(jconf.service.DebugSever(app.Name))

	// Create a Postgresql event listener
	ch := make(chan struct{})
	listener := pq.NewListener(jconf.service.DatabaseConfig(app.Name), 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if ev != pq.ListenerEventConnectionAttemptFailed {
			ch <- struct{}{}
		} else {
			log.WithError(err).Warningln("listener failed to establish connection")
		}
	})

	select {
	case <-ch:
	case <-time.After(2 * time.Minute):
		log.Fatalln("failed to create listener")
	}

	if err := listener.Listen("status_events"); err != nil {
		log.WithError(err).Fatalln("failed to listen on channel")
	}

	(&julia{
		app:      app.Model(),
		config:   jconf,
		redis:    redis.NewHelper(jconf.service, app.Name),
		notifier: notifier.NewNotifier(listener),
		process: &Process{
			in:  make(chan model.UCTNotification),
			out: make(chan model.UCTNotification),
		},
		ctx: context.TODO(),
	}).init()
}

func (julia *julia) init() {
	go julia.process.Run(julia.dispatch)

	// Open connection to postgresql
	log.Infoln("start monitoring PostgreSQL...")

	for {
		waitForNotification(julia.ctx, julia.notifier, julia.process.Recv)
	}
}

func (julia *julia) dispatch(uctNotification model.UCTNotification) {
	label := prometheus.Labels{
		"university_name": uctNotification.University.TopicName,
		"status":          uctNotification.Status,
	}
	notificationsOut.With(label).Inc()
	log.WithFields(log.Fields{"topic": uctNotification.TopicName, "university_name": uctNotification.University.TopicName}).Infoln("queueing")
	if notificationBytes, err := uctNotification.Marshal(); err != nil {
		log.WithError(err).Fatalln("failed to marshall notification")
	} else if _, err := julia.redis.Client.Set(notification.MainQueueData+uctNotification.TopicName, notificationBytes, time.Hour).Result(); err != nil {
		log.WithError(err).Warningln("failed to set notification data")
	} else if julia.redis.RPush(notification.MainQueue, uctNotification.TopicName); err != nil {
		log.WithError(err).Warningln("failed to push notification unto queue")
	}
}

func waitForNotification(ctx context.Context, l notifier.Notifier, onNotify func(notification *model.UCTNotification)) {
	for {
		select {
		case message, ok := <-l.Notify():
			if !ok {
				return
			}
			if message == "" {
				continue
			}

			go func() {
				var uctNotification model.UCTNotification
				if err := ffjson.UnmarshalFast([]byte(message), &uctNotification); err != nil {
					log.WithError(err).Errorln("failed to unmarsahll notification")
					return
				}

				label := prometheus.Labels{
					"university_name": uctNotification.University.TopicName,
					"status":          uctNotification.Status,
				}

				notificationsIn.With(label).Inc()

				onNotify(&uctNotification)
			}()

			// Received no notification from the database for 60 seconds.
		case <-time.After(1 * time.Minute):
			go l.Ping()
		case <-ctx.Done():
			return
		}
	}
}
