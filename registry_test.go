package wasmflux

import (
	"fmt"
	"testing"
)

func TestRegistry_ProvideInject(t *testing.T) {
	r := NewRegistry()

	r.Provide("counter.store", "fake-store")
	got := r.Inject("counter.store")
	if got != "fake-store" {
		t.Errorf("got %v, want %q", got, "fake-store")
	}
}

func TestRegistry_InjectMissing(t *testing.T) {
	r := NewRegistry()

	got := r.Inject("nonexistent")
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestRegistry_Has(t *testing.T) {
	r := NewRegistry()

	if r.Has("x") {
		t.Error("expected Has to return false")
	}
	r.Provide("x", 42)
	if !r.Has("x") {
		t.Error("expected Has to return true")
	}
}

func TestRegistry_Services(t *testing.T) {
	r := NewRegistry()

	r.Provide("a", 1)
	r.Provide("b", 2)

	names := r.Services()
	if len(names) != 2 {
		t.Errorf("got %d services, want 2", len(names))
	}
}

func TestRegistry_Overwrite(t *testing.T) {
	r := NewRegistry()

	r.Provide("svc", "v1")
	r.Provide("svc", "v2")

	got := r.Inject("svc")
	if got != "v2" {
		t.Errorf("got %v, want %q", got, "v2")
	}
}

// ---------------------------------------------------------------------------
// Benchmarks: DI overhead measurement
// ---------------------------------------------------------------------------

// BenchmarkRegistry_Provide measures the cost of registering a service.
// Expected: very fast, only called during Init (not on hot path).
func BenchmarkRegistry_Provide(b *testing.B) {
	r := NewRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Provide("counter.store", i)
	}
}

// BenchmarkRegistry_Inject measures the cost of looking up a service.
// Expected: very fast (RLock + map lookup), only called during Start.
func BenchmarkRegistry_Inject(b *testing.B) {
	r := NewRegistry()
	r.Provide("counter.store", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Inject("counter.store")
	}
}

// BenchmarkRegistry_Inject_TypeAssert measures Inject + type assertion.
// This is what InjectAs does internally.
func BenchmarkRegistry_Inject_TypeAssert(b *testing.B) {
	r := NewRegistry()
	r.Provide("counter.store", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := r.Inject("counter.store")
		_ = v.(int)
	}
}

// BenchmarkDirectFieldAccess is the baseline: direct struct field access.
// Compare with Inject to see the overhead of DI.
func BenchmarkDirectFieldAccess(b *testing.B) {
	type holder struct{ store int }
	h := holder{store: 42}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.store
	}
}

// BenchmarkRegistry_Inject_ManyServices tests lookup with many registered services.
func BenchmarkRegistry_Inject_ManyServices(b *testing.B) {
	r := NewRegistry()
	for i := 0; i < 100; i++ {
		r.Provide(fmt.Sprintf("svc.%d", i), i)
	}
	r.Provide("target", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Inject("target")
	}
}
