//go:build js && wasm

package wasmflux

import (
	"fmt"

	"github.com/heojeongbo/wasmflux-go/bridge"
	"github.com/heojeongbo/wasmflux-go/event"
	"github.com/heojeongbo/wasmflux-go/log"
)

// Module is the extension point for the framework.
// Developers implement this interface to add functionality.
// Lifecycle methods are called in order: Init -> Start -> Stop.
type Module interface {
	// Name returns a unique identifier for the module.
	Name() string

	// Init is called once during app initialization.
	// Use it to set up state and register bridge functions.
	Init(ctx ModuleContext) error

	// Start is called after all modules are initialized.
	// Use it to begin processing (start loops, subscribe to events, etc.).
	Start() error

	// Stop is called during shutdown in reverse registration order.
	// Use it to clean up resources.
	Stop() error
}

// ModuleContext provides access to framework services during Init.
// It also acts as a dependency injection container via Provide/Inject.
type ModuleContext struct {
	Bridge   *bridge.Bridge
	Bus      *event.Bus
	Logger   *log.Logger
	registry *Registry
}

// Provide registers a named service in the DI container.
// Call this during Init() to make services available to other modules.
func (c ModuleContext) Provide(name string, service any) {
	c.registry.Provide(name, service)
}

// Inject retrieves a named service from the DI container.
// Returns nil if not found. Call during Start() after all Init() calls complete.
func (c ModuleContext) Inject(name string) any {
	return c.registry.Inject(name)
}

// MustInject retrieves a named service and panics if not found.
func (c ModuleContext) MustInject(name string) any {
	s := c.registry.Inject(name)
	if s == nil {
		panic(fmt.Sprintf("wasmflux: service %q not found", name))
	}
	return s
}

// InjectAs is a generic helper to inject and type-assert a service.
func InjectAs[T any](ctx ModuleContext, name string) (T, bool) {
	v := ctx.Inject(name)
	if v == nil {
		var zero T
		return zero, false
	}
	t, ok := v.(T)
	return t, ok
}
