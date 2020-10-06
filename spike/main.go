package spike

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/storage"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	uctfirestore "github.com/tevjef/uct-backend/common/firestore"
	"github.com/tevjef/uct-backend/common/middleware"
	"github.com/tevjef/uct-backend/common/middleware/cache"
	"github.com/tevjef/uct-backend/common/middleware/trace"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	_ "github.com/tevjef/uct-backend/common/metrics"
	"gopkg.in/alecthomas/kingpin.v2"
)

type spike struct {
	app             *kingpin.ApplicationModel
	firebaseApp     *firebase.App
	storageClient   *storage.Client
	firestoreClient *firestore.Client
	uctFSClient     *uctfirestore.Client
	logger          *log.Entry

	ctx context.Context
}

var engine *gin.Engine
var initOnce sync.Once
var appInstance *spike

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyLevel: "severity",
			log.FieldKeyMsg:   "message",
		},
	})
	log.SetLevel(log.DebugLevel)

}

func Spike(w http.ResponseWriter, r *http.Request) {
	MainFunc(w, r)
}

func MainFunc(w http.ResponseWriter, r *http.Request) {
	initOnce.Do(func() {
		ctx := context.Background()
		app := kingpin.New("spike", "A command-line application to serve university course information")

		logger := log.WithFields(log.Fields{})

		firebaseConf := &firebase.Config{
			StorageBucket: "universitycoursetracker.appspot.com",
		}

		var cred option.ClientOption
		if os.Getenv("USER") == "tevjef" {
			cred = option.WithCredentialsFile(os.Getenv("GOOGLE_CREDENTIALS"))
		} else {
			credentials, _ := google.FindDefaultCredentials(ctx)
			cred = option.WithCredentials(credentials)
		}

		firebaseApp, err := firebase.NewApp(ctx, firebaseConf, cred)
		if err != nil {
			log.WithError(err).Errorln("failed to crate firebase app")
		}

		storageClient, err := firebaseApp.Storage(ctx)
		if err != nil {
			log.WithError(err).Errorln("failed to crate firebase storage client")
		}

		firestoreClient, err := firebaseApp.Firestore(ctx)
		if err != nil {
			log.WithError(err).Errorln("failed to create firestore client")
		}

		uctFSClient := uctfirestore.NewClient(ctx, firestoreClient, logger)

		appInstance = &spike{
			app:             app.Model(),
			firebaseApp:     firebaseApp,
			storageClient:   storageClient,
			firestoreClient: firestoreClient,
			uctFSClient:     uctFSClient,
			logger:          logger,
			ctx:             ctx,
		}

		initGin()
	})

	appInstance.init(w, r)
}

func initGin() {
	engine = gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.Ginrus())
	engine.Use(trace.TraceMiddleware)
	engine.Use(uctfirestore.Firestore(appInstance.uctFSClient))
	engine.Use(cache.Cache(cache.NewInMemoryStore(10 * time.Second)))
	// does not cache and defaults to json
	v1 := engine.Group("/v1")
	{
		v1.Use(middleware.ContentNegotiation(middleware.JsonContentType))
		v1.Use(middleware.Decorator)

		v1.GET("/universities", universitiesHandler(0))
		v1.GET("/university/:topic", universityHandler(0))
		v1.GET("/subjects/:topic/:season/:year", subjectsHandler(0))
		v1.GET("/subject/:topic", subjectHandler(0))
		v1.GET("/courses/:topic", coursesHandler(0))
		//v1.GET("/course/:topic", courseHandler(0))
		v1.GET("/section/:topic", sectionHandler(0))
		v1.POST("/subscription", subscriptionHandler())
		v1.POST("/notification", notificationHandler())
	}

	// v2 caches responses and defaults to protobuf
	v2 := engine.Group("/v2")
	{
		v2.Use(middleware.ContentNegotiation(middleware.ProtobufContentType))
		v2.Use(middleware.Decorator)

		v2.GET("/universities", universitiesHandler(time.Minute))
		v2.GET("/university/:topic", universityHandler(time.Minute))
		v2.GET("/subjects/:topic/:season/:year", subjectsHandler(time.Minute))
		v2.GET("/subject/:topic", subjectHandler(10*time.Second))
		v2.GET("/courses/:topic", coursesHandler(10*time.Second))
		//v2.GET("/course/:topic", courseHandler(10*time.Second))
		v2.GET("/course/:topic/hotness/view", hotnessHandler(10*time.Second))
		v2.GET("/section/:topic", sectionHandler(10*time.Second))
		v2.POST("/subscription", subscriptionHandler())
		v2.POST("/notification", notificationHandler())
	}

	static := engine.Group("/static")
	static.GET("/:file", serveStaticFromGithub)
}

func (spike *spike) init(wr http.ResponseWriter, req *http.Request) {
	engine.ServeHTTP(wr, req)
}
