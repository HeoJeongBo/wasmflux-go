//go:build js && wasm

package bridge

import "syscall/js"

// Value provides typed accessors around js.Value to reduce boilerplate
// and prevent type assertion panics at call sites.
type Value struct {
	v js.Value
}

// WrapValue wraps a raw js.Value.
func WrapValue(v js.Value) Value {
	return Value{v: v}
}

// Raw returns the underlying js.Value.
func (v Value) Raw() js.Value {
	return v.v
}

// IsUndefined reports whether the value is JS undefined.
func (v Value) IsUndefined() bool {
	return v.v.IsUndefined()
}

// IsNull reports whether the value is JS null.
func (v Value) IsNull() bool {
	return v.v.IsNull()
}

// IsNullish reports whether the value is JS null or undefined.
func (v Value) IsNullish() bool {
	return v.v.IsUndefined() || v.v.IsNull()
}

// String returns the value as a Go string. Returns "" if undefined/null.
func (v Value) String() string {
	if v.IsNullish() {
		return ""
	}
	return v.v.String()
}

// Int returns the value as int. Returns 0 if undefined/null.
func (v Value) Int() int {
	if v.IsNullish() {
		return 0
	}
	return v.v.Int()
}

// Float returns the value as float64. Returns 0 if undefined/null.
func (v Value) Float() float64 {
	if v.IsNullish() {
		return 0
	}
	return v.v.Float()
}

// Bool returns the value as bool. Returns false if undefined/null.
func (v Value) Bool() bool {
	if v.IsNullish() {
		return false
	}
	return v.v.Bool()
}

// Get retrieves a property as a wrapped Value.
func (v Value) Get(name string) Value {
	return Value{v: v.v.Get(name)}
}

// Index retrieves an array element as a wrapped Value.
func (v Value) Index(i int) Value {
	return Value{v: v.v.Index(i)}
}

// Length returns the length property.
func (v Value) Length() int {
	return v.v.Length()
}

// Call calls a method and returns a wrapped Value.
func (v Value) Call(name string, args ...any) Value {
	return Value{v: v.v.Call(name, args...)}
}

// Set sets a property on the JS object.
func (v Value) Set(name string, val any) {
	v.v.Set(name, val)
}
