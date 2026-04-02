package common

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

func NewUuid() string {
	return uuid.New().String()
}

func Try(f func() error) error {
	c := make(chan error, 1)
	catchPanic(c, f)
	return <-c
}

func catchPanic(c chan error, f func() error) {
	defer func() {
		if r := recover(); r != nil {
			c <- errors.New(fmt.Sprintf("Panic in func: %v", r))
		}
		close(c)
	}()

	err := f()
	if err != nil {
		c <- err
	}
}

func BoolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func StrPtr(s string) *string {
	return &s
}

func Float32Ptr(f float32) *float32 {
	return &f
}

func Float64Ptr(f float64) *float64 {
	return &f
}

func IntPtr(i int) *int {
	return &i
}

func Int32Ptr(i int32) *int32 {
	return &i
}

func Int64Ptr(i int64) *int64 {
	return &i
}

func BoolPtr(b bool) *bool {
	return &b
}

// Value functions safely dereference pointers, returning zero value if nil.
// These are useful when working with optional fields that are represented as pointers.

func StrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func Float32Val(f *float32) float32 {
	if f == nil {
		return 0
	}
	return *f
}

func Float64Val(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func IntVal(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func Int32Val(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func Int64Val(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func BoolVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func ExecuteWithRetries(fn func() error, numRetries int, base float64, maximum float64) (int, error) {
	var err error
	i := 1
	for ; i <= numRetries+1; i++ {
		err = fn()
		if err == nil {
			return i - 1, nil
		}

		if i <= numRetries {
			backoffDuration := math.Pow(base, float64(i))
			if backoffDuration > maximum {
				backoffDuration = maximum
			}
			time.Sleep(time.Duration(backoffDuration) * time.Second)
		}
	}

	return i - 1, fmt.Errorf("failed after %d retries: %w", numRetries, err)
}

// ExecuteWithRetriesContext is like ExecuteWithRetries but supports context cancellation.
// The fn receives context so it can also check for cancellation.
func ExecuteWithRetriesContext(ctx context.Context, fn func(ctx context.Context) error, numRetries int, base float64, maximum float64) (int, error) {
	var err error
	i := 1
	for ; i <= numRetries+1; i++ {
		select {
		case <-ctx.Done():
			return i - 1, ctx.Err()
		default:
		}

		err = fn(ctx)
		if err == nil {
			return i - 1, nil
		}

		if i <= numRetries {
			backoffDuration := math.Pow(base, float64(i))
			if backoffDuration > maximum {
				backoffDuration = maximum
			}

			select {
			case <-ctx.Done():
				return i - 1, ctx.Err()
			case <-time.After(time.Duration(backoffDuration) * time.Second):
			}
		}
	}

	return i - 1, fmt.Errorf("failed after %d retries: %w", numRetries, err)
}
