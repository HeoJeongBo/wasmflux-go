package util

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastTime time.Time
}

// NewRateLimiter creates a rate limiter that allows rate operations per second
// with a maximum burst of burst operations.
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	return &RateLimiter{
		tokens:   float64(burst),
		max:      float64(burst),
		rate:     rate,
		lastTime: time.Now(),
	}
}

// Allow reports whether an operation is allowed. Consumes one token if so.
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refill()
	if r.tokens >= 1 {
		r.tokens--
		return true
	}
	return false
}

// AllowN reports whether n operations are allowed. Consumes n tokens if so.
func (r *RateLimiter) AllowN(n int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refill()
	needed := float64(n)
	if r.tokens >= needed {
		r.tokens -= needed
		return true
	}
	return false
}

// Wait blocks until a token is available.
func (r *RateLimiter) Wait() {
	for {
		if r.Allow() {
			return
		}
		time.Sleep(time.Millisecond)
	}
}

// Tokens returns the current number of available tokens.
func (r *RateLimiter) Tokens() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refill()
	return r.tokens
}

func (r *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(r.lastTime).Seconds()
	r.tokens += elapsed * r.rate
	if r.tokens > r.max {
		r.tokens = r.max
	}
	r.lastTime = now
}
