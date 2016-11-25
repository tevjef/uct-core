package harmony

import (
	"testing"
	"time"
	"uct/redis"

	"github.com/stretchr/testify/assert"

	"strconv"
	"uct/common/conf"

	_ "github.com/Sirupsen/logrus"
	log "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
)

func setupClient() *redishelper.RedisWrapper {
	c := conf.Config{}
	c.Redis.Host = "redis"
	c.Redis.Port = "6379"
	c.Redis.Db = 0
	c.Redis.Password = ""
	return redishelper.New(c, "snanitycheck")
}

func setup(timeQuantum time.Duration, appId string) *RedisSync {
	return New(setupClient(), timeQuantum, appId)
}

func teardown(rsync *RedisSync) {
	rsync.uctRedis.Client.FlushAll()
}

func TestClientConnection(t *testing.T) {
	rsync := setup(1*time.Minute, uuid.NewV4().String())

	rsync.ping()

	teardown(rsync)
}

func TestRedisSync_ClaimPosition_Same_Result_Same_Id(t *testing.T) {
	rsync := setup(0, "test")
	rsync.registerInstance()
	expected := rsync.instance.position
	rsync.registerInstance()
	result := rsync.instance.position

	assert.Equal(t, expected, result)

	teardown(rsync)
}

func TestRedisSync_ClaimPosition_Different_Result_Different_Id(t *testing.T) {
	rsync := setup(0, "test")

	rsync.registerInstance()
	expected := rsync.instance.position + 1

	rsync = setup(0, "test2")
	rsync.registerInstance()

	result := rsync.instance.position

	assert.Equal(t, expected, result)

	teardown(rsync)
}

func TestRedisSync_calculateOffset_client_1(t *testing.T) {
	rsync := setup(1*time.Minute, "test")

	rsync.ping()
	rsync.registerInstance()

	expected := int64(0)
	result := rsync.updateOffset()

	assert.Equal(t, expected, result)

	teardown(rsync)
}

func TestRedisSync_calculateOffset_client_3(t *testing.T) {
	rsync := setup(1*time.Minute, "test1")

	rsync.ping()
	rsync.registerInstance()

	rsync = setup(1*time.Minute, "test2")
	rsync.ping()
	rsync.registerInstance()

	rsync = setup(1*time.Minute, "test3")
	rsync.ping()
	rsync.registerInstance()

	expected := int64(40)
	result := rsync.updateOffset()

	assert.Equal(t, expected, result)

	teardown(rsync)
}

func TestRedisSync_calculateOffset_client_n(t *testing.T) {
	N := 59
	for i := 0; i < N; i++ {
		rsync := setup(1*time.Minute, "test"+strconv.Itoa(i))
		rsync.ping()
		rsync.registerInstance()
	}

	rsync := setup(1*time.Minute, "last")
	rsync.ping()
	rsync.registerInstance()

	expected := int64(59)

	result := rsync.updateOffset()

	assert.Equal(t, expected, result)

	teardown(rsync)
}

func TestRedisSync_Sync(t *testing.T) {

	go func() {
		time.Sleep(0 * time.Second)

		rsync := setup(1*time.Minute, "test1")

		for instance := range rsync.Sync(make(chan bool)) {
			log.WithFields(log.Fields{"offset": instance.offset.Seconds(), "instances": rsync.instance.count, "position": rsync.instance.count}).Println("test1")
		}
	}()

	go func() {
		time.Sleep(3 * time.Second)

		rsync := setup(1*time.Minute, "test2")

		for instance := range rsync.Sync(make(chan bool)) {
			log.WithFields(log.Fields{"offset": instance.offset.Seconds(), "instances": rsync.instance.count, "position": rsync.instance.count}).Println("test2")
		}
	}()

	go func() {
		time.Sleep(5 * time.Second)

		rsync := setup(1*time.Minute, "test3")
		channel := make(chan bool)
		go func() {
			time.Sleep(10 * time.Second)
			channel <- true
			log.Println("BOOM!!! TEST3")
		}()

		for instance := range rsync.Sync(channel) {
			log.WithFields(log.Fields{"offset": instance.offset.Seconds(), "instances": rsync.instance.count, "position": rsync.instance.count}).Println("test3")
		}

	}()

	go func() {
		time.Sleep(9 * time.Second)

		rsync := setup(1*time.Minute, "test4")

		channel := make(chan bool)

		for instance := range rsync.Sync(channel) {
			log.WithFields(log.Fields{"offset": instance.offset.Seconds(), "instances": rsync.instance.count, "position": rsync.instance.count}).Println("test4")
		}
	}()

	select {}
}

