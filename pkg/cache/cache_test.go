package cache

import (
	"context"
	"testing"
	"time"

	"github.com/eko/gocache/lib/v4/store"
)

func TestCacheSetAndGet(t *testing.T) {
	var c = NewInMemoryCache[int](time.Second, time.Second*5)
	ctx := context.Background()
	err := c.Set(ctx, "key1", 1)
	if err != nil {
		t.FailNow()
	}
	i, err := c.Get(ctx, "key1")
	if err != nil || i != 1 {
		t.FailNow()
	}
}

func TestCacheGetInvalidKey(t *testing.T) {
	var c = NewInMemoryCache[int](time.Second, time.Second*5)
	ctx := context.Background()
	_, err := c.Get(ctx, "key1")
	switch err.(type) {
	case *store.NotFound:
		//Good!
		break
	default:
		t.FailNow()
	}
}
