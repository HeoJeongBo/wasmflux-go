package util

import (
	"math"
	"time"
)

// RetryOption configures retry behavior.
type RetryOption func(*retryConfig)

type retryConfig struct {
	maxAttempts int
	baseDelay   time.Duration
	maxDelay    time.Duration
	backoff     func(attempt int, base time.Duration) time.Duration
}

// WithMaxAttempts sets the maximum number of retry attempts.
func WithMaxAttempts(n int) RetryOption {
	return func(c *retryConfig) { c.maxAttempts = n }
}

// WithBaseDelay sets the initial delay between retries.
func WithBaseDelay(d time.Duration) RetryOption {
	return func(c *retryConfig) { c.baseDelay = d }
}

// WithMaxDelay sets the maximum delay cap for exponential backoff.
func WithMaxDelay(d time.Duration) RetryOption {
	return func(c *retryConfig) { c.maxDelay = d }
}

// WithLinearBackoff uses linear backoff (delay * attempt).
func WithLinearBackoff() RetryOption {
	return func(c *retryConfig) {
		c.backoff = func(attempt int, base time.Duration) time.Duration {
			return base * time.Duration(attempt+1)
		}
	}
}

// WithExponentialBackoff uses exponential backoff (delay * 2^attempt).
func WithExponentialBackoff() RetryOption {
	return func(c *retryConfig) {
		c.backoff = func(attempt int, base time.Duration) time.Duration {
			return base * time.Duration(math.Pow(2, float64(attempt)))
		}
	}
}

// Retry executes fn up to maxAttempts times until it succeeds.
// Returns the last error if all attempts fail.
func Retry(fn func() error, opts ...RetryOption) error {
	cfg := &retryConfig{
		maxAttempts: 3,
		baseDelay:   100 * time.Millisecond,
		maxDelay:    10 * time.Second,
		backoff: func(attempt int, base time.Duration) time.Duration {
			return base * time.Duration(math.Pow(2, float64(attempt)))
		},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	var lastErr error
	for i := 0; i < cfg.maxAttempts; i++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if i < cfg.maxAttempts-1 {
			delay := cfg.backoff(i, cfg.baseDelay)
			if delay > cfg.maxDelay {
				delay = cfg.maxDelay
			}
			time.Sleep(delay)
		}
	}
	return lastErr
}
