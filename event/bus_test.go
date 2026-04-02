package event

import (
	"sync"
	"testing"
)

func TestBus_EmitOn(t *testing.T) {
	bus := NewBus()

	var received Event
	bus.On("test", func(e Event) {
		received = e
	})

	bus.Emit("test", "payload")

	if received.Topic != "test" {
		t.Errorf("topic: %q, want %q", received.Topic, "test")
	}
	if received.Data != "payload" {
		t.Errorf("data: %v, want %q", received.Data, "payload")
	}
}

func TestBus_Unsubscribe(t *testing.T) {
	bus := NewBus()

	count := 0
	unsub := bus.On("test", func(_ Event) {
		count++
	})

	bus.Emit("test", nil)
	unsub()
	bus.Emit("test", nil)

	if count != 1 {
		t.Errorf("count: %d, want 1", count)
	}
}

func TestBus_MultipleHandlers(t *testing.T) {
	bus := NewBus()

	count := 0
	bus.On("test", func(_ Event) { count++ })
	bus.On("test", func(_ Event) { count++ })
	bus.On("other", func(_ Event) { count++ })

	bus.Emit("test", nil)

	if count != 2 {
		t.Errorf("count: %d, want 2", count)
	}
}

func TestBus_EmitWithTimestamp(t *testing.T) {
	bus := NewBus()

	var ts float64
	bus.On("test", func(e Event) {
		ts = e.Timestamp
	})

	bus.EmitWithTimestamp("test", nil, 123.456)

	if ts != 123.456 {
		t.Errorf("timestamp: %f, want 123.456", ts)
	}
}

func TestBus_HasHandlers(t *testing.T) {
	bus := NewBus()

	if bus.HasHandlers("test") {
		t.Error("expected no handlers")
	}

	unsub := bus.On("test", func(_ Event) {})
	if !bus.HasHandlers("test") {
		t.Error("expected handlers")
	}

	unsub()
	if bus.HasHandlers("test") {
		t.Error("expected no handlers after unsub")
	}
}

func TestBus_Clear(t *testing.T) {
	bus := NewBus()
	bus.On("a", func(_ Event) {})
	bus.On("b", func(_ Event) {})
	bus.Clear()

	if bus.HasHandlers("a") || bus.HasHandlers("b") {
		t.Error("expected no handlers after clear")
	}
}

func TestBus_ConcurrentEmit(t *testing.T) {
	bus := NewBus()

	var mu sync.Mutex
	count := 0
	bus.On("test", func(_ Event) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Emit("test", nil)
		}()
	}
	wg.Wait()

	if count != 100 {
		t.Errorf("count: %d, want 100", count)
	}
}

func TestBus_Topics(t *testing.T) {
	bus := NewBus()
	bus.On("alpha", func(_ Event) {})
	bus.On("beta", func(_ Event) {})

	topics := bus.Topics()
	if len(topics) != 2 {
		t.Errorf("topics: %d, want 2", len(topics))
	}
}

func TestBus_HandlerCount(t *testing.T) {
	bus := NewBus()
	bus.On("x", func(_ Event) {})
	bus.On("x", func(_ Event) {})
	bus.On("y", func(_ Event) {})

	if bus.HandlerCount("x") != 2 {
		t.Errorf("x count: %d, want 2", bus.HandlerCount("x"))
	}
	if bus.HandlerCount("y") != 1 {
		t.Errorf("y count: %d, want 1", bus.HandlerCount("y"))
	}
	if bus.HandlerCount("z") != 0 {
		t.Errorf("z count: %d, want 0", bus.HandlerCount("z"))
	}
}

func TestBus_EmitAsync(t *testing.T) {
	bus := NewBus()
	ch := make(chan int, 1)
	bus.On("async", func(e Event) {
		ch <- e.Data.(int)
	})

	bus.EmitAsync("async", 42)

	select {
	case v := <-ch:
		if v != 42 {
			t.Errorf("got %d, want 42", v)
		}
	case <-make(chan struct{}):
		t.Fatal("timeout waiting for async emit")
	}
}

func TestBus_Once_FiresOnce(t *testing.T) {
	bus := NewBus()
	count := 0
	bus.Once("once", func(_ Event) { count++ })

	bus.Emit("once", nil)
	bus.Emit("once", nil)
	bus.Emit("once", nil)

	if count != 1 {
		t.Errorf("count: %d, want 1", count)
	}
}

func TestBus_EmitNoHandlers(t *testing.T) {
	bus := NewBus()
	// Should not panic.
	bus.Emit("nonexistent", "data")
}

func BenchmarkBus_Emit(b *testing.B) {
	bus := NewBus()
	bus.On("test", func(_ Event) {})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Emit("test", nil)
	}
}
