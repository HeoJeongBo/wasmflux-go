//go:build js && wasm

package bridge

import (
	"fmt"
	"syscall/js"
)

// ArgError is returned when a JS argument is missing or has wrong type.
type ArgError struct {
	Index    int
	Expected string
	Got      string
}

func (e *ArgError) Error() string {
	return fmt.Sprintf("arg[%d]: expected %s, got %s", e.Index, e.Expected, e.Got)
}

// ArgString safely extracts a string argument at index i.
func ArgString(args []js.Value, i int) (string, error) {
	if i >= len(args) {
		return "", &ArgError{Index: i, Expected: "string", Got: "missing"}
	}
	v := args[i]
	if v.IsUndefined() || v.IsNull() {
		return "", &ArgError{Index: i, Expected: "string", Got: "null/undefined"}
	}
	if v.Type() != js.TypeString {
		return "", &ArgError{Index: i, Expected: "string", Got: v.Type().String()}
	}
	return v.String(), nil
}

// ArgInt safely extracts an int argument at index i.
func ArgInt(args []js.Value, i int) (int, error) {
	if i >= len(args) {
		return 0, &ArgError{Index: i, Expected: "number", Got: "missing"}
	}
	v := args[i]
	if v.IsUndefined() || v.IsNull() {
		return 0, &ArgError{Index: i, Expected: "number", Got: "null/undefined"}
	}
	if v.Type() != js.TypeNumber {
		return 0, &ArgError{Index: i, Expected: "number", Got: v.Type().String()}
	}
	return v.Int(), nil
}

// ArgFloat safely extracts a float64 argument at index i.
func ArgFloat(args []js.Value, i int) (float64, error) {
	if i >= len(args) {
		return 0, &ArgError{Index: i, Expected: "number", Got: "missing"}
	}
	v := args[i]
	if v.IsUndefined() || v.IsNull() {
		return 0, &ArgError{Index: i, Expected: "number", Got: "null/undefined"}
	}
	if v.Type() != js.TypeNumber {
		return 0, &ArgError{Index: i, Expected: "number", Got: v.Type().String()}
	}
	return v.Float(), nil
}

// ArgBool safely extracts a bool argument at index i.
func ArgBool(args []js.Value, i int) (bool, error) {
	if i >= len(args) {
		return false, &ArgError{Index: i, Expected: "boolean", Got: "missing"}
	}
	v := args[i]
	if v.IsUndefined() || v.IsNull() {
		return false, &ArgError{Index: i, Expected: "boolean", Got: "null/undefined"}
	}
	if v.Type() != js.TypeBoolean {
		return false, &ArgError{Index: i, Expected: "boolean", Got: v.Type().String()}
	}
	return v.Bool(), nil
}

// ArgObject safely extracts an object argument at index i.
func ArgObject(args []js.Value, i int) (js.Value, error) {
	if i >= len(args) {
		return js.Undefined(), &ArgError{Index: i, Expected: "object", Got: "missing"}
	}
	v := args[i]
	if v.IsUndefined() || v.IsNull() {
		return js.Undefined(), &ArgError{Index: i, Expected: "object", Got: "null/undefined"}
	}
	if v.Type() != js.TypeObject {
		return js.Undefined(), &ArgError{Index: i, Expected: "object", Got: v.Type().String()}
	}
	return v, nil
}

// ArgOptionalString extracts a string argument, returning fallback if missing.
func ArgOptionalString(args []js.Value, i int, fallback string) string {
	if i >= len(args) || args[i].IsUndefined() || args[i].IsNull() {
		return fallback
	}
	return args[i].String()
}

// ArgOptionalInt extracts an int argument, returning fallback if missing.
func ArgOptionalInt(args []js.Value, i int, fallback int) int {
	if i >= len(args) || args[i].IsUndefined() || args[i].IsNull() {
		return fallback
	}
	return args[i].Int()
}

// ArgOptionalFloat extracts a float64 argument, returning fallback if missing.
func ArgOptionalFloat(args []js.Value, i int, fallback float64) float64 {
	if i >= len(args) || args[i].IsUndefined() || args[i].IsNull() {
		return fallback
	}
	return args[i].Float()
}
