//go:build js && wasm

package bridge

import (
	"fmt"
	"syscall/js"
)

// Await blocks until a JS Promise resolves or rejects.
// Returns the resolved value or an error wrapping the rejection.
func Await(promise js.Value) (js.Value, error) {
	ch := make(chan js.Value, 1)
	errCh := make(chan error, 1)

	onResolve := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 {
			ch <- args[0]
		} else {
			ch <- js.Undefined()
		}
		return nil
	})
	defer onResolve.Release()

	onReject := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 {
			errCh <- fmt.Errorf("promise rejected: %s", args[0].Call("toString").String())
		} else {
			errCh <- fmt.Errorf("promise rejected")
		}
		return nil
	})
	defer onReject.Release()

	promise.Call("then", onResolve).Call("catch", onReject)

	select {
	case v := <-ch:
		return v, nil
	case err := <-errCh:
		return js.Undefined(), err
	}
}

// AwaitFunc is a convenience wrapper that awaits a Promise-returning JS function call.
func AwaitFunc(obj js.Value, method string, args ...any) (js.Value, error) {
	promise := obj.Call(method, args...)
	return Await(promise)
}

// NewPromise creates a JS Promise backed by a Go function.
// The function receives resolve and reject callbacks.
func NewPromise(fn func(resolve, reject func(value any))) js.Value {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) any {
		resolveJS := args[0]
		rejectJS := args[1]

		go fn(
			func(value any) { resolveJS.Invoke(value) },
			func(value any) { rejectJS.Invoke(value) },
		)
		return nil
	})

	promise := js.Global().Get("Promise").New(handler)
	handler.Release()
	return promise
}
