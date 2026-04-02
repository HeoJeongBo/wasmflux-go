//go:build js && wasm

package tick

import (
	"syscall/js"

	"github.com/heojeongbo/wasmflux-go/internal/jsutil"
)

// RAFLoop wraps requestAnimationFrame into a managed loop.
// It allocates exactly one js.Func for the entire lifetime of the loop,
// reusing it every frame to avoid GC pressure.
type RAFLoop struct {
	fn      func(dt float64) // user callback, receives delta time in ms
	jsFn    js.Func          // allocated once, reused every frame
	rafID   int
	running bool
	lastT   float64 // last frame timestamp
}

// NewRAFLoop creates a RAF loop. The callback receives the delta time
// in milliseconds since the last frame.
func NewRAFLoop(fn func(dt float64)) *RAFLoop {
	r := &RAFLoop{fn: fn}
	r.jsFn = js.FuncOf(func(_ js.Value, args []js.Value) any {
		if !r.running {
			return nil
		}
		now := args[0].Float() // DOMHighResTimeStamp from RAF
		dt := float64(0)
		if r.lastT > 0 {
			dt = now - r.lastT
		}
		r.lastT = now
		r.fn(dt)
		r.rafID = jsutil.RequestAnimationFrame(r.jsFn)
		return nil
	})
	return r
}

// Start begins the animation frame loop.
func (r *RAFLoop) Start() {
	if r.running {
		return
	}
	r.running = true
	r.lastT = 0
	r.rafID = jsutil.RequestAnimationFrame(r.jsFn)
}

// Stop halts the animation frame loop.
func (r *RAFLoop) Stop() {
	r.running = false
	if r.rafID > 0 {
		jsutil.CancelAnimationFrame(r.rafID)
		r.rafID = 0
	}
}

// IsRunning reports whether the loop is active.
func (r *RAFLoop) IsRunning() bool {
	return r.running
}

// Release frees the underlying js.Func. Call after Stop when done.
func (r *RAFLoop) Release() {
	r.Stop()
	r.jsFn.Release()
}
