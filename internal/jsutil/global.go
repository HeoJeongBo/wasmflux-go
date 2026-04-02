//go:build js && wasm

package jsutil

import "syscall/js"

// Cached global JS references to avoid repeated lookups.
var (
	Global    = js.Global()
	Document  = Global.Get("document")
	Console   = Global.Get("console")
	Window    = Global.Get("window")
	WasmFlux  = Global.Get("wasmflux")
	Undefined = js.Undefined()
	Null      = js.Null()
)

// Performance returns the JS performance object.
func Performance() js.Value {
	return Global.Get("performance")
}

// Now returns performance.now() as float64 milliseconds.
// More precise than time.Now() and avoids allocation.
func Now() float64 {
	return Performance().Call("now").Float()
}

// RequestAnimationFrame schedules fn to run on the next frame.
// Returns the request ID for cancellation.
func RequestAnimationFrame(fn js.Func) int {
	return Global.Call("requestAnimationFrame", fn).Int()
}

// CancelAnimationFrame cancels a scheduled animation frame.
func CancelAnimationFrame(id int) {
	Global.Call("cancelAnimationFrame", id)
}

// SetTimeout schedules fn after delay milliseconds. Returns timer ID.
func SetTimeout(fn js.Func, delay float64) int {
	return Global.Call("setTimeout", fn, delay).Int()
}

// ClearTimeout cancels a scheduled timeout.
func ClearTimeout(id int) {
	Global.Call("clearTimeout", id)
}

// SetInterval schedules fn to repeat every interval milliseconds. Returns timer ID.
func SetInterval(fn js.Func, interval float64) int {
	return Global.Call("setInterval", fn, interval).Int()
}

// ClearInterval cancels a repeating interval.
func ClearInterval(id int) {
	Global.Call("clearInterval", id)
}
