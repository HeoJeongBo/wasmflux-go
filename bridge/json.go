//go:build js && wasm

package bridge

import (
	"encoding/json"
	"syscall/js"
)

// Marshal converts a Go value to a JS object via JSON serialization.
// Supports structs, maps, slices, and all JSON-compatible types.
func Marshal(v any) (js.Value, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return js.Undefined(), err
	}
	return js.Global().Get("JSON").Call("parse", string(b)), nil
}

// MarshalValue is like Marshal but returns a wrapped Value.
func MarshalValue(v any) (Value, error) {
	jsv, err := Marshal(v)
	if err != nil {
		return Value{}, err
	}
	return WrapValue(jsv), nil
}

// Unmarshal converts a JS value to a Go value via JSON serialization.
// dst must be a pointer to the target type.
func Unmarshal(v js.Value, dst any) error {
	str := js.Global().Get("JSON").Call("stringify", v).String()
	return json.Unmarshal([]byte(str), dst)
}

// UnmarshalValue is like Unmarshal but accepts a wrapped Value.
func UnmarshalValue(v Value, dst any) error {
	return Unmarshal(v.Raw(), dst)
}

// ToMap converts a JS object to a map[string]any.
func ToMap(v js.Value) (map[string]any, error) {
	var m map[string]any
	if err := Unmarshal(v, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// ToSlice converts a JS array to a []any.
func ToSlice(v js.Value) ([]any, error) {
	var s []any
	if err := Unmarshal(v, &s); err != nil {
		return nil, err
	}
	return s, nil
}
