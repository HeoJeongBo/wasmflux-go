package errors

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// Recover executes fn and recovers from any panic, returning it as an error.
// Useful for wrapping JS callbacks that may throw.
func Recover(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = newPanicError(r)
		}
	}()
	fn()
	return nil
}

// RecoverFunc wraps a function so that any panic is caught and returned as an error.
func RecoverFunc(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = newPanicError(r)
		}
	}()
	return fn()
}

// PanicError wraps a recovered panic with a Go stack trace.
type PanicError struct {
	Value any
	Stack string
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("panic: %v\n%s", e.Value, e.Stack)
}

func newPanicError(r any) *PanicError {
	return &PanicError{
		Value: r,
		Stack: captureStack(3), // skip: captureStack, newPanicError, deferred func
	}
}

// captureStack returns a formatted stack trace, skipping the first skip frames.
func captureStack(skip int) string {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip, pcs)
	if n == 0 {
		return ""
	}

	var sb strings.Builder
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		sb.WriteString(frame.Function)
		sb.WriteByte('\n')
		sb.WriteByte('\t')
		sb.WriteString(frame.File)
		sb.WriteByte(':')
		sb.WriteString(strconv.Itoa(frame.Line))
		sb.WriteByte('\n')
		if !more {
			break
		}
	}
	return sb.String()
}

// CaptureStack returns the current goroutine's stack trace as a string.
// Useful for attaching to errors for debugging.
func CaptureStack() string {
	return captureStack(2)
}
