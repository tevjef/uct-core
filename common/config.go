package common

import (
	"fmt"
	"log"
	"os"
	"encoding/json"
)

type Env int

const (
	UCT_DBHOST Env = iota
	UCT_DBUSER
	UCT_DBPASSWORD
	UCT_DBNAME
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
	default:
		return ""
	}
}

type Config struct {
	DBUser string
	DBHost string
	DBPassword string
	DBName string
}

func (config *Config) New(file *os.File) string {
	var user string
	var host string
	var pass string
	var name string


	decoder := json.NewDecoder(file)
	configuration := Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Println("Error while decoding config file checking enrironment:", err)
	} else {
		return
	}

	user = os.Getenv(UCT_DBUSER)
	host = os.Getenv(UCT_DBHOST)
	pass = os.Getenv(UCT_DBPASSWORD)
	name = os.Getenv(UCT_DBNAME)


	return fmt.Sprintf("postgres://%s:%s@%s:5432/%s", config.DBUser, config.DBPassword, config.DBHost, config.DBName)
}

