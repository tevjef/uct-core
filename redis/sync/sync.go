package sync

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"time"
	"uct/redis"
)

type RedisSync struct {
	guid           string
	instanceId     string
	position       int64
	instances      int64
	timeQuantum    time.Duration
	uctRedis       *v1.RedisWrapper
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

func New(uctRedis *v1.RedisWrapper, timeQuantum time.Duration, appId string) *RedisSync {
	// Setup namespaces
	nsSpace = uctRedis.NameSpace + ":sync"
	nsInstances = nsSpace + ":instance"
	nsHealth = nsSpace + ":health"

	rs := &RedisSync{
		guid:           appId,
		timeQuantum:    timeQuantum,
		uctRedis:       uctRedis,
		position:       -1,
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

	return
}

func (rsync *RedisSync) Sync(cancel chan bool) <-chan int {
	c := make(chan int)

	go func() {
		ticker := time.NewTicker(rsync.syncInterval)
		for {
			select {
			case t := <-ticker.C:
				_ = t
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
					if count := rsync.getInstanceCount(); count < rsync.instances && count != 0 {
						rsync.unregisterAll()
					}

					// Store the current number of instances for future reference
					rsync.instances = rsync.getInstanceCount()

					// Calculate the offset given a duration and channel it so that the application update it's offset
					c <- int(rsync.calculateOffset())
				}()
			case <-cancel:
				ticker.Stop()
				close(c)
			}
		}
	}()

	return c

}

// Deletes the list at key `nsInstances`
func (rsync *RedisSync) unregisterAll() {
	rsync.uctRedis.Client.Del(nsInstances)
}

// Pushes this instanceId for this instance on the list nsInstances if the
// instanceId is not already on the list. Saves it index on the list where
// the instanceId was pushed to
func (rsync *RedisSync) registerInstance() {
	if _, err := rsync.uctRedis.RPushNotExist(nsInstances, rsync.instanceId); err != nil {
		log.WithError(err).Panic("failed to claim position in list:", nsInstances)
	}

	rsync.position = rsync.getPosition()

	log.WithFields(log.Fields{"position": rsync.position, "id": rsync.instanceId}).Infoln("makeClaim")
}

// Get the index position on the list where the instance resides
func (rsync *RedisSync) getPosition() int64 {
	if index, err := rsync.uctRedis.Exists(nsInstances, rsync.instanceId); err != nil {
		log.WithError(err).Panic("failed to check if key exists in list:", nsInstances)
		return -1
	} else {
		return index
	}
}

// Get the number of instance that have performed a ping, it finds
// instances by pattern matching the prefix of the instanceId
func (rsync *RedisSync) getInstanceCount() int64 {
	if count, err := rsync.uctRedis.Count(nsHealth + ":*"); err != nil {
		log.WithError(err).Panic("failed to get number of instances")
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
		log.WithError(err).Panic("failed to perform health check for this instance")
	}
}

func (rsync *RedisSync) calculateOffset() int64 {
	time := int64(rsync.timeQuantum.Seconds())
	instances := rsync.getInstanceCount()
	position := rsync.position

	return calculateOffset(time, instances, position)
}

func calculateOffset(time, instances, position int64) int64 {
	n := time * position
	d := instances

	offset := int64((n / d))
	log.WithFields(log.Fields{"offset": offset, "time": time,
		"instances": instances, "position": position}).Debugln("calculateOffset")

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
