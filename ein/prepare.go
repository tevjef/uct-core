package main

import (
	"fmt"
	"net"
	"os"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
)

type DatabaseHandler interface {
	insert(query string, data interface{}) (id int64)
	update(query string, data interface{}) (id int64)
	upsert(insertQuery, updateQuery string, data interface{}) (id int64)
	exists(query string, data interface{}) (id int64)
	prepareStatements()
}

type DatabaseHandlerImpl struct {
	database   *sqlx.DB
	statements map[string]*sqlx.NamedStmt
}

func (db DatabaseHandlerImpl) insert(query string, data interface{}) (id int64) {
	// model.TimeTrack(time.Now(), "insert")

	insertionsCh <- 1
	typeName := fmt.Sprintf("%T", data)
	if rows, err := db.getCachedStmt(query).Queryx(data); err != nil {
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

func (db DatabaseHandlerImpl) update(query string, data interface{}) (id int64) {
	// model.TimeTrack(time.Now(), "update")
	typeName := fmt.Sprintf("%T", data)

	for i := 0; i < 5; i++ {
		if rows, err := db.getCachedStmt(query).Queryx(data); err != nil {
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

func (db DatabaseHandlerImpl) upsert(insertQuery, updateQuery string, data interface{}) (id int64) {
	// model.TimeTrack(time.Now(), "upsert")
	upsertsCh <- 1
	if id = db.update(updateQuery, data); id != 0 {
	} else if id == 0 {
		id = db.insert(insertQuery, data)
	}
	return
}

func (db DatabaseHandlerImpl) exists(query string, data interface{}) (id int64) {
	typeName := fmt.Sprintf("%T", data)
	existentialCh <- 1

	if rows, err := db.getCachedStmt(query).Queryx(data); err != nil {
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

func (db DatabaseHandlerImpl) getCachedStmt(key string) *sqlx.NamedStmt {
	return db.statements[key]
}

func (db DatabaseHandlerImpl) prepare(query string) *sqlx.NamedStmt {
	if named, err := db.database.PrepareNamed(query); err != nil {
		log.WithError(err).Fatalln("failed to prepare query:", query)
		return nil
	} else {
		return named
	}
}

func (db DatabaseHandlerImpl) prepareStatements() {
	queries := []string{UniversityInsertQuery,
		UniversityUpdateQuery,
		SemesterInsertQuery,
		SemesterUpdateQuery,
		SubjectExistQuery,
		SubjectInsertQuery,
		SubjectUpdateQuery,
		CourseUpdateQuery,
		CourseExistQuery,
		CourseInsertQuery,
		SectionInsertQuery,
		SectionUpdateQuery,
		MeetingUpdateQuery,
		MeetingInsertQuery,
		MeetingExistQuery,
		InstructorExistQuery,
		InstructorUpdateQuery,
		InstructorInsertQuery,
		BookUpdateQuery,
		BookInsertQuery,
		RegistrationUpdateQuery,
		RegistrationInsertQuery,
		MetaUniExistQuery,
		MetaUniUpdateQuery,
		MetaUniInsertQuery,
		MetaSubjectExistQuery,
		MetaSubjectUpdateQuery,
		MetaSubjectInsertQuery,
		MetaCourseExistQuery,
		MetaCourseUpdateQuery,
		MetaCourseInsertQuery,
		MetaSectionExistQuery,
		MetaSectionInsertQuery,
		MetaSectionUpdateQuery,
		MetaSectionExistQuery,
		MetaMeetingInsertQuery,
		MetaMeetingUpdateQuery,
		SerialSubjectUpdateQuery,
		SerialCourseUpdateQuery,
		SerialSectionUpdateQuery}

	for _, query := range queries {
		db.statements[query] = db.prepare(query)
	}
}
