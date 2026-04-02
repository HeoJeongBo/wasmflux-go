package util

import (
	"context"
	"fmt"
	"runtime"
	"sync"
)

// Group manages a set of goroutines with shared context and error collection.
// Similar to errgroup.Group but with panic recovery and context propagation.
type Group struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.Mutex
	errs   []error
}

// NewGroup creates a group with a derived context.
// Cancelling the parent context or calling the returned cancel func stops all goroutines.
func NewGroup(ctx context.Context) (*Group, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{ctx: ctx, cancel: cancel}, cancel
}

// Go launches a function in a new goroutine managed by the group.
// Panics are caught and converted to errors.
func (g *Group) Go(fn func(ctx context.Context) error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				g.mu.Lock()
				g.errs = append(g.errs, fmt.Errorf("goroutine panic: %v\n%s", r, buf[:n]))
				g.mu.Unlock()
			}
		}()
		if err := fn(g.ctx); err != nil {
			g.mu.Lock()
			g.errs = append(g.errs, err)
			g.mu.Unlock()
		}
	}()
}

// Wait blocks until all goroutines complete and returns all collected errors.
func (g *Group) Wait() []error {
	g.wg.Wait()
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.errs
}

// WaitFirst blocks until all goroutines complete. Returns the first error, or nil.
func (g *Group) WaitFirst() error {
	g.wg.Wait()
	g.mu.Lock()
	defer g.mu.Unlock()
	if len(g.errs) > 0 {
		return g.errs[0]
	}
	return nil
}

// Context returns the group's context.
func (g *Group) Context() context.Context {
	return g.ctx
}

// Cancel cancels the group's context.
func (g *Group) Cancel() {
	g.cancel()
}
