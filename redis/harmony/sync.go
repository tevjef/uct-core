package harmony

import (
	"fmt"
	"os"
	"time"
	"uct/redis"

	log "github.com/Sirupsen/logrus"
)

type RedisSync struct {
	instance       Instance
	uctRedis       *redishelper.RedisWrapper
	syncInterval   time.Duration
	syncExpiration time.Duration
}

type Instance struct {
	guid        string
	id          string
	position    int64
	count       int64
	offset      time.Duration
	timeQuantum time.Duration
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
		instance: Instance{
			guid:        appId,
			timeQuantum: timeQuantum,
			position:    -1,
			offset:      -1,
			id:          nsHealth + ":" + appId,
		},
		uctRedis:       uctRedis,
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

func (rsync *RedisSync) Sync(cancel chan bool) <-chan Instance {
	instanceConfigChan := make(chan Instance)

	go rsync.beginSync(instanceConfigChan, cancel)

	return instanceConfigChan
}

func (rsync *RedisSync) beginSync(instanceConfig chan<- Instance, cancel <-chan bool) {
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
				if instanceCount < rsync.instance.count && instanceCount != 0 {
					rsync.unregisterAll()
				}

				// Store the current number of instances for future reference
				rsync.instance.count = rsync.getInstanceCount()

				// Calculate the offset given a duration and channel it so that the application update it's offset
				oldOffset := rsync.instance.offset
				rsync.instance.offset = time.Duration(rsync.calculateOffset()) * time.Second

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
	rsync.uctRedis.Client.Del(nsInstances)
}

// Pushes this instanceId for this instance on the list nsInstances if the
// instanceId is not already on the list. Saves it index on the list where
// the instanceId was pushed to
func (rsync *RedisSync) registerInstance() {
	// Reset list expiration
	rsync.uctRedis.Client.Expire(nsInstances, rsync.syncExpiration)
	if _, err := rsync.uctRedis.RPushNotExist(nsInstances, rsync.instance.id); err != nil {
		log.WithError(err).Fatalln("failed to claim position in list:", nsInstances)
	}

	rsync.instance.position = rsync.getPosition()
}

// Get the index position on the list where the instance resides
func (rsync *RedisSync) getPosition() int64 {
	if index, err := rsync.uctRedis.Exists(nsInstances, rsync.instance.id); err != nil {
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
	if _, err := rsync.uctRedis.Client.Set(rsync.instance.id, 1, duration).Result(); err != nil {
		log.WithError(err).Fatalln("failed to perform health check for this instance")
	}
}

func (rsync *RedisSync) calculateOffset() int64 {
	time := int64(rsync.instance.timeQuantum.Seconds())
	instances := rsync.getInstanceCount()
	position := rsync.instance.position

	return calculateOffset(time, instances, position)
}

func calculateOffset(interval, instances, position int64) int64 {
	n := interval * position
	d := instances

	offset := int64((n / d))

	if offset > interval {
		log.WithFields(log.Fields{"offset": offset, "interval": interval, "instances": instances, "position": position}).Warnln("offset is more than interval")
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
