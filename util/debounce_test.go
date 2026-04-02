package util

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestDebounce(t *testing.T) {
	var count atomic.Int32
	fn := Debounce(func() { count.Add(1) }, 50*time.Millisecond)

	// Rapid calls — only last should fire.
	for i := 0; i < 10; i++ {
		fn()
	}

	time.Sleep(100 * time.Millisecond)
	if c := count.Load(); c != 1 {
		t.Errorf("count: %d, want 1", c)
	}
}

func TestDebounce_ResetTimer(t *testing.T) {
	var count atomic.Int32
	fn := Debounce(func() { count.Add(1) }, 50*time.Millisecond)

	fn()
	time.Sleep(30 * time.Millisecond)
	fn() // reset timer
	time.Sleep(30 * time.Millisecond)

	if c := count.Load(); c != 0 {
		t.Errorf("should not have fired yet, count: %d", c)
	}

	time.Sleep(40 * time.Millisecond)
	if c := count.Load(); c != 1 {
		t.Errorf("count: %d, want 1", c)
	}
}

func TestDebounceWithArgs(t *testing.T) {
	var last atomic.Int32
	fn := DebounceWithArgs(func(n int) { last.Store(int32(n)) }, 50*time.Millisecond)

	fn(1)
	fn(2)
	fn(3) // only this should be used

	time.Sleep(100 * time.Millisecond)
	if v := last.Load(); v != 3 {
		t.Errorf("last: %d, want 3", v)
	}
}

func TestThrottle(t *testing.T) {
	var count atomic.Int32
	fn := Throttle(func() { count.Add(1) }, 50*time.Millisecond)

	fn() // executes immediately
	fn() // throttled
	fn() // throttled

	if c := count.Load(); c != 1 {
		t.Errorf("count: %d, want 1 (first call only)", c)
	}

	time.Sleep(60 * time.Millisecond)
	fn() // should execute after interval
	if c := count.Load(); c != 2 {
		t.Errorf("count: %d, want 2", c)
	}
}

func TestThrottleWithArgs(t *testing.T) {
	var last atomic.Int32
	fn := ThrottleWithArgs(func(n int) { last.Store(int32(n)) }, 50*time.Millisecond)

	fn(10) // executes
	fn(20) // throttled

	if v := last.Load(); v != 10 {
		t.Errorf("last: %d, want 10", v)
	}
}

func TestThrottleTrailing(t *testing.T) {
	var count atomic.Int32
	fn := ThrottleTrailing(func() { count.Add(1) }, 50*time.Millisecond)

	fn() // leading call
	fn() // pending trailing

	if c := count.Load(); c != 1 {
		t.Errorf("count: %d, want 1 (leading)", c)
	}

	time.Sleep(100 * time.Millisecond)
	if c := count.Load(); c != 2 {
		t.Errorf("count: %d, want 2 (leading + trailing)", c)
	}
}
