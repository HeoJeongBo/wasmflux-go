package flux

// Action represents a dispatched action with a type and payload.
type Action struct {
	Type    string
	Payload any
}

// NewAction creates an action with the given type and payload.
func NewAction(typ string, payload any) Action {
	return Action{Type: typ, Payload: payload}
}
