//go:build js && wasm

package log

import "syscall/js"

// ConsoleWriter writes logs to the browser console and the wasmflux HTML log panel.
type ConsoleWriter struct {
	console  js.Value
	wasmflux js.Value
}

// NewConsoleWriter creates a writer that outputs to browser console
// and the wasmflux.appendLog HTML panel.
func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{
		console:  js.Global().Get("console"),
		wasmflux: js.Global().Get("wasmflux"),
	}
}

func (w *ConsoleWriter) WriteLog(level Level, msg string) {
	method := "log"
	switch level {
	case LevelDebug:
		method = "debug"
	case LevelWarn:
		method = "warn"
	case LevelError:
		method = "error"
	}
	w.console.Call(method, msg)

	if !w.wasmflux.IsUndefined() {
		appendLog := w.wasmflux.Get("appendLog")
		if !appendLog.IsUndefined() {
			appendLog.Invoke(level.LowerString(), msg)
		}
	}
}
