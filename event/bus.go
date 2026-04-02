package event

import "sync"

// Bus is a high-performance topic-based event dispatcher.
// Handlers are stored in slices per topic for fast dispatch on the hot path.
// Uses RWMutex: readers (Emit) can run concurrently, writers (On) are exclusive.
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]entry
	nextID   uint64
}

type entry struct {
	id uint64
	fn Handler
}

// NewBus creates a new event bus.
func NewBus() *Bus {
	return &Bus{
		handlers: make(map[string][]entry),
	}
}

// On registers a handler for the given topic.
// Returns an unsubscribe function.
func (b *Bus) On(topic string, h Handler) func() {
	b.mu.Lock()
	id := b.nextID
	b.nextID++
	b.handlers[topic] = append(b.handlers[topic], entry{id: id, fn: h})
	b.mu.Unlock()

	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		entries := b.handlers[topic]
		for i, e := range entries {
			if e.id == id {
				b.handlers[topic] = append(entries[:i], entries[i+1:]...)
				return
			}
		}
	}
}

// Emit dispatches an event synchronously to all handlers for the topic.
// This is the hot path — uses RLock for concurrent read access.
func (b *Bus) Emit(topic string, data any) {
	b.mu.RLock()
	entries := b.handlers[topic]
	b.mu.RUnlock()

	if len(entries) == 0 {
		return
	}

	evt := Event{
		Topic: topic,
		Data:  data,
	}
	for _, e := range entries {
		e.fn(evt)
	}
}

// EmitWithTimestamp dispatches an event with an explicit timestamp.
// Useful when the timestamp comes from the JS side (performance.now()).
func (b *Bus) EmitWithTimestamp(topic string, data any, timestamp float64) {
	b.mu.RLock()
	entries := b.handlers[topic]
	b.mu.RUnlock()

	if len(entries) == 0 {
		return
	}

	evt := Event{
		Topic:     topic,
		Timestamp: timestamp,
		Data:      data,
	}
	for _, e := range entries {
		e.fn(evt)
	}
}

// EmitAsync dispatches an event asynchronously in a goroutine.
func (b *Bus) EmitAsync(topic string, data any) {
	go b.Emit(topic, data)
}

// HasHandlers reports whether any handlers are registered for the topic.
func (b *Bus) HasHandlers(topic string) bool {
	b.mu.RLock()
	n := len(b.handlers[topic])
	b.mu.RUnlock()
	return n > 0
}

// Clear removes all handlers for all topics.
func (b *Bus) Clear() {
	b.mu.Lock()
	b.handlers = make(map[string][]entry)
	b.mu.Unlock()
}

// Topics returns a list of all topics with registered handlers.
func (b *Bus) Topics() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	topics := make([]string, 0, len(b.handlers))
	for t, entries := range b.handlers {
		if len(entries) > 0 {
			topics = append(topics, t)
		}
	}
	return topics
}

// HandlerCount returns the number of handlers registered for a topic.
func (b *Bus) HandlerCount(topic string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.handlers[topic])
}

// Once registers a handler that fires only once, then auto-unsubscribes.
func (b *Bus) Once(topic string, h Handler) {
	var unsub func()
	unsub = b.On(topic, func(e Event) {
		h(e)
		unsub()
	})
}
