package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/controlplane-com/libs-go/pkg/types"
	"github.com/eko/gocache/lib/v4/store"
	redisStore "github.com/eko/gocache/store/redis/v4"
	"reflect"
	"time"
)

type CplnRedisStore[T CplnCacheEntry] struct {
	redisStore.RedisStore
	ttl time.Duration
}

func (c *CplnRedisStore[T]) Get(ctx context.Context, key any) (any, error) {
	v, err := c.RedisStore.Get(ctx, key)
	if err != nil {
		return v, err
	}
	return c.unmarshalValue(v)
}

func (c *CplnRedisStore[T]) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	v, d, err := c.RedisStore.GetWithTTL(ctx, key)
	if err != nil {
		return v, d, err
	}
	v, err = c.unmarshalValue(v)
	if err != nil {
		return nil, d, err
	}
	return v, d, err
}

func (c *CplnRedisStore[T]) Set(ctx context.Context, key any, value any, options ...store.Option) error {
	b, err := c.marshalValue(value)
	if err != nil {
		return err
	}
	options = append(options, store.WithExpiration(c.ttl))
	return c.RedisStore.Set(ctx, key, b, options...)
}

func (c *CplnRedisStore[T]) Delete(ctx context.Context, key any) error {
	return c.RedisStore.Delete(ctx, key)
}

func (c *CplnRedisStore[T]) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	return c.RedisStore.Invalidate(ctx, options...)
}

func (c *CplnRedisStore[T]) Clear(ctx context.Context) error {
	return c.RedisStore.Clear(ctx)
}

func (c *CplnRedisStore[T]) GetType() string {
	return c.RedisStore.GetType()
}

func (c *CplnRedisStore[T]) marshalValue(v any) ([]byte, error) {
	switch i := v.(type) {
	case T:
		return i.MarshalBinary()
	default:
		var t T
		return nil, errors.New(fmt.Sprintf("Expected a value of type %s, but received a value of type %s", types.GetTypeName(reflect.TypeOf(v)), types.GetTypeName(reflect.TypeOf(t))))
	}
}

func (c *CplnRedisStore[T]) unmarshalValue(v any) (any, error) {
	switch i := v.(type) {
	case string:
		var target T
		pTarget := &target
		_, err := types.EnsureConcreteValue(reflect.ValueOf(pTarget).Elem())
		if err != nil {
			return nil, err
		}

		err = target.UnmarshalBinary([]byte(i))
		if err != nil {
			return nil, err
		}
		return target, nil
	case []byte:
		var target T
		err := target.UnmarshalBinary(i)
		if err != nil {
			return target, nil
		}
		return nil, err
	}
	return nil, errors.New(fmt.Sprintf("received a value of type %s from the underlying store. Expected string or []byte.", types.GetTypeName(reflect.TypeOf(v))))
}
