package cache

import (
	"encoding"
	"encoding/json"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/store/go_cache/v4"
	redis_store "github.com/eko/gocache/store/redis/v4"
	"github.com/redis/go-redis/v9"
)
import gocache "github.com/patrickmn/go-cache"

type CacheInterface[T any] interface {
	cache.CacheInterface[T]
}

type Cache[T any] struct {
	*cache.Cache[T]
}

type CplnCacheEntry interface {
	encoding.BinaryUnmarshaler
	encoding.BinaryMarshaler
}

type BinaryTranscoder[T any] interface {
	Encode(T) ([]byte, error)
	Decode([]byte) (T, error)
}

type JsonTranscoder[T any] struct {
}

func (j JsonTranscoder[T]) Encode(t T) ([]byte, error) {
	return json.Marshal(t)
}

func (j JsonTranscoder[T]) Decode(bytes []byte) (T, error) {
	var t T
	if bytes == nil {
		bytes = []byte("null")
	}
	return t, json.Unmarshal(bytes, &t)
}

type JsonCplnCacheEntry[T any] struct {
	payload T
}

func NewJsonCplnCacheEntry[T any](payload T) *JsonCplnCacheEntry[T] {
	return &JsonCplnCacheEntry[T]{payload: payload}
}

func (a *JsonCplnCacheEntry[T]) Payload() T {
	return a.payload
}

func (a *JsonCplnCacheEntry[T]) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &a.payload)
}

func (a *JsonCplnCacheEntry[T]) MarshalBinary() (data []byte, err error) {
	return json.Marshal(a.payload)
}

func NewInMemoryCache[T any](ttl time.Duration, garbageCollectionRate time.Duration) *Cache[T] {
	c := gocache.New(ttl, garbageCollectionRate)
	store := go_cache.NewGoCache(c)
	return &Cache[T]{Cache: cache.New[T](store)}
}

func NewRedisCache[T CplnCacheEntry](ttl time.Duration, redisClient *redis.Client) *Cache[T] {
	if redisClient == nil {
		redisClient = DefaultRedisClient()
	}
	store := &CplnRedisStore[T]{RedisStore: *redis_store.NewRedis(redisClient), ttl: ttl}
	return &Cache[T]{Cache: cache.New[T](store)}
}

func DefaultRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{Addr: "cpln-redis:6379"})
}
