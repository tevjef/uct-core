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
	UCT_DBHOST Env = iota
	UCT_DBUSER
	UCT_DBPASSWORD
	UCT_DBNAME
	UCT_DBPORT
)

func (env Env) String() string {
	switch env {
	case UCT_DBHOST:
		return "UCT_DBHOST"
	case UCT_DBNAME:
		return "UCT_DBNAME"
	case UCT_DBPASSWORD:
		return "UCT_DBPASSWORD"
	case UCT_DBUSER:
		return "UCT_DBUSER"
	case UCT_DBPORT:
		return "UCT_DBPORT"
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

	return c
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

func (c Config) GetDbConfig() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.Db.User, c.Db.Password, c.Db.Host,c.Db.Port, c.Db.Name)
}