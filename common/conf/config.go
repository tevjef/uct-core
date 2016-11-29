package conf

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
)

type pprof map[string]server
type scrapers map[string]*scraper

type Config struct {
	AppName  string
	Postgres Postgres `toml:"postgres"`
	Redis    Redis    `toml:"redis"`
	Pprof    pprof    `toml:"pprof"`
	InfluxDb InfluxDb `toml:"influxdb"`
	Spike    spike    `toml:"spike"`
	Julia    julia    `toml:"julia"`
	Hermes   hermes   `toml:"hermes"`
	Scrapers scrapers `toml:"scrapers"`
}

type spike struct {
	RedisDb int `toml:"redis_db" envconfig:"REDIS_DB"`
}

type julia struct {
}

type hermes struct {
	ApiKey string `toml:"api_key" envconfig:"FCM_API_KEY"`
}

type server struct {
	Host     string `toml:"host"`
	Password string `toml:"password"`
	Enabled  bool
}

type scraper struct {
	Interval string `toml:"interval"`
}

type Postgres struct {
	User     string `toml:"user" envconfig:"POSTGRES_USER"`
	Host     string `toml:"host"  envconfig:"POSTGRES_PORT_5432_TCP_ADDR"`
	Port     string `toml:"port" envconfig:"POSTGRES_PORT_5432_TCP_PORT"`
	Password string `toml:"password" envconfig:"POSTGRES_PASSWORD"`
	Name     string `toml:"name" envconfig:"POSTGRES_DB"`
	ConnMax  int    `toml:"connection_max" envconfig:"POSTGRES_MAX_CONNECTIONS"`
}

type Redis struct {
	Host     string `toml:"host" envconfig:"REDIS_PORT_6379_TCP_ADDR"`
	Port     string `toml:"port" envconfig:"REDIS_PORT_6379_TCP_PORT"`
	Password string `toml:"password" envconfig:"REDIS_PASSWORD"`
	Db       int    `toml:"db" envconfig:"REDIS_DB"`
}

type InfluxDb struct {
	User     string `toml:"user" envconfig:"INFLUXDB_ADMIN_USER"`
	Port     string `toml:"port" envconfig:"INFLUXDB_PORT_8086_TCP_PORT"`
	Host     string `toml:"host" envconfig:"INFLUXDB_PORT_8086_TCP_ADDR"`
	Password string `toml:"password" envconfig:"INFLUXDB_ADMIN_PASSWORD"`
}

func (pprof pprof) Get(key string) server {
	return pprof[key]
}

func (scrapers scrapers) Get(key string) *scraper {
	return scrapers[key]
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

func IsDebug() bool {
	value := os.Getenv("UCT_DEBUG")
	b, _ := strconv.ParseBool(value)
	return b
}

func OpenConfig(file *os.File) Config {
	c := Config{}
	if _, err := toml.DecodeReader(file, &c); err != nil {
		log.Fatalln("Error while decoding config file checking environment:", err)
		return c
	}

	c.fromEnvironment()

	return c
}

func (c *Config) fromEnvironment() {

	err := envconfig.Process("", &c.InfluxDb)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = envconfig.Process("", &c.Postgres)
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

func (c Config) GetDebugSever(appName string) *net.TCPAddr {
	value := c.Pprof[appName].Host
	if addr, err := net.ResolveTCPAddr("tcp", value); err != nil {
		log.Panicf("'%s' is not a valid TCP address: %s", value, err)
		return nil
	} else {
		return addr
	}
}

func (c Config) GetDbConfig(appName string) string {
	return fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s fallback_application_name=%s sslmode=disable",
		c.Postgres.User, c.Postgres.Name, c.Postgres.Password, c.Postgres.Host, c.Postgres.Port, appName)
}

func (c Config) GetInfluxAddr() string {
	return "http://" + c.InfluxDb.Host + ":" + c.InfluxDb.Port
}

func (c Config) GetRedisAddr() string {
	return c.Redis.Host + ":" + c.Redis.Port
}
