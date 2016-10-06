package sync

import (
	uct "uct/common"
	"testing"
	"uct/redis"
	"time"
	"github.com/stretchr/testify/assert"

	"github.com/satori/go.uuid"
	_ "github.com/Sirupsen/logrus"
	"strconv"
	"log"
	"uct/common/conf"
)

func setupClient() *v1.RedisWrapper {
	c := conf.Config{}
	c.Redis.Host = "localhost:32768"
	c.Redis.Db = 0
	c.Redis.Password = ""
	return v1.New(c, "snanitycheck")
}

func setup(timeQuantum time.Duration, appId string) *RedisSync {
	return New(setupClient(), timeQuantum, appId)
}

func teardown(rsync *RedisSync) {
	rsync.uctRedis.Client.FlushAll()
}

func TestClientConnection(t *testing.T) {
	rsync := setup(1 * time.Minute, uuid.NewV4().String())

	rsync.ping()

	teardown(rsync)
}

func TestRedisSync_ClaimPosition_Same_Result_Same_Id(t *testing.T) {
	rsync := setup(0, "test")
	rsync.registerInstance()
	expected := rsync.position
	rsync.registerInstance()
	result := rsync.position

	assert.Equal(t, expected, result)

	teardown(rsync)
}

func TestRedisSync_ClaimPosition_Different_Result_Different_Id(t *testing.T) {
	rsync := setup(0, "test")

	rsync.registerInstance()
	expected := rsync.position + 1

	rsync = setup(0, "test2")
	rsync.registerInstance()

	result := rsync.position

	assert.Equal(t, expected, result)

	teardown(rsync)
}

func TestRedisSync_calculateOffset_client_1(t *testing.T) {
	rsync := setup(1 * time.Minute, "test")

	rsync.ping()
	rsync.registerInstance()

	expected := int64(0)
	result := rsync.calculateOffset()

	assert.Equal(t, expected, result)

	teardown(rsync)
}

func TestRedisSync_calculateOffset_client_3(t *testing.T) {
	rsync := setup(1 * time.Minute, "test1")

	rsync.ping()
	rsync.registerInstance()

	rsync = setup(1 * time.Minute, "test2")
	rsync.ping()
	rsync.registerInstance()

	rsync = setup(1 * time.Minute, "test3")
	rsync.ping()
	rsync.registerInstance()

	expected := int64(40)
	result := rsync.calculateOffset()

	assert.Equal(t, expected, result)

	teardown(rsync)
}

func TestRedisSync_calculateOffset_client_n(t *testing.T) {
	N := 59
	for i := 0; i < N; i++ {
		rsync := setup(1 * time.Minute, "test" + strconv.Itoa(i))
		rsync.ping()
		rsync.registerInstance()
	}

	rsync := setup(1 * time.Minute, "last")
	rsync.ping()
	rsync.registerInstance()



	expected := int64(59)

	result := rsync.calculateOffset()

	assert.Equal(t, expected, result)

	teardown(rsync)
}

func TestRedisSync_Sync(t *testing.T) {

	go func() {
		time.Sleep(0 * time.Second)

		rsync := setup(1 * time.Minute, "test1")

		for range rsync.Sync(make(chan bool)) {

		}
	}()

	go func() {
		time.Sleep(3 * time.Second)

		rsync := setup(1 * time.Minute, "test2")

		for range rsync.Sync(make(chan bool)) {

		}
	}()

	go func() {
		time.Sleep(5 * time.Second)

		rsync := setup(1 * time.Minute, "test3")
		channel := make(chan bool)
		go func() {
			time.Sleep(10 * time.Second)
			channel <- true
			log.Println("BOOM!!! TEST3")
		}()

		for range rsync.Sync(channel) {

		}

	}()


	go func() {
		time.Sleep(9 * time.Second)

		rsync := setup(1 * time.Minute, "test4")
		log.Println("TEST4")

		channel := make(chan bool)

		for range rsync.Sync(channel) {

		}
	}()

	select {}
}