package cache

import (
	"gopkg.in/redis.v5"
	"time"
)

// Wraps the Redis client to meet the Cache interface.
type RedisStore struct {
	client            *redis.Client
	defaultExpiration time.Duration
}

// until redigo supports sharding/clustering, only one host will be in hostList
func NewRedisCache(host string, password string, database int, defaultExpiration time.Duration) *RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       database,
	})
	return &RedisStore{client, defaultExpiration}
}

func (c *RedisStore) Set(key string, value interface{}, expires time.Duration) error {
	b, err := serialize(value)
	if err != nil {
		return err
	}

	return c.client.Set(key, b, c.expiration(expires)).Err()
}

func (c *RedisStore) Add(key string, value interface{}, expires time.Duration) error {
	if exists, _ := c.Exists(key); exists {
		return ErrNotStored
	}
	return c.client.Set(key, value, c.expiration(expires)).Err()
}

func (c *RedisStore) Get(key string, ptrValue interface{}) error {
	item, err := c.client.Get(key).Bytes()

	if item == nil {
		return ErrCacheMiss
	}

	if err != nil {
		return err
	}

	return deserialize(item, ptrValue)
}

func (c *RedisStore) Replace(key string, value interface{}, expires time.Duration) error {
	if exists, _ := c.Exists(key); !exists {
		return ErrNotStored
	}

	err := c.Set(key, value, c.expiration(expires))
	if value == nil {
		return ErrNotStored
	} else {
		return err
	}
}

func (c *RedisStore) Exists(key string) (bool, error) {
	return c.client.Exists(key).Result()
}

func (c *RedisStore) Delete(key string) error {
	if exists, _ := c.Exists(key); !exists {
		return ErrCacheMiss
	}

	return c.client.Del(key).Err()
}

func (c *RedisStore) Increment(key string, delta uint64) (uint64, error) {
	// Check for existence *before* increment as per the cache contract.
	// redis will auto create the key, and we don't want that. Since we need to do increment
	// ourselves instead of natively via INCRBY (redis doesn't support wrapping), we get the value
	// and do the exists check this way to minimize calls to Redis
	val, err := c.client.Get(key).Uint64()
	if err != nil {
		return 0, ErrCacheMiss
	} else {
		var sum uint64 = val + delta
		err = c.client.Set(key, sum, 0).Err()
		if err != nil {
			return 0, err
		}
		return sum, nil
	}
}

func (c *RedisStore) Decrement(key string, delta uint64) (uint64, error) {
	// Check for existence *before* increment as per the cache contract.
	// redis will auto create the key, and we don't want that, hence the exists call
	val, err := c.client.Get(key).Uint64()
	if err != nil {
		return 0, ErrCacheMiss
	}
	// Decrement contract says you can only go to 0
	// so we go fetch the value and if the delta is greater than the amount,
	// 0 out the value
	if delta > val {
		decr := c.client.DecrBy(key, int64(val)).Val()
		return uint64(decr), nil
	}
	decr := c.client.DecrBy(key, int64(delta)).Val()
	return uint64(decr), nil
}

func (c *RedisStore) Flush() error {
	return c.client.FlushDb().Err()
}

func (c *RedisStore) expiration(expires time.Duration) time.Duration {
	switch expires {
	case DEFAULT:
		expires = c.defaultExpiration
	case FOREVER:
		expires = time.Duration(0)
	}
	return expires
}
