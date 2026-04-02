package event

// Event represents a dispatched event with topic and payload.
type Event struct {
	Topic     string
	Timestamp float64 // milliseconds (performance.now() on JS side)
	Data      any
}

// Handler is a function that processes an event.
type Handler func(Event)
