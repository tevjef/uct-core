package cache

import (
	"testing"
	"time"
)

// These tests require redis server running on localhost:6379 (the default)
const redisTestServer = "redis:6379"

var newRedisStore = func(t *testing.T, defaultExpiration time.Duration) CacheStore {
	redisCache := NewRedisCache(redisTestServer, "", 9, defaultExpiration)
	err := redisCache.Flush()
	if err != nil {
		t.Fatal(err.Error())
	}
	return redisCache
}

func TestRedisCache_TypicalGetSet(t *testing.T) {
	typicalGetSet(t, newRedisStore)
}

func TestRedisCache_IncrDecr(t *testing.T) {
	incrDecr(t, newRedisStore)
}

func TestRedisCache_Expiration(t *testing.T) {
	expiration(t, newRedisStore)
}

func TestRedisCache_EmptyCache(t *testing.T) {
	emptyCache(t, newRedisStore)
}

func TestRedisCache_Replace(t *testing.T) {
	testReplace(t, newRedisStore)
}

func TestRedisCache_Add(t *testing.T) {
	testAdd(t, newRedisStore)
}
