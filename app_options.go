//go:build js && wasm

package wasmflux

import (
	"github.com/heojeongbo/wasmflux-go/bridge"
	"github.com/heojeongbo/wasmflux-go/event"
	"github.com/heojeongbo/wasmflux-go/log"
)

// Option configures the App using the functional options pattern.
type Option func(*App)

// WithBridge sets a custom bridge instance.
func WithBridge(b *bridge.Bridge) Option {
	return func(a *App) {
		a.bridge = b
	}
}

// WithBus sets a custom event bus instance.
func WithBus(bus *event.Bus) Option {
	return func(a *App) {
		a.bus = bus
	}
}

// WithLogger sets a custom logger instance.
func WithLogger(l *log.Logger) Option {
	return func(a *App) {
		a.logger = l
	}
}

// WithLogLevel sets the log level (creates a default logger if none set).
func WithLogLevel(level log.Level) Option {
	return func(a *App) {
		a.logLevel = level
	}
}
