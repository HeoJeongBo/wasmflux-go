package util

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(1000, 3)

	// First 3 should succeed (burst).
	for i := 0; i < 3; i++ {
		if !rl.Allow() {
			t.Errorf("call %d should be allowed", i)
		}
	}
	// 4th should fail.
	if rl.Allow() {
		t.Error("4th call should be denied")
	}
}

func TestRateLimiter_AllowN(t *testing.T) {
	rl := NewRateLimiter(1000, 5)

	if !rl.AllowN(3) {
		t.Error("AllowN(3) should succeed with burst 5")
	}
	if !rl.AllowN(2) {
		t.Error("AllowN(2) should succeed with 2 remaining")
	}
	if rl.AllowN(1) {
		t.Error("AllowN(1) should fail with 0 remaining")
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	rl := NewRateLimiter(1000, 1) // 1000/sec, burst 1

	rl.Allow() // consume the 1 token
	if rl.Allow() {
		t.Error("should be denied immediately after consuming burst")
	}

	// Wait enough time for at least 1 token to refill.
	time.Sleep(5 * time.Millisecond)
	if !rl.Allow() {
		t.Error("should be allowed after refill")
	}
}

func TestRateLimiter_Tokens(t *testing.T) {
	rl := NewRateLimiter(100, 5)
	if rl.Tokens() != 5 {
		t.Errorf("tokens: %f, want 5", rl.Tokens())
	}

	rl.Allow()
	if rl.Tokens() >= 5 {
		t.Error("tokens should decrease after Allow")
	}
}

func TestRateLimiter_MaxBurst(t *testing.T) {
	rl := NewRateLimiter(0, 3) // zero rate = no refill, burst 3

	for i := 0; i < 3; i++ {
		if !rl.Allow() {
			t.Errorf("call %d should be allowed", i)
		}
	}
	if rl.Allow() {
		t.Error("should be denied after burst exhausted with zero rate")
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	rl := NewRateLimiter(10000, 1) // very high rate
	rl.Allow()                     // consume burst

	done := make(chan struct{})
	go func() {
		rl.Wait() // should unblock quickly at 10000/sec
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Wait should have unblocked")
	}
}
