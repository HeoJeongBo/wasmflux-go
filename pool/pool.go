package pool

import "sync"

// Pool is a generic wrapper around sync.Pool.
// It provides type-safe Get/Put operations and avoids interface{} boxing at call sites.
type Pool[T any] struct {
	p sync.Pool
}

// New creates a Pool with the given constructor function.
// The constructor is called when the pool is empty and Get is called.
func New[T any](fn func() T) *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{
			New: func() any { return fn() },
		},
	}
}

// Get retrieves an item from the pool, or creates one if the pool is empty.
func (p *Pool[T]) Get() T {
	return p.p.Get().(T)
}

// Put returns an item to the pool for reuse.
func (p *Pool[T]) Put(v T) {
	p.p.Put(v)
}
