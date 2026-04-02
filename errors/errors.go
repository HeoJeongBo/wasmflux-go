package errors

import (
	"errors"
	"fmt"
)

// Standard sentinel errors.
var (
	ErrNotInitialized = errors.New("wasmflux: not initialized")
	ErrAlreadyRunning = errors.New("wasmflux: already running")
	ErrStopped        = errors.New("wasmflux: stopped")
	ErrJSUndefined    = errors.New("wasmflux: js value is undefined")
	ErrJSNull         = errors.New("wasmflux: js value is null")
	ErrTimeout        = errors.New("wasmflux: timeout")
)

// WASMError wraps errors with WASM operation context.
type WASMError struct {
	Op  string // operation that failed
	Err error  // underlying error
}

func (e *WASMError) Error() string {
	return fmt.Sprintf("wasmflux %s: %v", e.Op, e.Err)
}

func (e *WASMError) Unwrap() error {
	return e.Err
}

// Wrap wraps an error with operation context.
func Wrap(err error, op string) error {
	if err == nil {
		return nil
	}
	return &WASMError{Op: op, Err: err}
}

// New creates a new error with the given message.
func New(msg string) error {
	return errors.New(msg)
}

// Is reports whether any error in err's chain matches target.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target.
func As(err error, target any) bool {
	return errors.As(err, target)
}
