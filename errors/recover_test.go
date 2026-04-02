package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestRecover_NoPanic(t *testing.T) {
	err := Recover(func() {})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestRecover_PanicString(t *testing.T) {
	err := Recover(func() { panic("boom") })
	if err == nil {
		t.Fatal("expected error")
	}
	var pe *PanicError
	if !errors.As(err, &pe) {
		t.Fatal("expected PanicError")
	}
	if pe.Value != "boom" {
		t.Errorf("Value: %v, want %q", pe.Value, "boom")
	}
	if pe.Stack == "" {
		t.Error("stack trace should not be empty")
	}
}

func TestRecover_PanicError(t *testing.T) {
	inner := errors.New("inner error")
	err := Recover(func() { panic(inner) })
	if err == nil {
		t.Fatal("expected error")
	}
	var pe *PanicError
	if !errors.As(err, &pe) {
		t.Fatal("expected PanicError")
	}
	if pe.Value != inner {
		t.Errorf("Value should be the original error")
	}
}

func TestRecover_PanicInt(t *testing.T) {
	err := Recover(func() { panic(42) })
	if err == nil {
		t.Fatal("expected error")
	}
	var pe *PanicError
	if !errors.As(err, &pe) {
		t.Fatal("expected PanicError")
	}
	if pe.Value != 42 {
		t.Errorf("Value: %v, want 42", pe.Value)
	}
}

func TestRecoverFunc_NoPanic(t *testing.T) {
	err := RecoverFunc(func() error { return nil })
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestRecoverFunc_ReturnsError(t *testing.T) {
	want := errors.New("normal error")
	err := RecoverFunc(func() error { return want })
	if err != want {
		t.Errorf("got %v, want %v", err, want)
	}
}

func TestRecoverFunc_Panic(t *testing.T) {
	err := RecoverFunc(func() error { panic("crash") })
	if err == nil {
		t.Fatal("expected error")
	}
	var pe *PanicError
	if !errors.As(err, &pe) {
		t.Fatal("expected PanicError")
	}
}

func TestPanicError_Error(t *testing.T) {
	pe := &PanicError{Value: "test", Stack: "goroutine 1:\nmain.go:10"}
	msg := pe.Error()
	if !strings.Contains(msg, "panic: test") {
		t.Errorf("should contain 'panic: test', got %q", msg)
	}
	if !strings.Contains(msg, "goroutine 1") {
		t.Errorf("should contain stack trace, got %q", msg)
	}
}

func TestCaptureStack(t *testing.T) {
	stack := CaptureStack()
	if stack == "" {
		t.Error("stack should not be empty")
	}
	if !strings.Contains(stack, "TestCaptureStack") {
		t.Errorf("stack should contain caller name, got %q", stack)
	}
}
