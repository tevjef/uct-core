package cache

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/tevjef/uct-core/common/model"
	"github.com/tevjef/uct-core/spike/middleware"
	mtrace "github.com/tevjef/uct-core/spike/middleware/trace"
)

const (
	DEFAULT            = time.Duration(0)
	FOREVER            = time.Duration(-1)
	CacheMiddlewareKey = "spike.cache"
)

var (
	PageCachePrefix = "uct:spike:cache"
	ErrCacheMiss    = errors.New("cache: key not found.")
	ErrNotStored    = errors.New("cache: not stored.")
)

type CacheStore interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}, expire time.Duration) error
	Add(key string, value interface{}, expire time.Duration) error
	Replace(key string, data interface{}, expire time.Duration) error
	Delete(key string) error
	Increment(key string, data uint64) (uint64, error)
	Decrement(key string, data uint64) (uint64, error)
	Flush() error
}

type responseCache struct {
	Status int
	Header http.Header
	Data   []byte
}

type cachedWriter struct {
	gin.ResponseWriter
	status  int
	written bool
	store   CacheStore
	expire  time.Duration
	key     string
	c       *gin.Context
}

func urlEscape(prefix string, u string) string {
	key := url.QueryEscape(u)
	if len(key) > 200 {
		h := sha1.New()
		io.WriteString(h, u)
		key = string(h.Sum(nil))
	}
	var buffer bytes.Buffer
	buffer.WriteString(prefix)
	buffer.WriteString(":")
	buffer.WriteString(key)
	return buffer.String()
}

func newCachedWriter(store CacheStore, expire time.Duration, writer gin.ResponseWriter, key string, ctx *gin.Context) *cachedWriter {
	return &cachedWriter{writer, 0, false, store, expire, key, ctx}
}

func (w *cachedWriter) WriteHeader(code int) {
	w.status = code
	w.written = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *cachedWriter) Status() int {
	return w.status
}

func (w *cachedWriter) Written() bool {
	return w.written
}

func (w *cachedWriter) Write(data []byte) (int, error) {
	ret, err := w.ResponseWriter.Write(data)

	if err == nil {
		if value, exists := w.c.Get(middleware.ResponseKey); exists {
			response, _ := value.(model.Response)
			if b, err := response.Marshal(); err != nil {
				log.WithError(err).Errorln("error while marshaling response while caching")
				return ret, err
			} else {
				val := responseCache{
					w.status,
					w.Header(),
					b,
				}

				err = w.store.Set(w.key, val, w.expire)
				if err != nil {
					log.WithError(err).Errorln("error while setting data in cache")

				}
			}

		}
	}

	return ret, err
}

// Cache Middleware
func Cache(store CacheStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(CacheMiddlewareKey, store)
		c.Next()
	}
}

// Cache Decorator
func CachePage(handle gin.HandlerFunc, expire time.Duration) gin.HandlerFunc {
	return CachePageWithPolicy(handle, PolicyWithExpiration(expire))
}

// Cache Decorator
func CachePageWithPolicy(handle gin.HandlerFunc, policy *Policy) gin.HandlerFunc {
	if policy == nil {
		policy = &Policy{}
	}

	return func(c *gin.Context) {
		span := mtrace.NewSpan(c, "cache.CachePageWithPolicy")
		defer span.Finish()

		var cache responseCache
		store := cacheStoreFromContext(c)

		c.Header(policy.CacheControl())
		if policy.ServerMaxAge == 0 {
			handle(c)
			return
		}

		key := urlEscape(PageCachePrefix, c.Request.URL.RequestURI())
		if err := store.Get(key, &cache); err != nil {
			// replace writer
			writer := newCachedWriter(store, policy.ServerMaxAge, c.Writer, key, c)
			c.Writer = writer
			handle(c)
		} else {
			for k, vals := range cache.Header {
				for _, v := range vals {
					c.Writer.Header().Set(k, v)
				}
			}

			var response model.Response

			if err := response.Unmarshal(cache.Data); err != nil {
				log.WithError(err).Errorln("error while unmarshaling cached response")
				handle(c)
			}

			c.Set(middleware.MetaKey, *response.Meta)
			c.Set(middleware.ResponseKey, response)
		}
	}
}

func cacheStoreFromContext(c *gin.Context) CacheStore {
	return c.Value(CacheMiddlewareKey).(CacheStore)
}
