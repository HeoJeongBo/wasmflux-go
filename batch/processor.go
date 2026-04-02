package batch

import "sync"

// Processor accumulates items and flushes either when the batch
// is full or when Flush is called manually.
// This reduces JS↔Go boundary crossings for high-frequency data.
type Processor[T any] struct {
	mu      sync.Mutex
	buf     []T
	pos     int
	cap     int
	flushFn func(items []T)
}

// NewProcessor creates a batch processor with the given capacity and flush callback.
// The flush function receives the accumulated items when the batch is full.
func NewProcessor[T any](capacity int, flush func(items []T)) *Processor[T] {
	if capacity <= 0 {
		capacity = 64
	}
	return &Processor[T]{
		buf:     make([]T, capacity),
		cap:     capacity,
		flushFn: flush,
	}
}

// Push adds an item to the batch. If the batch is full, it flushes automatically.
func (p *Processor[T]) Push(item T) {
	p.mu.Lock()
	p.buf[p.pos] = item
	p.pos++
	if p.pos >= p.cap {
		items := make([]T, p.pos)
		copy(items, p.buf[:p.pos])
		p.pos = 0
		p.mu.Unlock()
		p.flushFn(items)
		return
	}
	p.mu.Unlock()
}

// PushBatch adds multiple items. May trigger one or more flushes.
func (p *Processor[T]) PushBatch(items []T) {
	for _, item := range items {
		p.Push(item)
	}
}

// Flush manually flushes any accumulated items, even if the batch is not full.
func (p *Processor[T]) Flush() {
	p.mu.Lock()
	if p.pos == 0 {
		p.mu.Unlock()
		return
	}
	items := make([]T, p.pos)
	copy(items, p.buf[:p.pos])
	p.pos = 0
	p.mu.Unlock()
	p.flushFn(items)
}

// Len returns the number of items currently in the batch.
func (p *Processor[T]) Len() int {
	p.mu.Lock()
	n := p.pos
	p.mu.Unlock()
	return n
}

// Reset discards all accumulated items without flushing.
func (p *Processor[T]) Reset() {
	p.mu.Lock()
	p.pos = 0
	p.mu.Unlock()
}
