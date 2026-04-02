package ring

import "iter"

// Buffer is a generic ring buffer.
// Data is continuously written; when the cursor reaches the end,
// it overwrites from the beginning. Based on the ring.Slice pattern.
type Buffer[T any] struct {
	vs  []T
	cur int
	len int
	cap int
}

// NewBuffer creates a new ring buffer with the given capacity.
func NewBuffer[T any](capacity int) *Buffer[T] {
	if capacity <= 0 {
		capacity = 1
	}
	return &Buffer[T]{
		vs:  make([]T, capacity),
		cap: capacity,
	}
}

// Write writes a single item to the ring buffer.
func (b *Buffer[T]) Write(v T) {
	b.vs[b.cur] = v
	b.cur = (b.cur + 1) % b.cap
	if b.len < b.cap {
		b.len++
	}
}

// WriteBatch writes multiple items to the ring buffer.
func (b *Buffer[T]) WriteBatch(vs []T) {
	l := len(vs)
	if l >= b.cap {
		copy(b.vs, vs[l-b.cap:])
		b.cur = 0
		b.len = b.cap
		return
	}

	n := copy(b.vs[b.cur:], vs)
	if n >= l {
		b.cur += n
		b.len = min(b.len+n, b.cap)
		return
	}

	m := copy(b.vs, vs[n:])
	b.cur = m
	b.len = b.cap
}

// Read copies data from oldest to newest into dst. Returns the number of items copied.
func (b *Buffer[T]) Read(dst []T) int {
	if b.len < b.cap {
		return copy(dst, b.vs[:b.cur])
	}

	n := copy(dst, b.vs[b.cur:])
	if n >= len(dst) {
		return n
	}

	m := copy(dst[n:], b.vs[:b.cur])
	return n + m
}

// Drain reads all data into dst and clears the buffer. Returns the number of items read.
func (b *Buffer[T]) Drain(dst []T) int {
	n := b.Read(dst)
	b.Clear()
	return n
}

// Values returns all data as a new slice, oldest to newest.
func (b *Buffer[T]) Values() []T {
	result := make([]T, b.Len())
	b.Read(result)
	return result
}

// View iterates over data from oldest to newest.
// If f returns false, iteration stops.
func (b *Buffer[T]) View(f func(T) bool) {
	if b.len == 0 {
		return
	}

	if b.len < b.cap {
		for i := 0; i < b.cur; i++ {
			if !f(b.vs[i]) {
				return
			}
		}
		return
	}

	for i := b.cur; i < b.cap; i++ {
		if !f(b.vs[i]) {
			return
		}
	}
	for i := 0; i < b.cur; i++ {
		if !f(b.vs[i]) {
			return
		}
	}
}

// BackwardView iterates over data from newest to oldest.
func (b *Buffer[T]) BackwardView(f func(T) bool) {
	if b.len == 0 {
		return
	}

	if b.len < b.cap {
		for i := b.cur - 1; i >= 0; i-- {
			if !f(b.vs[i]) {
				return
			}
		}
		return
	}

	for i := b.cur - 1; i >= 0; i-- {
		if !f(b.vs[i]) {
			return
		}
	}
	for i := b.cap - 1; i >= b.cur; i-- {
		if !f(b.vs[i]) {
			return
		}
	}
}

// Iter returns an iterator from oldest to newest.
func (b *Buffer[T]) Iter() iter.Seq[T] {
	return b.View
}

// BackwardIter returns an iterator from newest to oldest.
func (b *Buffer[T]) BackwardIter() iter.Seq[T] {
	return b.BackwardView
}

// Len returns the number of valid items in the buffer.
func (b *Buffer[T]) Len() int {
	return b.len
}

// Cap returns the buffer capacity.
func (b *Buffer[T]) Cap() int {
	return b.cap
}

// Clear resets the buffer to its initial state.
func (b *Buffer[T]) Clear() {
	var zero T
	for i := 0; i < b.cap; i++ {
		b.vs[i] = zero
	}
	b.cur = 0
	b.len = 0
}
