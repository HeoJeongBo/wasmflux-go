//go:build js && wasm && debug

package debug

import (
	"runtime"
	"syscall/js"
)

// Inspector exposes runtime state to the browser JS console for debugging.
type Inspector struct {
	ns     js.Value
	fields map[string]js.Func
}

// NewInspector creates an inspector and attaches it to wasmflux._debug in JS.
func NewInspector() *Inspector {
	global := js.Global()
	ns := global.Get("wasmflux")
	if ns.IsUndefined() {
		ns = js.Global().Get("Object").New()
		global.Set("wasmflux", ns)
	}

	debugNs := js.Global().Get("Object").New()
	ns.Set("_debug", debugNs)

	ins := &Inspector{
		ns:     debugNs,
		fields: make(map[string]js.Func),
	}

	ins.expose("memstats", func(_ js.Value, _ []js.Value) any {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		obj := js.Global().Get("Object").New()
		obj.Set("alloc", m.Alloc)
		obj.Set("totalAlloc", m.TotalAlloc)
		obj.Set("sys", m.Sys)
		obj.Set("numGC", m.NumGC)
		obj.Set("goroutines", runtime.NumGoroutine())
		return obj
	})

	ins.expose("gc", func(_ js.Value, _ []js.Value) any {
		runtime.GC()
		return nil
	})

	return ins
}

func (ins *Inspector) expose(name string, fn func(js.Value, []js.Value) any) {
	jsFn := js.FuncOf(fn)
	ins.fields[name] = jsFn
	ins.ns.Set(name, jsFn)
}

// Release frees all exposed JS functions.
func (ins *Inspector) Release() {
	for name, fn := range ins.fields {
		fn.Release()
		ins.ns.Delete(name)
	}
	ins.fields = nil
}
