package wasmflux

import "sync"

// Registry is the service registry for dependency injection.
// Modules call Provide() during Init to register services,
// and Inject() during Start to retrieve them.
//
// Performance: Provide/Inject use sync.RWMutex. Since they are only called
// during Init/Start (not on the 60Hz hot path), the overhead is negligible.
// Once injected, modules hold direct references — no further DI lookups at runtime.
type Registry struct {
	mu       sync.RWMutex
	services map[string]any
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		services: make(map[string]any),
	}
}

// Provide stores a named service.
func (r *Registry) Provide(name string, service any) {
	r.mu.Lock()
	r.services[name] = service
	r.mu.Unlock()
}

// Inject retrieves a named service. Returns nil if not found.
func (r *Registry) Inject(name string) any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.services[name]
}

// Has reports whether a named service is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.services[name]
	return ok
}

// Services returns the names of all registered services.
func (r *Registry) Services() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	return names
}
