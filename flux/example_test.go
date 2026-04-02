package flux_test

import (
	"fmt"

	"github.com/heojeongbo/wasmflux-go/flux"
)

func ExampleStore() {
	type State struct {
		Count int
	}

	store := flux.NewStore(
		State{Count: 0},
		func(state State, action flux.Action) State {
			switch action.Type {
			case "increment":
				state.Count++
			case "add":
				state.Count += action.Payload.(int)
			}
			return state
		},
	)

	store.Subscribe(func(s State) {
		fmt.Println("count:", s.Count)
	})

	store.Dispatch(flux.NewAction("increment", nil))
	store.Dispatch(flux.NewAction("add", 10))
	// Output:
	// count: 1
	// count: 11
}

func ExampleStore_SubscribeDelta() {
	type State struct {
		Value int
	}

	store := flux.NewStore(
		State{Value: 0},
		func(state State, action flux.Action) State {
			state.Value = action.Payload.(int)
			return state
		},
	)

	store.SubscribeDelta(func(old, new State) {
		fmt.Printf("%d -> %d\n", old.Value, new.Value)
	})

	store.Dispatch(flux.NewAction("set", 42))
	store.Dispatch(flux.NewAction("set", 100))
	// Output:
	// 0 -> 42
	// 42 -> 100
}
