package harmony

import (
	"fmt"
	"os"
	"sync"
	"time"
	"uct/common/redis"
	"uct/common/redis/lock"

	log "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
)

type redisSync struct {
	instance       *instance
	uctRedis       *redis.Helper
	syncInterval   time.Duration
	syncExpiration time.Duration
	listMu         *lock.Lock
	countMu        *lock.Lock
	mu             sync.Mutex

	keyspace     string
	instanceList string
	healthSpace  string
}

type instance struct {
	id          string
	pos         int64
	c           int64
	off         time.Duration
	timeQuantum time.Duration
	mu          sync.Mutex
}

const (
	envRedisSyncInterval   = "UCT_REDIS_SYNC_INTERVAL"
	envRedisSyncExpiration = "UCT_REDIS_SYNC_EXPIRATION"
)

type config struct {
	id            string
	interval      time.Duration
	syncFrequency time.Duration
	expiration    time.Duration
}

type option func(*config)

func newSync(helper *redis.Helper, options ...option) *redisSync {
	var config config

	for _, option := range options {
		option(&config)
	}

	if config.expiration == 0 {
		config.expiration = 4 * time.Second
	}

	if env := os.Getenv(envRedisSyncExpiration); env != "" {
		if dur := parseDuration(env); dur > 0 {
			config.expiration = dur
		}
	}

	if config.id == "" {
		config.id = uuid.NewV4().String()
	}

	if config.interval == 0 {
		config.interval = time.Minute
	}

	if config.syncFrequency == 0 {
		config.syncFrequency = 2 * time.Second
	}

	if env := os.Getenv(envRedisSyncInterval); env != "" {
		if dur := parseDuration(env); dur > 0 {
			config.syncFrequency = dur
		}
	}

	// Setup keyspaces
	keyspace := helper.NameSpace + ":sync"
	instanceList := keyspace + ":instance"
	healthSpace := keyspace + ":health"

	rs := &redisSync{
		instance: &instance{
			timeQuantum: config.interval,
			pos:         -1,
			off:         -1,
			id:          healthSpace + ":" + config.id,
		},
		uctRedis:       helper,
		syncInterval:   config.syncFrequency,
		syncExpiration: config.expiration,
		listMu:         lock.NewLock(helper.Client, instanceList+":lock", &lock.LockOptions{}),
		countMu:        lock.NewLock(helper.Client, keyspace+":count:lock", &lock.LockOptions{}),
		keyspace:       keyspace,
		instanceList:   instanceList,
		healthSpace:    healthSpace,
	}

	return rs
}

func (rsync *redisSync) sync(ctx context.Context) <-chan instance {
	instanceConfigChan := make(chan instance)

	go rsync.beginSync(ctx, instanceConfigChan)

	return instanceConfigChan
}

func (rsync *redisSync) beginSync(ctx context.Context, instanceConfig chan<- instance) {

	defer func() {
		log.Warningln("SYNC ENDING!!!!!")
	}()
	// Clean up previous instances
	if keys, err := rsync.uctRedis.FindAll(rsync.healthSpace + ":*"); err != nil {
		log.WithError(err).Fatalln("failed to retrieve all keys during clean up", rsync.instance.id)
	} else if len(keys) == 0 {
		// do nothing
	} else if err := rsync.uctRedis.Client.Del(keys...).Err(); err != nil {
		log.WithError(err).WithField("keys", keys).Fatalln("failed to delete all keys during clean up", rsync.instance.id)
	} else if err := rsync.uctRedis.Client.Del(rsync.instanceList).Err(); err != nil {
		log.WithError(err).Fatalln("failed to delete instance list", rsync.instance.id)
	}

	ticker := time.NewTicker(rsync.syncInterval)

	var lastCount int64

	for {
		select {
		case <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Error("Recovered from panic", r)
					}

				}()

				defer func() {
					// Store the current number of instances for future reference
					lastCount = rsync.instance.count()
				}()

				// A list maintains the number of running instances.
				// Register instance to list if not already exists
				rsync.registerInstance()

				// Get the number of currently alive instances, if its less than the last count and not 0
				// Unregister all instances. They will all register themselves on their next ping
				//log.WithFields(log.Fields{"count": rsync.instance.count(), "last_count": lastCount}).Println()
				if currentCount := rsync.instance.count(); currentCount < lastCount && lastCount != 0 {
					rsync.unregisterAll() /* */
					return
				}

				log.WithFields(log.Fields{
					"offset":    rsync.instance.off.String(),
					"instances": rsync.instance.count(),
					"position":  rsync.instance.position(),
					"id": rsync.instance.id,
					"zlast": lastCount}).Infoln()
				// Send instance
				instanceConfig <- *rsync.instance
			}()
		case <-ctx.Done():
			log.Warningln("CONTEXT DONE!!!!!")

			return
		}
	}
}

