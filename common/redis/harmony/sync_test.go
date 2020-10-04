package harmony

import (
	"context"
	"sort"
	"strconv"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	_ "github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/tevjef/uct-backend/common/conf"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/common/redis"
)

func getClient(t *testing.T) *redis.Helper {
	c := conf.Config{}
	c.Redis.Host = "redis"
	c.Redis.Port = "6379"
	c.Redis.Db = 11
	c.Redis.Password = ""

	if client, err := model.OpenRedis(c.RedisAddr(), c.Redis.Password, c.Redis.Db, 1); err != nil {
		t.Fatal(err.Error())
		return nil
	} else if err := client.Ping().Err(); err != nil {
		t.Fatal(err.Error())
		return nil
	} else {
		return &redis.Helper{
			NameSpace: "sync_test",
			Client:    client,
		}
	}
}

func setup(t *testing.T, timeQuantum time.Duration, appId string) *redisSync {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{})
	return newSync(getClient(t), func(config *syncConfig) {
		config.interval = timeQuantum
		config.id = appId
	})
}

func teardown(t *testing.T) {
	if c := getClient(t); c != nil {
		if err := c.Client.FlushDb().Err(); err != nil {
			t.Fatal(err.Error())
		}
	}
}

func TestClientConnection(t *testing.T) {
	rsync := setup(t, 1*time.Minute, uuid.NewV4().String())

	for i := 0; i < 10000; i++ {
		if err := rsync.ping(); err != nil {
			t.Errorf("ping failed: %s", err)
		}
	}

	teardown(t)
}

func TestRedisSync_registerInstance(t *testing.T) {
	rsync := setup(t, time.Minute, "instance")

	for i := 0; i < 100; i++ {
		if err := rsync.registerInstance(); err != nil {
			t.Errorf("registerInstance failed: %s", err)
		}
	}

	position := rsync.instance.position()
	count := rsync.instance.count()
	offset := rsync.instance.offset()

	assert.Equal(t, int64(0), position)
	assert.Equal(t, int64(1), count)
	assert.Equal(t, time.Second*0, offset)

	teardown(t)
}

func TestRedisSync_registerMultipleInstance(t *testing.T) {
	rsync := &redisSync{}

	for i := 0; i < 60; i++ {
		rsync = setup(t, time.Minute, "instance"+strconv.Itoa(i))
		if err := rsync.registerInstance(); err != nil {
			t.Errorf("registerInstance failed: %s", err)
		}
	}

	position := rsync.instance.position()
	count := rsync.instance.count()
	offset := rsync.instance.offset()

	assert.Equal(t, int64(59), position)
	assert.Equal(t, int64(60), count)
	assert.Equal(t, time.Second*59, offset)

	teardown(t)
}

func TestRedisSync_Sync(t *testing.T) {
	results := make(chan time.Duration)
	go func() {
		rsync := setup(t, 1*time.Minute, "instance1")
		for instance := range rsync.sync(context.Background()) {
			results <- instance.offset()
		}

	}()

	for i := 0; i < 1; i++ {
		assert.Equal(t, time.Duration(0), <-results)
	}

	teardown(t)
}

func TestRedisSync_SyncMultiple(t *testing.T) {
	results := make(chan time.Duration)

	expectedTimes := []int{
		int(time.Duration(0)),
		int(time.Duration(time.Second * 30)),
	}

	times := []int{}

	go func() {
		rsync := setup(t, 1*time.Minute, "instance1")
		for instance := range rsync.sync(context.Background()) {
			results <- instance.offset()
		}

	}()

	go func() {
		rsync := setup(t, 1*time.Minute, "instance2")
		for instance := range rsync.sync(context.Background()) {
			results <- instance.offset()
		}
	}()

	for i := 0; i < 2; i++ {
		times = append(times, int(<-results))
	}

	assert.Equal(t, 2, len(times))
	sort.Ints(times)
	assert.EqualValues(t, expectedTimes, times)

	teardown(t)
}

func TestRedisSync_SyncMultipleWithDeath(t *testing.T) {

	go func() {
		time.Sleep(0 * time.Second)

		rsync := setup(t, 1*time.Minute, "instance1")

		for instance := range rsync.sync(context.Background()) {
			log.WithFields(log.Fields{
				"offset":    instance.off.Seconds(),
				"instances": rsync.instance.count(),
				"position":  rsync.instance.position(),
				"id":        instance.id}).Println()
		}
	}()

	go func() {
		time.Sleep(3 * time.Second)

		rsync := setup(t, 1*time.Minute, "instance2")

		for instance := range rsync.sync(context.Background()) {
			log.WithFields(log.Fields{
				"offset":    instance.off.Seconds(),
				"instances": rsync.instance.count(),
				"position":  rsync.instance.position(),
				"id":        instance.id}).Println()
		}
	}()

	go func() {
		time.Sleep(5 * time.Second)

		rsync := setup(t, 1*time.Minute, "instance3")
		ctx, cf := context.WithCancel(context.Background())

		go func() {
			time.Sleep(11 * time.Second)
			log.Println("BOOM instance3")
			cf()
		}()

		for instance := range rsync.sync(ctx) {
			log.WithFields(log.Fields{
				"offset":    instance.off.Seconds(),
				"instances": rsync.instance.count(),
				"position":  rsync.instance.position(),
				"id":        instance.id}).Println()
		}

	}()

	go func() {
		time.Sleep(9 * time.Second)
		rsync := setup(t, 1*time.Minute, "instance4")
		for instance := range rsync.sync(context.Background()) {
			log.WithFields(log.Fields{
				"offset":    instance.off.Seconds(),
				"instances": rsync.instance.count(),
				"position":  rsync.instance.position(),
				"id":        instance.id}).Println()
		}
	}()

	select {
	case <-time.After(time.Second * 30):
	}

	//teardown(t)
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
