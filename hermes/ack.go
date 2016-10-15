package main

import (
	"github.com/jmoiron/sqlx"
	log "github.com/Sirupsen/logrus"
	"fmt"
	"github.com/pkg/errors"
)


func acknowledgeNotification(notificationId, messageId int64) (id int64) {
	args := map[string]interface{}{"notification_id": notificationId, "message_id": messageId}

	if rows, err := getCachedStmt(AckNotificationQuery).Queryx(args); err != nil {
		log.WithFields(log.Fields{"args": args}).Panic(err)
	} else {
		count := 0
		for rows.Next() {
			count++
			if err = rows.Scan(&id); err != nil {
				log.WithFields(log.Fields{"args": args}).Panic(err)
			}
			rows.Close()
		}
		if count > 1 {
			log.WithFields(log.Fields{"args": args}).Panic("Multiple rows updated at once")
		}
		if id == 0 {
			log.WithFields(log.Fields{"args": args}).Panic(errors.New("Id is 0 retuinig fro updating notification in database"))
		}
	}
	return
}

func getCachedStmt(query string) *sqlx.NamedStmt {
	if stmt := preparedStmts[query]; stmt == nil {
		preparedStmts[query] = prepare(query)
	}
	return preparedStmts[query]
}

func prepare(query string) *sqlx.NamedStmt {
	if named, err := database.PrepareNamed(query); err != nil {
		log.Panicln(fmt.Errorf("Error: %s Query: %s", query, err))
		return nil
	} else {
		return named
	}
}

func prepareAllStmts() {
	queries := []string{
		AckNotificationQuery,
	}

	for _, query := range queries {
		preparedStmts[query] = prepare(query)
	}
}

var (
	AckNotificationQuery = `UPDATE notification SET (ack_at, message_id) = (now(), :message_id) WHERE id = :notification_id RETURNING notification.id`
)