func (rsync *redisSync) unregisterAll() {
	rsync.listMu.Lock()
	defer rsync.listMu.Unlock()

	if _, err := rsync.uctRedis.ListSize(rsync.instanceList); err == nil {
		log.Debugln("Deleting list", rsync.instance.id)
		rsync.uctRedis.Client.Del(rsync.instanceList)
	}
}

// Pushes this instance.Id for this instance on the list nsInstances if the
// instance.Id is not already on the list.
func (rsync *redisSync) registerInstance() {
	// Place health marker to say this instance is still alive
	// If "some time" goes by without another ping, this instance
	// is considered to be dead
	rsync.ping()

	if _, err := rsync.listMu.Lock(); err != nil {
		log.WithError(err).Fatalln("failed to aquire lock in registerInstance", rsync.instance.id)
	} else if _, err = rsync.uctRedis.RPushNotExist(rsync.instanceList, rsync.instance.id); err != nil {
		log.WithError(err).Fatalln("failed to claim position in list:", rsync.instance.id)
	} else if err = rsync.listMu.Unlock(); err != nil {
	} else if err = rsync.listMu.Unlock(); err != nil {
		log.WithError(err).Fatalln("failed to release lock in registerInstance", rsync.instance.id)
	} else {
		rsync.updateInstanceCount()
		rsync.updatePosition()
	}

}

// Get the index position on the list where the instance resides
func (rsync *redisSync) updatePosition() {
	rsync.listMu.Lock()
	defer rsync.listMu.Unlock()

	if index, err := rsync.uctRedis.Exists(rsync.instanceList, rsync.instance.id); err != nil {
		log.WithError(err).Fatalln("failed to check if key exists in list:", rsync.instanceList)
	} else {
		rsync.instance.mu.Lock()
		rsync.instance.pos = index
		rsync.updateOffset()
		rsync.instance.mu.Unlock()
	}
}

func (inst *instance) position() int64 {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.pos
}

func (inst *instance) offset() time.Duration {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.off
}

// Get the number of instance that have performed a ping, it finds
// instances by pattern matching the prefix of the instanceId
func (rsync *redisSync) updateInstanceCount() {
	rsync.countMu.Lock()
	defer rsync.countMu.Unlock()

	if count, err := rsync.uctRedis.Count(rsync.healthSpace + ":*"); err != nil {
		log.WithError(err).Fatalln("failed to get number of instances")
	} else {
		rsync.instance.mu.Lock()
		rsync.instance.c = count
		rsync.updateOffset()
		rsync.instance.mu.Unlock()
	}
}

func (inst *instance) count() int64 {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	return inst.c
}

func (rsync *redisSync) ping() {

	rsync.pingWithExpiration(rsync.syncExpiration)
}

// Ping sets its instanceId on the redis.
func (rsync *redisSync) pingWithExpiration(duration time.Duration) {
	if err := rsync.uctRedis.Client.Set(rsync.instance.id, 1, duration).Err(); err != nil {
		log.WithError(err).Fatalln("failed to perform health check for this instance")
	} else if err := rsync.uctRedis.Client.Expire(rsync.instanceList, 5*time.Second).Err(); err != nil {
		log.WithError(err).Fatalln("failed to reset list expiration")
	}
}

func (rsync *redisSync) updateOffset() {
	t := int64(rsync.instance.timeQuantum.Seconds())
	instances := rsync.instance.c
	position := rsync.instance.pos
	rsync.instance.off = time.Duration(calculateOffset(t, instances, position)) * time.Second

	//log.WithFields(log.Fields{
	//	"offset":      rsync.instance.off.Seconds(),
	//	"interval":    rsync.instance.timeQuantum.Seconds(),
	//	"instances":   rsync.instance.c,
	//	"position":    rsync.instance.pos,
	//	"instance_id": rsync.instance.id}).Debugln(rsync.instance.id)

}

func calculateOffset(interval, instances, position int64) int64 {

	offset := (interval * position) / instances

	if offset > interval {
		log.WithFields(log.Fields{
			"offset":    offset,
			"interval":  interval,
			"instances": instances,
			"position":  position}).Warnln("offset is more than interval")
		offset = interval
	}

	if offset < 0 || position < 0 {
		return 0
	}

	return offset
}

func parseDuration(duration string) time.Duration {
	if dur, err := time.ParseDuration(duration); err != nil {
		log.WithError(fmt.Errorf("error parseing duration %s", duration)).Errorln()
		return -1
	} else {
		return dur
	}
}
