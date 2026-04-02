//go:build js && wasm

package bridge

import (
	"sync"
	"syscall/js"
)

// CallbackPool manages reusable js.Func instances to reduce GC pressure.
// Instead of creating and releasing js.Func for each use,
// callbacks can be checked out and returned.
type CallbackPool struct {
	mu   sync.Mutex
	pool []js.Func
	fns  []func(this js.Value, args []js.Value) any
}

// NewCallbackPool creates a pool pre-allocated with the given capacity.
func NewCallbackPool(capacity int) *CallbackPool {
	cp := &CallbackPool{
		pool: make([]js.Func, 0, capacity),
		fns:  make([]func(this js.Value, args []js.Value) any, 0, capacity),
	}
	return cp
}

// Acquire creates or reuses a js.Func that calls fn.
// The returned index can be used to update or release the callback.
func (cp *CallbackPool) Acquire(fn func(this js.Value, args []js.Value) any) (js.Func, int) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	idx := len(cp.fns)
	cp.fns = append(cp.fns, fn)

	jsFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		cp.mu.Lock()
		f := cp.fns[idx]
		cp.mu.Unlock()
		return f(this, args)
	})
	cp.pool = append(cp.pool, jsFn)

	return jsFn, idx
}

// Update replaces the function at the given index without creating a new js.Func.
func (cp *CallbackPool) Update(idx int, fn func(this js.Value, args []js.Value) any) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	if idx >= 0 && idx < len(cp.fns) {
		cp.fns[idx] = fn
	}
}

// ReleaseAll frees all pooled js.Func instances.
func (cp *CallbackPool) ReleaseAll() {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	for _, fn := range cp.pool {
		fn.Release()
	}
	cp.pool = nil
	cp.fns = nil
}
