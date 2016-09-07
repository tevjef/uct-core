package common

import (
	log "github.com/Sirupsen/logrus"
	"fmt"
	"os"
	"net"
	"github.com/BurntSushi/toml"
	"strconv"
)

type Env int

const (
	UCT_DB_HOST Env = iota
	UCT_DB_USER
	UCT_DB_PASSWORD
	UCT_DB_NAME
	UCT_DB_PORT

	UCT_INFLUX_HOST
	UCT_INFLUX_USER
	UCT_INFLUX_PASSWORD

	UCT_REDIS_HOST
	UCT_REDIS_DB
	UCT_REDIS_PASSWORD
	
	UCT_SCRAPER_RUTGERS_INTERVAL
	UCT_SCRAPER_RUTGERS_OFFSET

	UCT_SPIKE_API_KEY

)

func (env Env) String() string {
	switch env {
	case UCT_DB_HOST:
		return "DB_PORT_5432_TCP_ADDR"
	case UCT_DB_NAME:
		return "DB_ENV_POSTGRES_DB"
	case UCT_DB_PASSWORD:
		return "DB_ENV_POSTGRES_PASSWORD"
	case UCT_DB_USER:
		return "DB_ENV_POSTGRES_USER"
	case UCT_DB_PORT:
		return "DB_PORT_5432_TCP_PORT"

	case UCT_INFLUX_HOST:
		return "UCT_INFLUX_HOST"
	case UCT_INFLUX_USER:
		return "UCT_INFLUX_USER"
	case UCT_INFLUX_PASSWORD:
		return "UCT_INFLUX_PASSWORD"

	case UCT_REDIS_HOST:
		return "UCT_REDIS_HOST"
	case UCT_REDIS_DB:
		return "UCT_REDIS_DB"
	case UCT_REDIS_PASSWORD:
		return "UCT_REDIS_PASSWORD"
		
	case UCT_SCRAPER_RUTGERS_INTERVAL:
		return "UCT_SCRAPER_RUTGERS_INTERVAL"
	case UCT_SCRAPER_RUTGERS_OFFSET:
		return "UCT_SCRAPER_RUTGERS_OFFSET"


	case UCT_SPIKE_API_KEY:
		return "UCT_SPIKE_API_KEY"
	default:
		return ""
	}
}

type pprof map[string]server
type scrapers map[string]scraper

type Config struct {
	Db     database `toml:"postgres"`
	Redis     redis `toml:"redis"`
	Pprof  pprof `toml:"pprof"`
	Influx Influx `toml:"influx"`
	Spike spike `toml:"spike"`
	Hermes hermes `toml:"hermes"`
	Scrapers scrapers `toml:"scrapers"`
}

type redis struct {
	server
	Host string
	Db int
}

type spike struct {

}

type hermes struct {
	ApiKey string `toml:"api_key"`
}

type server struct {
	Host string `toml:"host"`
	Password string `toml:"password"`
	Enabled bool
}

type database struct {
	User     string `toml:"user"`
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	Password string `toml:"password"`
	Name     string `toml:"name"`
	ConnMax int `toml:"connection_max"`
}

type scraper struct {
	interval string
	offset string
}

type Influx struct {
	User     string `toml:"user"`
	Host     string `toml:"host"`
	Password string `toml:"password"`
}

func (pprof pprof) Get(key string) server {
	return pprof[key]
}

func NewConfig(file *os.File) Config {
	c := Config{}
	if _, err := toml.DecodeReader(file, &c); err != nil {
		log.Fatalln("Error while decoding config file checking environment:", err)
		return c
	}

	c.fromEnvironment()

	return c
}

func (c *Config) fromEnvironment() {
	// Database
	c.Db.User = bindEnv(c.Db.User, UCT_DB_USER)
	c.Db.Host = bindEnv(c.Db.Host, UCT_DB_HOST)
	c.Db.Port = bindEnv(c.Db.Port, UCT_DB_PASSWORD)
	c.Db.Name = bindEnv(c.Db.Name, UCT_DB_NAME)
	c.Db.Port = bindEnv(c.Db.Port, UCT_DB_PORT)

	// Influx
	c.Influx.User = bindEnv(c.Influx.User, UCT_INFLUX_HOST)
	c.Influx.Host = bindEnv(c.Influx.Host, UCT_INFLUX_USER)
	c.Influx.Password = bindEnv(c.Influx.Password, UCT_INFLUX_PASSWORD)

	// Redis
	c.Redis.Db = int(bindEnvInt(c.Redis.Db, UCT_REDIS_HOST))
	c.Redis.Host = bindEnv(c.Redis.Host, UCT_REDIS_DB)
	c.Redis.Password = bindEnv(c.Redis.Password, UCT_REDIS_PASSWORD)
}

func bindEnv(defValue string, env fmt.Stringer) string {
	value := os.Getenv(env.String())
	if value != "" {
		return value
	} else {
		return defValue
	}
}

func bindEnvInt(defValue int, env fmt.Stringer) int64 {
	value := bindEnv(strconv.Itoa(defValue), env)
	i, err := strconv.ParseInt(value, 10, 64)
	CheckError(err)
	return i
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
		c.Db.User, c.Db.Name, c.Db.Password, c.Db.Host,c.Db.Port, appName)
}