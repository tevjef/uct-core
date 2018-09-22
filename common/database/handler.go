package database

import (
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"syscall"

	"context"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tevjef/uct-backend/common/try"
)

type Handler interface {
	Insert(query string, data interface{}) (id int64)
	Update(query string, data interface{}) (id int64)
	Upsert(insertQuery, updateQuery string, data interface{}) (id int64)
	Exists(query string, data interface{}) (id int64)
	Select(query string, dest interface{}, args interface{}) error
	Get(query string, dest interface{}, args interface{}) error
	Stats() *Stats
	ResetStats()
}

type handlerImpl struct {
	appName    string
	database   *sqlx.DB
	statements map[string]*sqlx.NamedStmt
	stats      *Stats
}

type Stats struct {
	Insertions int64
	Updates    int64
	Upserts    int64
	Exists     int64
}

func NewHandler(appName string, database *sqlx.DB, queries []string) Handler {
	impl := handlerImpl{
		appName:    appName,
		database:   database,
		statements: map[string]*sqlx.NamedStmt{},
		stats:      &Stats{},
	}

	impl.prepareStatements(queries)

	return impl
}

// Setter defines a context that enables setting values.
type Setter interface {
	Set(string, interface{})
}

const key = "databasehandler"

func FromContext(ctx context.Context) Handler {
	return ctx.Value(key).(Handler)
}

func ToContext(s Setter, h Handler) {
	s.Set(key, h)
}

func (db handlerImpl) invoke(query string, data interface{}, fields log.Fields) (id int64, err error) {
	typeName := fmt.Sprintf("%T", data)

	err = try.Do(func(attempt int) (retry bool, err error) {
		if rows, err := db.getCachedStmt(query).Queryx(data); err != nil {
			return isRetriable(err)
		} else {
			count := 0
			for rows.Next() {
				count++

				if err = rows.Scan(&id); err != nil {
					log.WithFields(log.Fields{"type": typeName, "data": data}).WithFields(fields).Panicln(err)
				}
				rows.Close()
				log.WithFields(log.Fields{"type": typeName, "id": id}).WithFields(fields).Debugln()
			}
			if count > 1 {
				log.WithFields(log.Fields{"type": typeName, "data": data}).WithFields(fields).Warningln("multiple rows affected")
			}
			return false, nil
		}
	})

	if err != nil {
		log.WithError(err).Fatalln()
	}

	return
}

func isRetriable(err error) (bool, error) {
	if isConnectionError(err) {
		return true, errors.New("connection error")
	}

	return false, err
}

func (db handlerImpl) Insert(query string, data interface{}) (id int64) {
	var err error

	id, err = db.invoke(query, data, log.Fields{db.appName + "_op": "insert"})
	if err != nil {
		log.WithError(err).Fatalln()
	}

	atomic.AddInt64(&db.stats.Insertions, 1)
	return id
}

func (db handlerImpl) Update(query string, data interface{}) (id int64) {
	var err error

	id, err = db.invoke(query, data, log.Fields{db.appName + "_op": "update"})
	if err != nil {
		log.WithError(err).Fatalln()
	}

	atomic.AddInt64(&db.stats.Updates, 1)
	return id
}

func (db handlerImpl) Exists(query string, data interface{}) (id int64) {
	var err error

	id, err = db.invoke(query, data, log.Fields{db.appName + "_op": "exists"})
	if err != nil {
		log.WithError(err).Fatalln()
	}

	atomic.AddInt64(&db.stats.Exists, 1)
	return
}

func (db handlerImpl) Upsert(insertQuery, updateQuery string, data interface{}) (id int64) {
	if id = db.Update(updateQuery, data); id != 0 {
	} else if id == 0 {
		id = db.Insert(insertQuery, data)
	}

	atomic.AddInt64(&db.stats.Upserts, 1)
	return
}

func (db handlerImpl) Select(query string, dest interface{}, args interface{}) error {
	return db.getCachedStmt(query).Select(dest, args)
}

func (db handlerImpl) Get(query string, dest interface{}, args interface{}) error {
	return db.getCachedStmt(query).Get(dest, args)
}

func (db handlerImpl) Stats() *Stats {
	return &Stats{
		Insertions: atomic.LoadInt64(&db.stats.Insertions),
		Updates:    atomic.LoadInt64(&db.stats.Updates),
		Exists:     atomic.LoadInt64(&db.stats.Exists),
		Upserts:    atomic.LoadInt64(&db.stats.Upserts),
	}
}

func (db handlerImpl) ResetStats() {
	atomic.SwapInt64(&db.stats.Insertions, 0)
	atomic.SwapInt64(&db.stats.Updates, 0)
	atomic.SwapInt64(&db.stats.Exists, 0)
	atomic.SwapInt64(&db.stats.Upserts, 0)
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

func (db handlerImpl) getCachedStmt(key string) *sqlx.NamedStmt {
	return db.statements[key]
}

func (db handlerImpl) prepare(query string) *sqlx.NamedStmt {
	if named, err := db.database.PrepareNamed(query); err != nil {
		log.WithError(err).Fatalln("failed to prepare query:", query)
		return nil
	} else {
		return named
	}
}

func (db handlerImpl) prepareStatements(queries []string) {
	for _, query := range queries {
		db.statements[query] = db.prepare(query)
	}
}
