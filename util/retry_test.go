package util

import (
	"errors"
	"testing"
)

func TestRetry_ImmediateSuccess(t *testing.T) {
	err := Retry(func() error { return nil }, WithBaseDelay(0))
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestRetry_SuccessOnThirdAttempt(t *testing.T) {
	attempts := 0
	err := Retry(func() error {
		attempts++
		if attempts < 3 {
			return errors.New("not yet")
		}
		return nil
	}, WithMaxAttempts(5), WithBaseDelay(0))

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if attempts != 3 {
		t.Errorf("attempts: %d, want 3", attempts)
	}
}

func TestRetry_AllFail(t *testing.T) {
	want := errors.New("persistent failure")
	err := Retry(func() error {
		return want
	}, WithMaxAttempts(3), WithBaseDelay(0))

	if err == nil {
		t.Fatal("expected error")
	}
	if err != want {
		t.Errorf("got %v, want %v", err, want)
	}
}

func TestRetry_MaxAttempts(t *testing.T) {
	attempts := 0
	Retry(func() error {
		attempts++
		return errors.New("fail")
	}, WithMaxAttempts(5), WithBaseDelay(0))

	if attempts != 5 {
		t.Errorf("attempts: %d, want 5", attempts)
	}
}

func TestRetry_LinearBackoff(t *testing.T) {
	attempts := 0
	Retry(func() error {
		attempts++
		if attempts < 2 {
			return errors.New("fail")
		}
		return nil
	}, WithMaxAttempts(3), WithBaseDelay(0), WithLinearBackoff())

	if attempts != 2 {
		t.Errorf("attempts: %d, want 2", attempts)
	}
}

func TestRetry_WithMaxDelay(t *testing.T) {
	attempts := 0
	Retry(func() error {
		attempts++
		if attempts < 3 {
			return errors.New("fail")
		}
		return nil
	}, WithMaxAttempts(5), WithBaseDelay(0), WithMaxDelay(0))

	if attempts != 3 {
		t.Errorf("attempts: %d, want 3", attempts)
	}
}

func TestRetry_ExponentialBackoff(t *testing.T) {
	attempts := 0
	Retry(func() error {
		attempts++
		if attempts < 2 {
			return errors.New("fail")
		}
		return nil
	}, WithMaxAttempts(3), WithBaseDelay(0), WithExponentialBackoff())

	if attempts != 2 {
		t.Errorf("attempts: %d, want 2", attempts)
	}
}
