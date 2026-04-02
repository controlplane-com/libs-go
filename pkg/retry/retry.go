package retry

import (
	"context"
	"math"
	"time"
)

// Config holds retry configuration
type Config struct {
	InitialBackoff time.Duration // Starting backoff duration (default: 100ms)
	MaxBackoff     time.Duration // Maximum backoff duration (default: 30s)
	Multiplier     float64       // Backoff multiplier (default: 2.0)
}

// DefaultConfig returns sensible retry defaults
func DefaultConfig() Config {
	return Config{
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
	}
}

// RetryableFunc is a function that returns (done bool, err error)
// done=true means success, stop retrying
// done=false with nil error means retry
// done=false with error means retry (error is logged)
type RetryableFunc func() (done bool, err error)

// WithExponentialBackoff retries fn indefinitely with exponential backoff until:
// - fn returns done=true
// - ctx is cancelled
func WithExponentialBackoff(ctx context.Context, cfg Config, fn RetryableFunc) error {
	if cfg.InitialBackoff == 0 {
		cfg.InitialBackoff = 100 * time.Millisecond
	}
	if cfg.MaxBackoff == 0 {
		cfg.MaxBackoff = 30 * time.Second
	}
	if cfg.Multiplier == 0 {
		cfg.Multiplier = 2.0
	}

	backoff := cfg.InitialBackoff
	attempt := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		done, err := fn()
		if done {
			return err
		}

		// Calculate next backoff
		attempt++
		backoff = time.Duration(float64(cfg.InitialBackoff) * math.Pow(cfg.Multiplier, float64(attempt)))
		if backoff > cfg.MaxBackoff {
			backoff = cfg.MaxBackoff
		}

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
	}
}
