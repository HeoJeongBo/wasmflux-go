package flux

import (
	"testing"

	"github.com/heojeongbo/wasmflux-go/log"
)

type testState struct {
	Count int
}

func testReducer(state testState, action Action) testState {
	switch action.Type {
	case "increment":
		state.Count++
	case "decrement":
		state.Count--
	case "add":
		state.Count += action.Payload.(int)
	}
	return state
}

func TestStore_Dispatch(t *testing.T) {
	store := NewStore(testState{Count: 0}, testReducer)

	store.Dispatch(NewAction("increment", nil))
	store.Dispatch(NewAction("increment", nil))
	store.Dispatch(NewAction("decrement", nil))

	state := store.GetState()
	if state.Count != 1 {
		t.Errorf("count: %d, want 1", state.Count)
	}
}

func TestStore_Subscribe(t *testing.T) {
	store := NewStore(testState{Count: 0}, testReducer)

	var lastState testState
	store.Subscribe(func(s testState) {
		lastState = s
	})

	store.Dispatch(NewAction("increment", nil))

	if lastState.Count != 1 {
		t.Errorf("subscriber count: %d, want 1", lastState.Count)
	}
}

func TestStore_Unsubscribe(t *testing.T) {
	store := NewStore(testState{Count: 0}, testReducer)

	callCount := 0
	unsub := store.Subscribe(func(_ testState) {
		callCount++
	})

	store.Dispatch(NewAction("increment", nil))
	unsub()
	store.Dispatch(NewAction("increment", nil))

	if callCount != 1 {
		t.Errorf("call count: %d, want 1", callCount)
	}
}

func TestStore_WithPayload(t *testing.T) {
	store := NewStore(testState{Count: 0}, testReducer)

	store.Dispatch(NewAction("add", 10))

	state := store.GetState()
	if state.Count != 10 {
		t.Errorf("count: %d, want 10", state.Count)
	}
}

func TestStore_SubscribeDelta(t *testing.T) {
	store := NewStore(testState{Count: 0}, testReducer)

	var oldCount, newCount int
	store.SubscribeDelta(func(old, new testState) {
		oldCount = old.Count
		newCount = new.Count
	})

	store.Dispatch(NewAction("increment", nil))
	if oldCount != 0 || newCount != 1 {
		t.Errorf("delta: %d→%d, want 0→1", oldCount, newCount)
	}
}

func TestStore_SubscriberCount(t *testing.T) {
	store := NewStore(testState{}, testReducer)
	if store.SubscriberCount() != 0 {
		t.Error("expected 0 subscribers")
	}

	unsub := store.Subscribe(func(_ testState) {})
	store.SubscribeDelta(func(_, _ testState) {})
	if store.SubscriberCount() != 2 {
		t.Errorf("count: %d, want 2", store.SubscriberCount())
	}

	unsub()
	if store.SubscriberCount() != 1 {
		t.Errorf("count after unsub: %d, want 1", store.SubscriberCount())
	}
}

func TestStore_DispatchCount(t *testing.T) {
	store := NewStore(testState{}, testReducer)
	store.Dispatch(NewAction("increment", nil))
	store.Dispatch(NewAction("increment", nil))
	store.Dispatch(NewAction("increment", nil))

	if store.DispatchCount() != 3 {
		t.Errorf("dispatch count: %d, want 3", store.DispatchCount())
	}
}

func TestStore_Middleware(t *testing.T) {
	var logged []string
	mockLogger := func(next func(Action)) func(Action) {
		return func(a Action) {
			logged = append(logged, a.Type)
			next(a)
		}
	}

	store := NewStore(testState{}, testReducer, mockLogger)
	store.Dispatch(NewAction("increment", nil))
	store.Dispatch(NewAction("add", 5))

	if len(logged) != 2 {
		t.Fatalf("logged: %d, want 2", len(logged))
	}
	if logged[0] != "increment" || logged[1] != "add" {
		t.Errorf("logged: %v", logged)
	}
}

func TestStore_ConcurrentDispatch(t *testing.T) {
	store := NewStore(testState{Count: 0}, testReducer)

	done := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func() {
			store.Dispatch(NewAction("increment", nil))
			done <- struct{}{}
		}()
	}
	for i := 0; i < 100; i++ {
		<-done
	}

	if store.GetState().Count != 100 {
		t.Errorf("count: %d, want 100", store.GetState().Count)
	}
}

func TestStore_LoggerMiddleware(t *testing.T) {
	w := &logCapture{}
	logger := log.NewLogger(w, log.LevelDebug)

	store := NewStore(testState{}, testReducer, LoggerMiddleware(logger))
	store.Dispatch(NewAction("increment", nil))

	if len(w.msgs) == 0 {
		t.Error("LoggerMiddleware should have logged")
	}
}

type logCapture struct {
	msgs []string
}

func (w *logCapture) WriteLog(_ log.Level, msg string) {
	w.msgs = append(w.msgs, msg)
}

func TestStore_ThunkMiddleware(t *testing.T) {
	store := NewStore(testState{Count: 0}, testReducer, ThunkMiddleware())

	// Dispatch a thunk that dispatches two actions.
	store.Dispatch(NewAction("thunk", func(dispatch func(Action)) {
		dispatch(NewAction("increment", nil))
		dispatch(NewAction("add", 10))
	}))

	state := store.GetState()
	if state.Count != 11 {
		t.Errorf("count: %d, want 11", state.Count)
	}
}

func TestStore_ThunkMiddleware_RegularAction(t *testing.T) {
	store := NewStore(testState{Count: 0}, testReducer, ThunkMiddleware())

	// Regular action should pass through.
	store.Dispatch(NewAction("increment", nil))
	if store.GetState().Count != 1 {
		t.Errorf("count: %d, want 1", store.GetState().Count)
	}
}

func TestStore_SubscribeDelta_Unsubscribe(t *testing.T) {
	store := NewStore(testState{Count: 0}, testReducer)

	count := 0
	unsub := store.SubscribeDelta(func(_, _ testState) {
		count++
	})

	store.Dispatch(NewAction("increment", nil))
	unsub()
	store.Dispatch(NewAction("increment", nil))

	if count != 1 {
		t.Errorf("count: %d, want 1 (unsubscribed after first)", count)
	}
}

func BenchmarkStore_Dispatch(b *testing.B) {
	store := NewStore(testState{Count: 0}, testReducer)
	action := NewAction("increment", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Dispatch(action)
	}
}
