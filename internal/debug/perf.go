//go:build js && wasm && debug

package debug

import "syscall/js"

var performance = js.Global().Get("performance")

// Mark creates a performance mark for profiling in browser devtools.
func Mark(name string) {
	performance.Call("mark", name)
}

// MeasureStart creates a start mark.
func MeasureStart(name string) {
	performance.Call("mark", name+":start")
}

// MeasureEnd creates an end mark and measures the duration.
// Returns the duration in milliseconds.
func MeasureEnd(name string) float64 {
	performance.Call("mark", name+":end")
	performance.Call("measure", name, name+":start", name+":end")
	entries := performance.Call("getEntriesByName", name, "measure")
	if entries.Length() == 0 {
		return 0
	}
	return entries.Index(entries.Length() - 1).Get("duration").Float()
}

// ClearMarks clears all performance marks.
func ClearMarks() {
	performance.Call("clearMarks")
}

// ClearMeasures clears all performance measures.
func ClearMeasures() {
	performance.Call("clearMeasures")
}
