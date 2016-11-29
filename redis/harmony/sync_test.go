package harmony

import (
	"strconv"
	"testing"
	"time"
	"uct/common/conf"
	"uct/redis"

	"sort"

	_ "github.com/Sirupsen/logrus"
	log "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func getClient() *redis.Helper {
	c := conf.Config{}
	c.Redis.Host = "redis"
	c.Redis.Port = "6379"
	c.Redis.Db = 0
	c.Redis.Password = ""
	return redis.NewHelper(c, "sync_test")
}

func setup(timeQuantum time.Duration, appId string) *redisSync {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{})
	return newSync(getClient(), func(config *config) {
		config.interval = timeQuantum
		config.id = appId
	})
}

func teardown() {
	getClient().Client.FlushAll()
}

func TestClientConnection(t *testing.T) {
	rsync := setup(1*time.Minute, uuid.NewV4().String())

	for i := 0; i < 10000; i++ {
		rsync.ping()
	}

	teardown()
}

func TestRedisSync_registerInstance(t *testing.T) {
	rsync := setup(time.Minute, "instance")

	for i := 0; i < 100; i++ {
		rsync.registerInstance()
	}

	position := rsync.instance.position()
	count := rsync.instance.count()
	offset := rsync.instance.offset()

	assert.Equal(t, int64(0), position)
	assert.Equal(t, int64(1), count)
	assert.Equal(t, time.Second*0, offset)

	teardown()
}

func TestRedisSync_registerMultipleInstance(t *testing.T) {
	rsync := &redisSync{}

	for i := 0; i < 60; i++ {
		rsync = setup(time.Minute, "instance"+strconv.Itoa(i))
		rsync.registerInstance()
	}

	position := rsync.instance.position()
	count := rsync.instance.count()
	offset := rsync.instance.offset()

	assert.Equal(t, int64(59), position)
	assert.Equal(t, int64(60), count)
	assert.Equal(t, time.Second*59, offset)

	teardown()
}

func TestRedisSync_Sync(t *testing.T) {
	cancel := make(chan struct{})
	results := make(chan time.Duration)
	go func() {
		rsync := setup(1*time.Minute, "instance1")
		for instance := range rsync.sync(cancel) {
			results <- instance.offset()
		}

	}()

	for i := 0; i < 1; i++ {
		assert.Equal(t, time.Duration(0), <-results)
	}

	teardown()
}

func TestRedisSync_SyncMultiple(t *testing.T) {
	results := make(chan time.Duration)

	expectedTimes := []int{
		int(time.Duration(0)),
		int(time.Duration(time.Second * 30)),
	}

	times := []int{}

	go func() {
		rsync := setup(1*time.Minute, "instance1")
		for instance := range rsync.sync(make(chan struct{})) {
			results <- instance.offset()
		}

	}()

	go func() {
		rsync := setup(1*time.Minute, "instance2")
		for instance := range rsync.sync(make(chan struct{})) {
			results <- instance.offset()
		}
	}()

	for i := 0; i < 2; i++ {
		times = append(times, int(<-results))
	}

	assert.Equal(t, 2, len(times))
	sort.Ints(times)
	assert.EqualValues(t, expectedTimes, times)

	teardown()
}

func TestRedisSync_SyncMultipleWithDeath(t *testing.T) {

	go func() {
		time.Sleep(0 * time.Second)

		rsync := setup(1*time.Minute, "instance1")

		for instance := range rsync.sync(make(chan struct{})) {
			log.WithFields(log.Fields{
				"offset":    instance.off.Seconds(),
				"instances": rsync.instance.count(),
				"position":  rsync.instance.count()}).Println("instance1")
		}
	}()

	go func() {
		time.Sleep(3 * time.Second)

		rsync := setup(1*time.Minute, "instance2")

		for instance := range rsync.sync(make(chan struct{})) {
			log.WithFields(log.Fields{
				"offset":    instance.off.Seconds(),
				"instances": rsync.instance.count(),
				"position":  rsync.instance.count()}).Println("instance2")
		}
	}()

	go func() {
		time.Sleep(5 * time.Second)

		rsync := setup(1*time.Minute, "instance3")
		channel := make(chan struct{})
		go func() {
			time.Sleep(11 * time.Second)
			channel <- struct{}{}
			log.Println("BOOM!!! instance3")
		}()

		for instance := range rsync.sync(channel) {
			log.WithFields(log.Fields{
				"offset":    instance.off.Seconds(),
				"instances": rsync.instance.count(),
				"position":  rsync.instance.count()}).Println("instance3")
		}

	}()

	go func() {
		time.Sleep(9 * time.Second)

		rsync := setup(1*time.Minute, "instance4")

		channel := make(chan struct{})

		for instance := range rsync.sync(channel) {
			log.WithFields(log.Fields{
				"offset":    instance.off.Seconds(),
				"instances": rsync.instance.count(),
				"position":  rsync.instance.count()}).Println("instance4")
		}
	}()

	select {
	case <-time.After(time.Second * 30):
	}

	teardown()
}

func Test_calculateOffset(t *testing.T) {
	type args struct {
		interval  int64
		instances int64
		position  int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{args: args{60, 1, 0}, want: 0},
		{args: args{60, 2, 0}, want: 0},
		{args: args{60, 2, 1}, want: 30},
		{args: args{60, 3, 0}, want: 0},
		{args: args{60, 3, 1}, want: 20},
		{args: args{60, 3, 2}, want: 40},
		{args: args{60, 4, 0}, want: 0},
		{args: args{60, 4, 1}, want: 15},
		{args: args{60, 4, 2}, want: 30},
		{args: args{60, 4, 3}, want: 45},
		{args: args{60, 60, 59}, want: 59},
	}
	for _, tt := range tests {
		if got := calculateOffset(tt.args.interval, tt.args.instances, tt.args.position); got != tt.want {
			t.Errorf("%q. calculateOffset() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
