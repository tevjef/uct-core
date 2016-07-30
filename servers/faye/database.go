package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"strconv"
)

var (
	database      *sqlx.DB
	preparedStmts = make(map[string]*sqlx.NamedStmt)
)

func acknowledgeNotification(topicName, messageId string) (id int64) {
	messageIdInt, _ := strconv.ParseInt(messageId, 10, 64)
	args := map[string]interface{}{"topic_name": topicName, "message_id": messageIdInt}

	if rows, err := GetCachedStmt(AckNotificationQuery).Queryx(args); err != nil {
		log.WithFields(log.Fields{"args": args}).Panic(err)
	} else {
		count := 0
		for rows.Next() {
			count++
			if err = rows.Scan(&id); err != nil {
				log.WithFields(log.Fields{"args": args}).Panic(err)
			}
			rows.Close()
			log.WithFields(log.Fields{"args": args}).Info()
		}
		if count > 1 {
			log.WithFields(log.Fields{"args": args}).Panic("Multiple rows updated at once")
		}
	}
	return
}

func GetCachedStmt(query string) *sqlx.NamedStmt {
	if stmt := preparedStmts[query]; stmt == nil {
		preparedStmts[query] = Prepare(query)
	}
	return preparedStmts[query]
}

func Prepare(query string) *sqlx.NamedStmt {
	if named, err := database.PrepareNamed(query); err != nil {
		log.Panicln(fmt.Errorf("Error: %s Query: %s", query, err))
		return nil
	} else {
		return named
	}
}

func PrepareAllStmts() {
	queries := []string{
		AckNotificationQuery,
	}

	for _, query := range queries {
		preparedStmts[query] = Prepare(query)
	}
}

var (
	AckNotificationQuery = `UPDATE notification SET (ack_at, message_id) = (now(), :message_id) WHERE topic_name = :topic_name RETURNING notification.id`
)
