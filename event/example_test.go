package event_test

import (
	"fmt"

	"github.com/heojeongbo/wasmflux-go/event"
)

func ExampleBus() {
	bus := event.NewBus()

	bus.On("greet", func(e event.Event) {
		fmt.Println("Hello,", e.Data)
	})

	bus.Emit("greet", "World")
	// Output: Hello, World
}

func ExampleBus_Once() {
	bus := event.NewBus()

	bus.Once("init", func(e event.Event) {
		fmt.Println("initialized:", e.Data)
	})

	bus.Emit("init", "app")
	bus.Emit("init", "again") // won't fire
	// Output: initialized: app
}

func ExampleBus_unsubscribe() {
	bus := event.NewBus()

	count := 0
	unsub := bus.On("tick", func(_ event.Event) {
		count++
	})

	bus.Emit("tick", nil)
	unsub()
	bus.Emit("tick", nil)

	fmt.Println(count)
	// Output: 1
}
