package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"golang.org/x/net/context"
	"uct/common/conf"
	"uct/redis"
)

type ein struct {
	app      *kingpin.ApplicationModel
	config   *einConfig
	redis    *redis.Helper
	postgres DatabaseHandler
	ctx      context.Context
}

type einConfig struct {
	service     conf.Config
	noDiff      bool
	fullUpsert  bool
	inputFormat string
}

type serial struct {
	TopicName string `db:"topic_name"`
	Data      []byte `db:"data"`
}

type serialSubject struct {
	serial
}

type serialCourse struct {
	serial
}

type serialSection struct {
	serial
}