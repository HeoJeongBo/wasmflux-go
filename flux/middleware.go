package flux

import (
	"github.com/heojeongbo/wasmflux-go/log"
)

// Middleware intercepts actions before they reach the reducer.
// It receives the next dispatch function and returns a wrapped dispatch function.
type Middleware func(next func(Action)) func(Action)

// LoggerMiddleware logs every dispatched action.
func LoggerMiddleware(logger *log.Logger) Middleware {
	return func(next func(Action)) func(Action) {
		return func(a Action) {
			logger.Debug("dispatch", log.String("action", a.Type))
			next(a)
		}
	}
}

// ThunkMiddleware allows dispatching functions instead of plain actions.
// If the payload is a func(func(Action)), it is called with the dispatch function.
func ThunkMiddleware() Middleware {
	return func(next func(Action)) func(Action) {
		return func(a Action) {
			if thunk, ok := a.Payload.(func(dispatch func(Action))); ok {
				thunk(next)
				return
			}
			next(a)
		}
	}
}
