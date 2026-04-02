package retry

import (
	"context"
	"testing"
	"time"
)

func TestWithExponentialBackoff_ImmediateSuccess(t *testing.T) {
	ctx := context.Background()
	calls := 0

	err := WithExponentialBackoff(ctx, DefaultConfig(), func() (bool, error) {
		calls++
		return true, nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestWithExponentialBackoff_RetryThenSuccess(t *testing.T) {
	ctx := context.Background()
	calls := 0

	cfg := Config{
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		Multiplier:     2.0,
	}

	err := WithExponentialBackoff(ctx, cfg, func() (bool, error) {
		calls++
		if calls < 3 {
			return false, nil // Retry
		}
		return true, nil // Success on 3rd call
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestWithExponentialBackoff_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0

	cfg := Config{
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
		Multiplier:     2.0,
	}

	// Cancel after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := WithExponentialBackoff(ctx, cfg, func() (bool, error) {
		calls++
		return false, nil // Always retry
	})

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	// Should have made at least 1 call before cancellation
	if calls < 1 {
		t.Errorf("expected at least 1 call, got %d", calls)
	}
}

func TestWithExponentialBackoff_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.InitialBackoff != 100*time.Millisecond {
		t.Errorf("expected InitialBackoff 100ms, got %v", cfg.InitialBackoff)
	}
	if cfg.MaxBackoff != 30*time.Second {
		t.Errorf("expected MaxBackoff 30s, got %v", cfg.MaxBackoff)
	}
	if cfg.Multiplier != 2.0 {
		t.Errorf("expected Multiplier 2.0, got %v", cfg.Multiplier)
	}
}
