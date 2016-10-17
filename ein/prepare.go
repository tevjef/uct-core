package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"net"
	"os"
	"syscall"
)

type DatabaseHandler interface {
	insert(query string, data interface{}) (id int64)
	update(query string, data interface{}) (id int64)
	upsert(insertQuery, updateQuery string, data interface{}) (id int64)
	exists(query string, data interface{}) (id int64)
}

type DatabaseHandlerImpl struct {
	Database *sqlx.DB
}

func (dbHandler DatabaseHandlerImpl) insert(query string, data interface{}) (id int64) {
	// model.TimeTrack(time.Now(), "insert")

	insertionsCh <- 1
	typeName := fmt.Sprintf("%T", data)
	if rows, err := GetCachedStmt(query).Queryx(data); err != nil {
		log.WithFields(log.Fields{"ein_op": "Insert", "type": typeName, "data": data}).Panic(err)
	} else {
		for rows.Next() {
			if err = rows.Scan(&id); err != nil {
				log.WithFields(log.Fields{"ein_op": "Insert", "type": typeName, "data": data}).Panic(err)
			}
			rows.Close()
			log.WithFields(log.Fields{"ein_op": "Insert", "type": typeName, "id": id}).Debug()
		}
	}
	return id
}

func (dbHandler DatabaseHandlerImpl) update(query string, data interface{}) (id int64) {
	// model.TimeTrack(time.Now(), "update")
	typeName := fmt.Sprintf("%T", data)

	for i := 0; i < 5; i++ {
		if rows, err := GetCachedStmt(query).Queryx(data); err != nil {
			if isConnectionError(err) {
				log.Errorf("Retry %d after error %s", i, err)
				continue
			} else {
				log.Panicln(err)
			}
		} else {
			count := 0
			for rows.Next() {
				count++

				if err = rows.Scan(&id); err != nil {
					log.WithFields(log.Fields{"ein_op": "Update", "type": typeName, "data": data}).Panic(err)
				}
				rows.Close()
				log.WithFields(log.Fields{"ein_op": "Update", "type": typeName, "id": id}).Debug()
			}
			if count > 1 {
				log.WithFields(log.Fields{"ein_op": "Update", "type": typeName, "data": data}).Panic("Multiple rows updated at once")
			}

			break
		}
	}

	updatesCh <- 1
	return id
}

func (dbHandler DatabaseHandlerImpl) upsert(insertQuery, updateQuery string, data interface{}) (id int64) {
	// model.TimeTrack(time.Now(), "upsert")
	upsertsCh <- 1
	if id = dbHandler.update(updateQuery, data); id != 0 {
	} else if id == 0 {
		id = dbHandler.insert(insertQuery, data)
	}
	return
}

func (dbHandler DatabaseHandlerImpl) exists(query string, data interface{}) (id int64) {
	typeName := fmt.Sprintf("%T", data)
	existentialCh <- 1

	if rows, err := GetCachedStmt(query).Queryx(data); err != nil {
		log.WithFields(log.Fields{"ein_op": "Exists", "type": typeName, "data": data}).Panic(err)
	} else {
		count := 0
		for rows.Next() {
			count++
			if err = rows.Scan(&id); err != nil {
				log.WithFields(log.Fields{"ein_op": "Exists", "type": typeName, "data": data}).Panic(err)
			}
			log.WithFields(log.Fields{"ein_op": "Exists", "type": typeName, "id": id}).Debug()
		}
		if count > 1 {
			log.WithFields(log.Fields{"ein_op": "Exists", "type": typeName, "data": data}).Panic("Multple rows exists")
		}
	}

	return
}

func isConnectionError(err error) bool {
	if operr, ok := err.(*net.OpError); ok {
		if syserr, ok := operr.Err.(*os.SyscallError); ok {
			if errNo := syserr.Err; errNo == syscall.ECONNRESET || errNo == syscall.EPIPE || errNo == syscall.EPROTOTYPE {
				return true
			}
		}
	}
	return false
}
