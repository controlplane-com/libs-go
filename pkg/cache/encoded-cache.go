package cache

import (
	"context"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	"time"
)

type EncodedCache[T any] struct {
	cache      cache.CacheInterface[string]
	transcoder BinaryTranscoder[T]
	ttl        time.Duration
}

func NewEncodedCache[T any](cache cache.CacheInterface[string], transcoder BinaryTranscoder[T], ttl time.Duration) cache.CacheInterface[T] {
	return &EncodedCache[T]{
		ttl:        ttl,
		cache:      cache,
		transcoder: transcoder,
	}
}

func (r *EncodedCache[T]) Get(ctx context.Context, key any) (T, error) {
	var p T
	cacheItem, err := r.cache.Get(ctx, key)
	if err != nil {
		return p, err
	}
	return r.transcoder.Decode([]byte(cacheItem))
}

func (r *EncodedCache[T]) Set(ctx context.Context, key any, object T, options ...store.Option) error {
	options = append(options, store.WithExpiration(r.ttl))
	b, err := r.transcoder.Encode(object)
	if err != nil {
		return err
	}
	return r.cache.Set(ctx, key, string(b), options...)
}

func (r *EncodedCache[T]) Delete(ctx context.Context, key any) error {
	return r.cache.Delete(ctx, key)
}

func (r *EncodedCache[T]) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	return r.cache.Invalidate(ctx, options...)
}

func (r *EncodedCache[T]) Clear(ctx context.Context) error {
	return r.cache.Clear(ctx)
}

func (r *EncodedCache[T]) GetType() string {
	return r.cache.GetType()
}
