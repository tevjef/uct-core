package common

import (
	log "github.com/Sirupsen/logrus"
	"fmt"
	"os"
	"net"
	"github.com/BurntSushi/toml"
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

	UCT_SPIKE_API_KEY

)

func (env Env) String() string {
	switch env {
	case UCT_DB_HOST:
		return "UCT_DB_HOST"
	case UCT_DB_NAME:
		return "UCT_DB_NAME"
	case UCT_DB_PASSWORD:
		return "UCT_DB_PASSWORD"
	case UCT_DB_USER:
		return "UCT_DB_USER"
	case UCT_DB_PORT:
		return "UCT_DB_PORT"
	case UCT_INFLUX_HOST:
		return "UCT_INFLUX_HOST"
	case UCT_INFLUX_USER:
		return "UCT_INFLUX_USER"
	case UCT_INFLUX_PASSWORD:
		return "UCT_INFLUX_PASSWORD"
	case UCT_SPIKE_API_KEY:
		return "UCT_SPIKE_API_KEY"
	default:
		return ""
	}
}

type pprof map[string]server

type Config struct {
	Db     database `toml:"postgres"`
	Pprof  pprof `toml:"pprof"`
	Influx Influx `toml:"influx"`
	Spike spike `toml:"spike"`
	Hermes hermes `toml:"hermes"`
}

type spike struct {

}

type hermes struct {
	ApiKey string `toml:"api_key"`
}

type server struct {
	Host string `toml:"host"`
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
}

func bindEnv(defValue string, env fmt.Stringer) string {
	value := os.Getenv(env.String())
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
		c.Db.User, c.Db.Name, c.Db.Password, c.Db.Host,c.Db.Port, appName)
}