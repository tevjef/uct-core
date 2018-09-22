package conf

import (
	"fmt"
	"net"
	_ "net/http/pprof"
	"os"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/tevjef/go-runtime-metrics/expvar"
)

var WorkingDir = os.Getenv("GOPATH") + "/src/github.com/tevjef/uct-backend/scrapers/cuny/"

type Config struct {
	AppName  string
	Postgres Postgres `toml:"postgres"`
	Redis    Redis    `toml:"redis"`
	Spike    spike    `toml:"spike"`
	Julia    julia    `toml:"julia"`
	Hermes   hermes   `toml:"hermes"`
}

type spike struct {
	RedisDb int `toml:"redis_db" envconfig:"REDIS_DB"`
}

type julia struct{}

type hermes struct {
	ApiKey string `toml:"api_key" envconfig:"FCM_API_KEY"`
}

type server struct {
	Host     string `toml:"host"`
	Password string `toml:"password"`
	Enabled  bool
}

type Postgres struct {
	User     string `toml:"user" envconfig:"POSTGRES_USER"`
	Host     string `toml:"host"  envconfig:"POSTGRES_HOST"`
	Port     string `toml:"port" envconfig:"POSTGRES_PORT"`
	Password string `toml:"password" envconfig:"POSTGRES_PASSWORD"`
	Name     string `toml:"name" envconfig:"POSTGRES_DB"`
	ConnMax  int    `toml:"connection_max" envconfig:"POSTGRES_MAX_CONNECTIONS"`
}

type Redis struct {
	Host     string `toml:"host" envconfig:"REDIS_SERVICE_HOST"`
	Port     string `toml:"port" envconfig:"REDIS_SERVICE_PORT"`
	Password string `toml:"password" envconfig:"REDIS_PASSWORD"`
	Db       int    `toml:"db" envconfig:"REDIS_DB"`
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
}

func OpenConfig(file *os.File) Config {
	return OpenConfigWithName(file, "")
}

func OpenConfigWithName(file *os.File, name string) Config {
	c := Config{}
	if _, err := toml.DecodeReader(file, &c); err != nil {
		log.Fatalln("error while decoding config file checking environment:", err)
	}

	c.fromEnvironment()
	c.AppName = name

	return c
}

func (c *Config) fromEnvironment() {
	err := envconfig.Process("", &c.Postgres)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = envconfig.Process("", &c.Redis)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = envconfig.Process("", &c.Hermes)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = envconfig.Process("", &c.Spike)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = envconfig.Process("", &c.Julia)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func bindEnv(defValue string, env string) string {
	value := os.Getenv(env)
	if value != "" {
		return value
	} else {
		return defValue
	}
}

func (c Config) DebugSever(appName string) net.Listener {
	listener, err := net.Listen("tcp", ":13100")
	if err != nil {
		listener, _ = net.Listen("tcp", ":0")
		log.Println("pprof on port...", listener.Addr().(*net.TCPAddr).Port)
	}

	return listener
}

func (c Config) DatabaseConfig(appName string) string {
	return fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s fallback_application_name=%s sslmode=disable",
		c.Postgres.User, c.Postgres.Name, c.Postgres.Password, c.Postgres.Host, c.Postgres.Port, appName)
}

func (c Config) RedisAddr() string {
	return c.Redis.Host + ":" + c.Redis.Port
}
