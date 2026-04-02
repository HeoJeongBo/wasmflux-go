//go:build js && wasm

package tick

import "sync"

// Scheduler manages multiple tickers (RAF loops and intervals) with unified lifecycle.
type Scheduler struct {
	mu        sync.Mutex
	rafs      []*RAFLoop
	intervals []*Interval
}

// NewScheduler creates a new scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// AddRAF creates and registers a new RAF loop. It is started immediately.
func (s *Scheduler) AddRAF(fn func(dt float64)) *RAFLoop {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := NewRAFLoop(fn)
	s.rafs = append(s.rafs, r)
	r.Start()
	return r
}

// AddInterval creates and registers a new interval.
func (s *Scheduler) AddInterval(fn func(), intervalMs float64) *Interval {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := NewInterval(fn, intervalMs)
	s.intervals = append(s.intervals, t)
	return t
}

// StopAll stops all managed tickers.
func (s *Scheduler) StopAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.rafs {
		r.Stop()
	}
	for _, t := range s.intervals {
		t.Stop()
	}
}

// Release stops and releases all managed tickers.
func (s *Scheduler) Release() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.rafs {
		r.Release()
	}
	for _, t := range s.intervals {
		t.Release()
	}
	s.rafs = nil
	s.intervals = nil
}
