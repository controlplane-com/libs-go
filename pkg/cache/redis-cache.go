package cache

import (
	"errors"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	redisStore "github.com/eko/gocache/store/redis/v4"
	"github.com/redis/go-redis/v9"
)

func NewTransparentRedisCache[T any](redisClient *redis.Client, transcoder BinaryTranscoder[T], ttl time.Duration) cache.CacheInterface[T] {
	return NewEncodedCache[T](cache.New[string](redisStore.NewRedis(redisClient)), transcoder, ttl)
}

type RedisMode string

var RedisModeHa = RedisMode("ha")
var RedisModeStandalone = RedisMode("standalone")

type RedisClientConfig struct {
	Hosts      []string
	Password   string
	MasterName string
	MaxRetries int
	Mode       RedisMode
}

var NoHostProvidedErr error = errors.New("no host provided")

func RedisClient(config RedisClientConfig) (*redis.Client, error) {
	var client *redis.Client
	switch config.Mode {
	case RedisModeHa:
		options := &redis.FailoverOptions{
			MasterName:    config.MasterName,
			SentinelAddrs: config.Hosts,
			Password:      config.Password,
		}
		if config.Password != "" {
			options.Password = config.Password
		}
		client = redis.NewFailoverClient(options)
		break
	default:
		if len(config.Hosts) == 0 {
			return nil, NoHostProvidedErr
		}
		options := &redis.Options{
			MaxRetries: config.MaxRetries,
			Addr:       config.Hosts[0],
		}
		if config.Password != "" {
			options.Password = config.Password
		}
		client = redis.NewClient(options)
	}
	return client, nil
}
