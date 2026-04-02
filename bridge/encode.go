//go:build js && wasm

package bridge

import (
	"fmt"
	"reflect"
	"syscall/js"
	"time"
)

// Encode converts a Go value into a JS value using reflection and struct tags.
// Unlike Marshal, this creates JS objects directly without JSON round-trip.
func Encode(v any) (js.Value, error) {
	if v == nil {
		return js.Null(), nil
	}
	return encodeValue(reflect.ValueOf(v))
}

func encodeValue(rv reflect.Value) (js.Value, error) {
	// Dereference pointers.
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return js.Null(), nil
		}
		rv = rv.Elem()
	}

	// Handle js.Value pass-through.
	if rv.Type() == reflect.TypeOf(js.Value{}) {
		return rv.Interface().(js.Value), nil
	}

	// Handle time.Time.
	if rv.Type() == reflect.TypeOf(time.Time{}) {
		t := rv.Interface().(time.Time)
		return js.ValueOf(t.Format(time.RFC3339)), nil
	}

	switch rv.Kind() {
	case reflect.Bool:
		return js.ValueOf(rv.Bool()), nil

	case reflect.String:
		return js.ValueOf(rv.String()), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return js.ValueOf(rv.Int()), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return js.ValueOf(rv.Uint()), nil

	case reflect.Float32, reflect.Float64:
		return js.ValueOf(rv.Float()), nil

	case reflect.Struct:
		return encodeStruct(rv)

	case reflect.Slice:
		if rv.IsNil() {
			return js.Null(), nil
		}
		return encodeSlice(rv)

	case reflect.Map:
		if rv.IsNil() {
			return js.Null(), nil
		}
		return encodeMap(rv)

	case reflect.Interface:
		if rv.IsNil() {
			return js.Null(), nil
		}
		return encodeValue(rv.Elem())

	default:
		return js.Undefined(), fmt.Errorf("encode: unsupported Go type %s", rv.Type())
	}
}

func encodeStruct(rv reflect.Value) (js.Value, error) {
	obj := js.Global().Get("Object").New()
	fields := getStructInfo(rv.Type())

	for _, fi := range fields {
		fv := rv.FieldByIndex(fi.index)

		// Skip zero values if omitempty.
		if fi.omitEmpty && fv.IsZero() {
			continue
		}

		encoded, err := encodeValue(fv)
		if err != nil {
			return js.Undefined(), fmt.Errorf("encode %s: %w", fi.jsName, err)
		}
		obj.Set(fi.jsName, encoded)
	}

	return obj, nil
}

func encodeSlice(rv reflect.Value) (js.Value, error) {
	n := rv.Len()
	arr := js.Global().Get("Array").New(n)

	for i := 0; i < n; i++ {
		encoded, err := encodeValue(rv.Index(i))
		if err != nil {
			return js.Undefined(), fmt.Errorf("encode [%d]: %w", i, err)
		}
		arr.SetIndex(i, encoded)
	}

	return arr, nil
}

func encodeMap(rv reflect.Value) (js.Value, error) {
	if rv.Type().Key().Kind() != reflect.String {
		return js.Undefined(), fmt.Errorf("encode: map key must be string, got %s", rv.Type().Key())
	}

	obj := js.Global().Get("Object").New()
	iter := rv.MapRange()

	for iter.Next() {
		key := iter.Key().String()
		encoded, err := encodeValue(iter.Value())
		if err != nil {
			return js.Undefined(), fmt.Errorf("encode .%s: %w", key, err)
		}
		obj.Set(key, encoded)
	}

	return obj, nil
}
