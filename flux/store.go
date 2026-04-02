package flux

import "sync"

// Store is a generic flux-like state container.
// S is the state type. Actions are dispatched through a reducer.
// Subscribers are notified synchronously after state changes.
type Store[S any] struct {
	mu               sync.RWMutex
	state            S
	reducer          func(state S, action Action) S
	subscribers      []subscriber[S]
	deltaSubscribers []deltaSubscriber[S]
	dispatch         func(Action) // the middleware-wrapped dispatch chain
	nextID           uint64
	nextDeltaID      uint64
	dispatchCount    uint64
}

type subscriber[S any] struct {
	id uint64
	fn func(S)
}

type deltaSubscriber[S any] struct {
	id uint64
	fn func(old, new S)
}

// NewStore creates a store with an initial state and a reducer function.
func NewStore[S any](initial S, reducer func(S, Action) S, middlewares ...Middleware) *Store[S] {
	s := &Store[S]{
		state:   initial,
		reducer: reducer,
	}

	// Build the middleware chain.
	base := func(a Action) {
		s.mu.Lock()
		oldState := s.state
		s.state = s.reducer(s.state, a)
		newState := s.state
		s.dispatchCount++
		subs := make([]subscriber[S], len(s.subscribers))
		copy(subs, s.subscribers)
		deltaSubs := make([]deltaSubscriber[S], len(s.deltaSubscribers))
		copy(deltaSubs, s.deltaSubscribers)
		s.mu.Unlock()

		for _, sub := range subs {
			sub.fn(newState)
		}
		for _, sub := range deltaSubs {
			sub.fn(oldState, newState)
		}
	}

	dispatch := base
	for i := len(middlewares) - 1; i >= 0; i-- {
		dispatch = middlewares[i](dispatch)
	}
	s.dispatch = dispatch

	return s
}

// Dispatch sends an action through the middleware chain and reducer.
func (s *Store[S]) Dispatch(a Action) {
	s.dispatch(a)
}

// GetState returns the current state.
func (s *Store[S]) GetState() S {
	s.mu.RLock()
	state := s.state
	s.mu.RUnlock()
	return state
}

// Subscribe registers a callback invoked after each state change.
// Returns an unsubscribe function.
func (s *Store[S]) Subscribe(fn func(S)) func() {
	s.mu.Lock()
	id := s.nextID
	s.nextID++
	s.subscribers = append(s.subscribers, subscriber[S]{id: id, fn: fn})
	s.mu.Unlock()

	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		for i, sub := range s.subscribers {
			if sub.id == id {
				s.subscribers = append(s.subscribers[:i], s.subscribers[i+1:]...)
				return
			}
		}
	}
}

// SubscribeDelta registers a callback that receives both old and new state.
// Useful for computing deltas or conditional rendering.
func (s *Store[S]) SubscribeDelta(fn func(old, new S)) func() {
	s.mu.Lock()
	id := s.nextDeltaID
	s.nextDeltaID++
	s.deltaSubscribers = append(s.deltaSubscribers, deltaSubscriber[S]{id: id, fn: fn})
	s.mu.Unlock()

	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		for i, sub := range s.deltaSubscribers {
			if sub.id == id {
				s.deltaSubscribers = append(s.deltaSubscribers[:i], s.deltaSubscribers[i+1:]...)
				return
			}
		}
	}
}

// SubscriberCount returns the total number of subscribers (regular + delta).
func (s *Store[S]) SubscriberCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.subscribers) + len(s.deltaSubscribers)
}

// DispatchCount returns the total number of dispatched actions.
func (s *Store[S]) DispatchCount() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dispatchCount
}
