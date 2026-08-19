package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/controlplane-com/libs-go/pkg/batches"
	"github.com/controlplane-com/libs-go/pkg/logging"
)

// CollectionCache defines a generic interface for caching collection entries keyed by name.
type CollectionCache[T any] interface {
	// Exists checks if an item exists in cache by name. This is intended to be faster than Get because it does not decode the item.
	Exists(ctx context.Context, name string) (bool, error)

	// Get retrieves an item from cache by name
	Get(ctx context.Context, name string) (*T, error)

	// Set stores an item in cache
	Set(ctx context.Context, item *T) error

	// SetMany stores multiple items in cache
	SetMany(ctx context.Context, items []*T) error

	// Delete removes an item from cache
	Delete(ctx context.Context, name string) error

	// DeleteMany removes multiple items from cache
	DeleteMany(ctx context.Context, names []string) error

	// ListAll retrieves all cached items
	ListAll(ctx context.Context) ([]*T, error)

	// Clear removes all items from cache
	Clear(ctx context.Context) error

	// IsValid checks if the cache index is valid
	IsValid(ctx context.Context) (bool, error)

	// MarkValid marks the cache as valid with TTL
	MarkValid(ctx context.Context, ttl time.Duration) error

	// Close releases the underlying client and its background goroutines
	Close() error
}

// RedisCollectionCache implements CollectionCache using Redis.
// It is parameterized over the cached type T and uses a configurable key prefix.
type RedisCollectionCache[T any] struct {
	redis     *redis.Client
	batchSize int
	keyPrefix string
	indexKey  string

	// toKey derives the item key from its logical name (usually identifier).
	toKey func(name string) string

	// marshal and unmarshal allow overriding JSON behavior if needed.
	marshal   func(*T) ([]byte, error)
	unmarshal func([]byte, *T) error

	// nameOf extracts the item's logical name for key construction in Set/SetMany/Clear.
	nameOf func(*T) (string, error)
}

// NewRedisCollectionCache creates a new Redis-backed generic collection cache.
// prefix is used as the key prefix; indexKey is the marker key used by IsValid/MarkValid.
func NewRedisCollectionCache[T any](
	config RedisClientConfig,
	batchSize int,
	prefix string,
	nameOf func(*T) (string, error),
) (CollectionCache[T], error) {
	client, err := RedisClient(config)
	if err != nil {
		return nil, err
	}

	toKey := func(name string) string {
		return fmt.Sprintf("%s:%s", prefix, name)
	}

	return &RedisCollectionCache[T]{
		redis:     client,
		batchSize: batchSize,
		keyPrefix: prefix,
		indexKey:  "--__index__--",
		toKey:     toKey,
		marshal: func(t *T) ([]byte, error) {
			return json.Marshal(t)
		},
		unmarshal: func(b []byte, t *T) error {
			return json.Unmarshal(b, t)
		},
		nameOf: nameOf,
	}, nil
}

func (c *RedisCollectionCache[T]) Exists(ctx context.Context, name string) (bool, error) {
	r, err := c.redis.Exists(ctx, c.toKey(name)).Result()
	if err != nil {
		return false, err
	}
	return r > 0, nil
}

func (c *RedisCollectionCache[T]) Get(ctx context.Context, name string) (*T, error) {
	val, err := c.redis.Get(ctx, c.toKey(name)).Result()
	if err != nil {
		return nil, err
	}
	var item T
	if err := c.unmarshal([]byte(val), &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (c *RedisCollectionCache[T]) Set(ctx context.Context, item *T) error {
	name, err := c.nameOf(item)
	if err != nil {
		return err
	}
	data, err := c.marshal(item)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, c.toKey(name), data, 0).Err()
}

func (c *RedisCollectionCache[T]) SetMany(ctx context.Context, items []*T) error {
	i := batches.NewBatchIterator(items, c.batchSize)
	for i.Next() {
		batch := i.Item()
		var keysAndValues []any
		for _, item := range batch {
			name, err := c.nameOf(item)
			if err != nil {
				return err
			}
			data, err := c.marshal(item)
			if err != nil {
				return err
			}
			keysAndValues = append(keysAndValues, c.toKey(name), data)
		}
		if err := c.redis.MSet(ctx, keysAndValues...).Err(); err != nil {
			return err
		}
	}
	return i.Error()
}

func (c *RedisCollectionCache[T]) Delete(ctx context.Context, name string) error {
	return c.redis.Del(ctx, c.toKey(name)).Err()
}

func (c *RedisCollectionCache[T]) DeleteMany(ctx context.Context, names []string) error {
	pipe := c.redis.TxPipeline()
	for _, name := range names {
		pipe.Del(ctx, c.toKey(name))
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisCollectionCache[T]) ListAll(ctx context.Context) ([]*T, error) {
	cmd := c.redis.Keys(ctx, fmt.Sprintf("%s:*", c.keyPrefix))
	result, err := cmd.Result()
	if err != nil {
		return nil, err
	}

	// Filter out the index key before processing
	indexKeyName := c.toKey(c.indexKey)
	var filteredKeys []string
	for _, key := range result {
		if key != indexKeyName {
			filteredKeys = append(filteredKeys, key)
		}
	}

	var cached []*T
	i := batches.NewBatchIterator(filteredKeys, c.batchSize)
	for i.Next() {
		values, err := c.redis.MGet(ctx, i.Item()...).Result()
		if err != nil {
			return nil, err
		}
		for _, value := range values {
			if value == nil {
				continue
			}
			strValue, ok := value.(string)
			if !ok {
				continue
			}
			var item T
			if err = c.unmarshal([]byte(strValue), &item); err != nil {
				logging.Logger().Sugar().Warnf("error while unmarshalling json from the redis collection cache: %v", err)
				continue
			}
			cached = append(cached, &item)
		}
	}
	if err = i.Error(); err != nil {
		return nil, err
	}
	return cached, nil
}

func (c *RedisCollectionCache[T]) Clear(ctx context.Context) error {
	items, err := c.ListAll(ctx)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	pipe := c.redis.TxPipeline()
	for _, item := range items {
		name, err := c.nameOf(item)
		if err != nil {
			return err
		}
		pipe.Del(ctx, c.toKey(name))
	}
	pipe.Del(ctx, c.toKey(c.indexKey))

	_, err = pipe.Exec(ctx)
	return err
}

func (c *RedisCollectionCache[T]) IsValid(ctx context.Context) (bool, error) {
	err := c.redis.Get(ctx, c.toKey(c.indexKey)).Err()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *RedisCollectionCache[T]) MarkValid(ctx context.Context, ttl time.Duration) error {
	return c.redis.SetEx(ctx, c.toKey(c.indexKey), true, ttl).Err()
}

func (c *RedisCollectionCache[T]) Close() error {
	return c.redis.Close()
}
