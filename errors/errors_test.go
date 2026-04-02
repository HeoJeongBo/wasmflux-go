package errors

import (
	"errors"
	"testing"
)

func TestWASMError_Error(t *testing.T) {
	err := &WASMError{Op: "bridge.call", Err: errors.New("timeout")}
	want := "wasmflux bridge.call: timeout"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestWASMError_Unwrap(t *testing.T) {
	inner := errors.New("inner")
	err := &WASMError{Op: "test", Err: inner}
	if !errors.Is(err, inner) {
		t.Error("Unwrap should return inner error")
	}
}

func TestWrap(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if Wrap(nil, "op") != nil {
			t.Error("Wrap(nil) should return nil")
		}
	})

	t.Run("non-nil", func(t *testing.T) {
		inner := errors.New("fail")
		err := Wrap(inner, "bridge")
		var we *WASMError
		if !errors.As(err, &we) {
			t.Fatal("expected WASMError")
		}
		if we.Op != "bridge" {
			t.Errorf("Op: %q, want %q", we.Op, "bridge")
		}
		if !errors.Is(err, inner) {
			t.Error("should unwrap to inner")
		}
	})
}

func TestNew(t *testing.T) {
	err := New("test error")
	if err.Error() != "test error" {
		t.Errorf("got %q", err.Error())
	}
}

func TestIs(t *testing.T) {
	if !Is(ErrNotInitialized, ErrNotInitialized) {
		t.Error("should match sentinel")
	}
	if Is(ErrNotInitialized, ErrTimeout) {
		t.Error("should not match different sentinel")
	}

	wrapped := Wrap(ErrTimeout, "op")
	if !Is(wrapped, ErrTimeout) {
		t.Error("Is should traverse Unwrap chain")
	}
}

func TestAs(t *testing.T) {
	inner := errors.New("inner")
	err := Wrap(inner, "test")

	var we *WASMError
	if !As(err, &we) {
		t.Fatal("As should find WASMError")
	}
	if we.Op != "test" {
		t.Errorf("Op: %q, want %q", we.Op, "test")
	}
}

func TestSentinelErrors(t *testing.T) {
	sentinels := []error{
		ErrNotInitialized,
		ErrAlreadyRunning,
		ErrStopped,
		ErrJSUndefined,
		ErrJSNull,
		ErrTimeout,
	}
	for _, s := range sentinels {
		if s == nil {
			t.Error("sentinel should not be nil")
		}
		if s.Error() == "" {
			t.Error("sentinel should have non-empty message")
		}
	}
}
