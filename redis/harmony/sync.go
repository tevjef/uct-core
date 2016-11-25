package harmony

import (
	"fmt"
	"os"
	"time"
	"uct/redis"

	log "github.com/Sirupsen/logrus"
	"sync"
	"uct/redis/lock"
)

type RedisSync struct {
	instance       Instance
	uctRedis       *redishelper.RedisWrapper
	syncInterval   time.Duration
	syncExpiration time.Duration
	listMu         lock.Lock
	countMu        lock.Lock
	mu             sync.Mutex

	nsSpace     string
	nsInstances string
	nsHealth    string
}

type Instance struct {
	id          string
	position    int64
	count       int64
	offset      time.Duration
	timeQuantum time.Duration
	mu          sync.Mutex
}

const (
	envRedisSyncInterval   = "UCT_REDIS_SYNC_INTERVAL"
	envRedisSyncExpiration = "UCT_REDIS_SYNC_EXPIRATION"
)

// number of instances
// -- by number of keys with the same id*
// -
// determine position
// -- push unto list
func New(uctRedis *redishelper.RedisWrapper, timeQuantum time.Duration, appId string) *RedisSync {
	// Setup namespaces
	nsSpace := uctRedis.NameSpace + ":sync"
	nsInstances := nsSpace + ":instance"
	nsHealth := nsSpace + ":health"

	rs := &RedisSync{
		instance: Instance{
			timeQuantum: timeQuantum,
			position:    -1,
			offset:      -1,
			id:          nsHealth + ":" + appId,
		},
		uctRedis:       uctRedis,
		syncInterval:   2 * time.Second,
		syncExpiration: 4 * time.Second,
		nsSpace:        nsSpace,
		nsInstances:    nsInstances,
		nsHealth:       nsHealth,
	}

	if env := os.Getenv(envRedisSyncInterval); env != "" {
		if dur := parseDuration(env); dur > 0 {
			rs.syncInterval = dur
		}
	}

	if env := os.Getenv(envRedisSyncExpiration); env != "" {
		if dur := parseDuration(env); dur > 0 {
			rs.syncExpiration = dur
		}
	}

	return rs
}

func (rsync *RedisSync) Sync(cancel chan bool) <-chan Instance {
	instanceConfigChan := make(chan Instance)

	go rsync.beginSync(instanceConfigChan, cancel)

	return instanceConfigChan
}

func (rsync *RedisSync) beginSync(instanceConfig chan<- Instance, cancel <-chan bool) {
	ticker := time.NewTicker(rsync.syncInterval)

	//var lastOffset int64
	//var lastPosition int64
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

				// Place health marker to say this instance is still alive
				// If n seconds goes by without another ping, this instance
				// is considered to be dead
				rsync.ping()

				// A list maintains the number of running instances.
				// Register instance to list if not already exists
				rsync.registerInstance()

				// Get the number of currently alive instances, if its less than the last count and not 0
				// Unregister all instances. They will all register themselves on their next ping
				if rsync.instance.count() < lastCount && lastCount != 0 {
					rsync.unregisterAll()
				}

				// Store the current number of instances for future reference
				lastCount = rsync.instance.count()

				// Calculate the offset given a duration and channel it so that the application update it's offset
				oldOffset := rsync.instance.offset
				rsync.instance.offset = time.Duration(rsync.updateOffset()) * time.Second

				// Send new offset on scale up and down???
				if rsync.instance.offset != oldOffset {
					log.Infoln("Sending new offset")
					instanceConfig <- rsync.instance
				}

			}()
		case <-cancel:
			ticker.Stop()
			close(instanceConfig)
		}
	}
}

// Deletes the list at key `nsInstances`
func (rsync *RedisSync) unregisterAll() {
	rsync.listMu.Lock()
	rsync.uctRedis.Client.Del(rsync.nsInstances)
	rsync.listMu.Unlock()
}

// Pushes this instanceId for this instance on the list nsInstances if the
// instanceId is not already on the list. Saves it index on the list where
// the instanceId was pushed to
func (rsync *RedisSync) registerInstance() {
	// Reset list expiration
	rsync.uctRedis.Client.Expire(rsync.nsInstances, rsync.syncExpiration)

	rsync.listMu.Lock()
	if _, err := rsync.uctRedis.RPushNotExist(rsync.nsInstances, rsync.instance.id); err != nil {
		log.WithError(err).Fatalln("failed to claim position in list:", rsync.nsInstances)
	}
	rsync.listMu.Unlock()

	rsync.updatePosition()
}

// Get the index position on the list where the instance resides
func (rsync *RedisSync) updatePosition() {
	rsync.listMu.Lock()
	defer rsync.listMu.Unlock()

	if index, err := rsync.uctRedis.Exists(rsync.nsInstances, rsync.instance.id); err != nil {
		log.WithError(err).Fatalln("failed to check if key exists in list:", rsync.nsInstances)
	} else {
		rsync.instance.mu.Lock()
		rsync.instance.position = index
		rsync.instance.mu.Unlock()
	}
}

func (inst *Instance) position() int64 {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.position
}

// Get the number of instance that have performed a ping, it finds
// instances by pattern matching the prefix of the instanceId
func (rsync *RedisSync) updateInstanceCount() {
	if count, err := rsync.uctRedis.Count(rsync.nsHealth + ":*"); err != nil {
		log.WithError(err).Fatalln("failed to get number of instances")
	} else {
		rsync.instance.mu.Lock()
		rsync.instance.count = count
		rsync.instance.mu.Unlock()
	}
}

func (inst *Instance) count() int64 {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	return inst.count
}

func (rsync *RedisSync) ping() {
	rsync.pingWithExpiration(rsync.syncExpiration)
}

// Ping sets its instanceId on the redis.
func (rsync *RedisSync) pingWithExpiration(duration time.Duration) {
	if _, err := rsync.uctRedis.Client.Set(rsync.instance.id, 1, duration).Result(); err != nil {
		log.WithError(err).Fatalln("failed to perform health check for this instance")
	}
}

func (rsync *RedisSync) updateOffset() int64 {
	time := int64(rsync.instance.timeQuantum.Seconds())
	rsync.instance.mu.Lock()
	instances := rsync.instance.count
	position := rsync.instance.position

	rsync.instance.offset = calculateOffset(time, instances, position)
	rsync.instance.mu.Unlock()

	return
}

func calculateOffset(interval, instances, position int64) int64 {
	n := interval * position
	d := instances

	offset := int64((n / d))

	if offset > interval {
		log.WithFields(log.Fields{"offset": offset, "interval": interval}).Warnln("offset is more than interval")
		offset = interval
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
