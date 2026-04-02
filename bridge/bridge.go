//go:build js && wasm

package bridge

import (
	"syscall/js"

	"github.com/heojeongbo/wasmflux-go/internal/jsutil"
)

// Bridge manages Go↔JS interop.
// It caches global references, pools callbacks, and provides
// typed methods for exposing Go functions and calling JS functions.
type Bridge struct {
	global  js.Value
	ns      js.Value // wasmflux namespace on window
	exposed map[string]js.Func
}

// New creates a new Bridge and initializes the wasmflux JS namespace.
func New() *Bridge {
	ns := jsutil.Global.Get("wasmflux")
	if ns.IsUndefined() {
		ns = js.Global().Get("Object").New()
		jsutil.Global.Set("wasmflux", ns)
	}

	return &Bridge{
		global:  jsutil.Global,
		ns:      ns,
		exposed: make(map[string]js.Func),
	}
}

// Expose registers a Go function callable from JS as wasmflux.<name>().
// The function receives JS arguments and may return a value.
func (b *Bridge) Expose(name string, fn func(this js.Value, args []js.Value) any) {
	if old, ok := b.exposed[name]; ok {
		old.Release()
	}
	jsFn := js.FuncOf(fn)
	b.exposed[name] = jsFn
	b.ns.Set(name, jsFn)
}

// ExposeSimple registers a Go function with no return value.
func (b *Bridge) ExposeSimple(name string, fn func(args []js.Value)) {
	b.Expose(name, func(_ js.Value, args []js.Value) any {
		fn(args)
		return nil
	})
}

// Call invokes a JS function in the wasmflux namespace.
func (b *Bridge) Call(name string, args ...any) js.Value {
	return b.ns.Call("call", append([]any{name}, args...)...)
}

// CallGlobal invokes a global JS function.
func (b *Bridge) CallGlobal(name string, args ...any) js.Value {
	return b.global.Call(name, args...)
}

// Get retrieves a property from the wasmflux namespace.
func (b *Bridge) Get(name string) js.Value {
	return b.ns.Get(name)
}

// Set sets a property on the wasmflux namespace.
func (b *Bridge) Set(name string, value any) {
	b.ns.Set(name, value)
}

// Global returns the cached global JS object.
func (b *Bridge) Global() js.Value {
	return b.global
}

// Namespace returns the wasmflux JS namespace object.
func (b *Bridge) Namespace() js.Value {
	return b.ns
}

// Release frees all registered JS functions.
// Must be called during shutdown to prevent memory leaks.
func (b *Bridge) Release() {
	for name, fn := range b.exposed {
		fn.Release()
		b.ns.Delete(name)
	}
	b.exposed = nil
}
