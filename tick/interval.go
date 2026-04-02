//go:build js && wasm

package tick

import (
	"syscall/js"

	"github.com/heojeongbo/wasmflux-go/internal/jsutil"
)

// Interval wraps setInterval with proper cleanup.
type Interval struct {
	jsFn    js.Func
	timerID int
	running bool
}

// NewInterval creates a repeating interval.
// fn is called every intervalMs milliseconds.
func NewInterval(fn func(), intervalMs float64) *Interval {
	t := &Interval{}
	t.jsFn = js.FuncOf(func(_ js.Value, _ []js.Value) any {
		fn()
		return nil
	})
	t.timerID = jsutil.SetInterval(t.jsFn, intervalMs)
	t.running = true
	return t
}

// Stop cancels the interval.
func (t *Interval) Stop() {
	if !t.running {
		return
	}
	t.running = false
	jsutil.ClearInterval(t.timerID)
}

// IsRunning reports whether the interval is active.
func (t *Interval) IsRunning() bool {
	return t.running
}

// Release stops the interval and frees the js.Func.
func (t *Interval) Release() {
	t.Stop()
	t.jsFn.Release()
}

// Timeout wraps setTimeout with proper cleanup.
type Timeout struct {
	jsFn    js.Func
	timerID int
	done    bool
}

// NewTimeout schedules fn to run once after delayMs milliseconds.
func NewTimeout(fn func(), delayMs float64) *Timeout {
	t := &Timeout{}
	t.jsFn = js.FuncOf(func(_ js.Value, _ []js.Value) any {
		t.done = true
		fn()
		return nil
	})
	t.timerID = jsutil.SetTimeout(t.jsFn, delayMs)
	return t
}

// Cancel cancels the timeout if it hasn't fired yet.
func (t *Timeout) Cancel() {
	if t.done {
		return
	}
	jsutil.ClearTimeout(t.timerID)
	t.done = true
}

// Release cancels the timeout and frees the js.Func.
func (t *Timeout) Release() {
	t.Cancel()
	t.jsFn.Release()
}
