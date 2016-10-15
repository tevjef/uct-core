package harmony

import (
	"fmt"
	"os"
	"time"
	"uct/redis"

	log "github.com/Sirupsen/logrus"
)

type RedisSync struct {
	guid           string
	instanceId     string
	position       int64
	Instances      int64
	offset         time.Duration
	timeQuantum    time.Duration
	uctRedis       *redishelper.RedisWrapper
	syncInterval   time.Duration
	syncExpiration time.Duration
}

var (
	nsSpace     string
	nsInstances string
	nsHealth    string
)

const (
	envRedisSyncInterval   = "UCT_REDIS_SYNC_INTERVAL"
	envRedisSyncExpiration = "UCT_REDIS_SYNC_EXPIRATION"
)

func New(uctRedis *redishelper.RedisWrapper, timeQuantum time.Duration, appId string) *RedisSync {
	// Setup namespaces
	nsSpace = uctRedis.NameSpace + ":sync"
	nsInstances = nsSpace + ":instance"
	nsHealth = nsSpace + ":health"

	rs := &RedisSync{
		guid:           appId,
		timeQuantum:    timeQuantum,
		uctRedis:       uctRedis,
		position:       -1,
		offset:         -1,
		instanceId:     nsHealth + ":" + appId,
		syncInterval:   2 * time.Second,
		syncExpiration: 4 * time.Second,
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

func (rsync *RedisSync) Sync(cancel chan bool) <-chan time.Duration {
	offsetChan := make(chan time.Duration)

	go rsync.beginSync(offsetChan, cancel)

	return offsetChan

}

func (rsync *RedisSync) beginSync(offset chan<- time.Duration, cancel <-chan bool) {
	ticker := time.NewTicker(rsync.syncInterval)
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

				// Get the number of currently alive instances, if it's less that the last count but not 0
				// Unregister all instances all instances. They will all reorder themselves on their next ping
				instanceCount := rsync.getInstanceCount()
				if instanceCount < rsync.Instances && instanceCount != 0 {
					rsync.unregisterAll()
				}

				// Calculate the offset given a duration and channel it so that the application update it's offset
				newOffset := time.Duration(rsync.calculateOffset()) * time.Second

				// Send new offset on scale up and down???
				if rsync.offset != newOffset {
					offset <- newOffset
				}

				// Store the current number of instances for future reference
				rsync.Instances = rsync.getInstanceCount()

				rsync.offset = newOffset
			}()
		case <-cancel:
			ticker.Stop()
			close(offset)
		}
	}
}

// Deletes the list at key `nsInstances`
func (rsync *RedisSync) unregisterAll() {
	rsync.uctRedis.Client.Del(nsInstances)
}

// Pushes this instanceId for this instance on the list nsInstances if the
// instanceId is not already on the list. Saves it index on the list where
// the instanceId was pushed to
func (rsync *RedisSync) registerInstance() {
	// Reset list expiration
	rsync.uctRedis.Client.Expire(nsInstances, rsync.syncExpiration)
	if _, err := rsync.uctRedis.RPushNotExist(nsInstances, rsync.instanceId); err != nil {
		log.WithError(err).Fatalln("failed to claim position in list:", nsInstances)
	}

	rsync.position = rsync.getPosition()
}

// Get the index position on the list where the instance resides
func (rsync *RedisSync) getPosition() int64 {
	if index, err := rsync.uctRedis.Exists(nsInstances, rsync.instanceId); err != nil {
		log.WithError(err).Fatalln("failed to check if key exists in list:", nsInstances)
		return -1
	} else {
		return index
	}
}

// Get the number of instance that have performed a ping, it finds
// instances by pattern matching the prefix of the instanceId
func (rsync *RedisSync) getInstanceCount() int64 {
	if count, err := rsync.uctRedis.Count(nsHealth + ":*"); err != nil {
		log.WithError(err).Fatalln("failed to get number of instances")
	} else {
		return count
	}
	return -1
}

func (rsync *RedisSync) ping() {
	rsync.pingWithExpiration(rsync.syncExpiration)
}

// Ping sets its instanceId on the redis.
func (rsync *RedisSync) pingWithExpiration(duration time.Duration) {
	if _, err := rsync.uctRedis.Client.Set(rsync.instanceId, 1, duration).Result(); err != nil {
		log.WithError(err).Fatalln("failed to perform health check for this instance")
	}
}

func (rsync *RedisSync) calculateOffset() int64 {
	time := int64(rsync.timeQuantum.Seconds())
	instances := rsync.getInstanceCount()
	position := rsync.position

	return calculateOffset(time, instances, position)
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
