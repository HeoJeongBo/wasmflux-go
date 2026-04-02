package util

import (
	"sync"
	"time"
)

// Debounce returns a function that delays invoking fn until after
// wait duration has elapsed since the last call.
// Useful for rate-limiting high-frequency events like resize, scroll, or input.
func Debounce(fn func(), wait time.Duration) func() {
	var mu sync.Mutex
	var timer *time.Timer

	return func() {
		mu.Lock()
		defer mu.Unlock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(wait, fn)
	}
}

// DebounceWithArgs returns a debounced function that passes the latest argument.
func DebounceWithArgs[T any](fn func(T), wait time.Duration) func(T) {
	var mu sync.Mutex
	var timer *time.Timer

	return func(arg T) {
		mu.Lock()
		defer mu.Unlock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(wait, func() { fn(arg) })
	}
}

// Throttle returns a function that invokes fn at most once per interval.
// The first call is executed immediately.
func Throttle(fn func(), interval time.Duration) func() {
	var mu sync.Mutex
	var lastCall time.Time

	return func() {
		mu.Lock()
		defer mu.Unlock()
		now := time.Now()
		if now.Sub(lastCall) >= interval {
			lastCall = now
			fn()
		}
	}
}

// ThrottleWithArgs returns a throttled function that passes the argument.
func ThrottleWithArgs[T any](fn func(T), interval time.Duration) func(T) {
	var mu sync.Mutex
	var lastCall time.Time

	return func(arg T) {
		mu.Lock()
		defer mu.Unlock()
		now := time.Now()
		if now.Sub(lastCall) >= interval {
			lastCall = now
			fn(arg)
		}
	}
}

// ThrottleTrailing is like Throttle but also fires a trailing call
// after the interval if there was a call during the cooldown.
func ThrottleTrailing(fn func(), interval time.Duration) func() {
	var mu sync.Mutex
	var lastCall time.Time
	var timer *time.Timer
	var pending bool

	return func() {
		mu.Lock()
		defer mu.Unlock()
		now := time.Now()
		if now.Sub(lastCall) >= interval {
			lastCall = now
			fn()
			return
		}
		// Schedule trailing call.
		pending = true
		if timer == nil {
			remaining := interval - now.Sub(lastCall)
			timer = time.AfterFunc(remaining, func() {
				mu.Lock()
				if pending {
					pending = false
					lastCall = time.Now()
					mu.Unlock()
					fn()
				} else {
					mu.Unlock()
				}
				mu.Lock()
				timer = nil
				mu.Unlock()
			})
		}
	}
}
