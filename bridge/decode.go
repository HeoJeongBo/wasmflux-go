//go:build js && wasm

package bridge

import (
	"fmt"
	"reflect"
	"syscall/js"
	"time"
)

// DecodeError reports a type mismatch during JS → Go decoding,
// including the full field path for debugging.
type DecodeError struct {
	Path     string // e.g. "state.items[2].name"
	Expected string // e.g. "string"
	Got      string // e.g. "number"
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("decode %s: expected %s, got %s", e.Path, e.Expected, e.Got)
}

// Decode converts a JS value into a Go value using reflection and struct tags.
// dst must be a non-nil pointer. Uses js/json struct tags for field name mapping.
// Unlike Unmarshal, this reads js.Value properties directly without JSON round-trip.
func Decode(v js.Value, dst any) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("decode: dst must be a non-nil pointer, got %T", dst)
	}
	return decodeValue(v, rv.Elem(), "")
}

func decodeValue(v js.Value, rv reflect.Value, path string) error {
	// Handle null/undefined: set to zero value.
	if v.IsUndefined() || v.IsNull() {
		rv.Set(reflect.Zero(rv.Type()))
		return nil
	}

	// Handle pointer types.
	if rv.Kind() == reflect.Ptr {
		if v.IsUndefined() || v.IsNull() {
			rv.Set(reflect.Zero(rv.Type()))
			return nil
		}
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return decodeValue(v, rv.Elem(), path)
	}

	// Handle js.Value escape hatch.
	if rv.Type() == reflect.TypeOf(js.Value{}) {
		rv.Set(reflect.ValueOf(v))
		return nil
	}

	// Handle time.Time.
	if rv.Type() == reflect.TypeOf(time.Time{}) {
		if v.Type() != js.TypeString {
			return &DecodeError{Path: path, Expected: "string (RFC3339)", Got: jsTypeName(v)}
		}
		t, err := time.Parse(time.RFC3339, v.String())
		if err != nil {
			return &DecodeError{Path: path, Expected: "RFC3339 datetime", Got: v.String()}
		}
		rv.Set(reflect.ValueOf(t))
		return nil
	}

	switch rv.Kind() {
	case reflect.Bool:
		if v.Type() != js.TypeBoolean {
			return &DecodeError{Path: path, Expected: "boolean", Got: jsTypeName(v)}
		}
		rv.SetBool(v.Bool())

	case reflect.String:
		if v.Type() != js.TypeString {
			return &DecodeError{Path: path, Expected: "string", Got: jsTypeName(v)}
		}
		rv.SetString(v.String())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Type() != js.TypeNumber {
			return &DecodeError{Path: path, Expected: "number", Got: jsTypeName(v)}
		}
		rv.SetInt(int64(v.Float()))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Type() != js.TypeNumber {
			return &DecodeError{Path: path, Expected: "number", Got: jsTypeName(v)}
		}
		f := v.Float()
		if f < 0 {
			return &DecodeError{Path: path, Expected: "non-negative number", Got: fmt.Sprintf("%g", f)}
		}
		rv.SetUint(uint64(f))

	case reflect.Float32, reflect.Float64:
		if v.Type() != js.TypeNumber {
			return &DecodeError{Path: path, Expected: "number", Got: jsTypeName(v)}
		}
		rv.SetFloat(v.Float())

	case reflect.Struct:
		if v.Type() != js.TypeObject {
			return &DecodeError{Path: path, Expected: "object", Got: jsTypeName(v)}
		}
		return decodeStruct(v, rv, path)

	case reflect.Slice:
		if v.Type() != js.TypeObject {
			return &DecodeError{Path: path, Expected: "array", Got: jsTypeName(v)}
		}
		return decodeSlice(v, rv, path)

	case reflect.Map:
		if v.Type() != js.TypeObject {
			return &DecodeError{Path: path, Expected: "object", Got: jsTypeName(v)}
		}
		return decodeMap(v, rv, path)

	case reflect.Interface:
		// For interface{} / any, store as basic Go types.
		rv.Set(reflect.ValueOf(jsToGo(v)))

	default:
		return fmt.Errorf("decode %s: unsupported Go type %s", path, rv.Type())
	}

	return nil
}

func decodeStruct(v js.Value, rv reflect.Value, path string) error {
	fields := getStructInfo(rv.Type())
	for _, fi := range fields {
		prop := v.Get(fi.jsName)
		fieldPath := joinPath(path, fi.jsName)
		fv := rv.FieldByIndex(fi.index)
		if err := decodeValue(prop, fv, fieldPath); err != nil {
			return err
		}
	}
	return nil
}

func decodeSlice(v js.Value, rv reflect.Value, path string) error {
	n := v.Length()
	slice := reflect.MakeSlice(rv.Type(), n, n)
	for i := 0; i < n; i++ {
		elemPath := fmt.Sprintf("%s[%d]", path, i)
		if err := decodeValue(v.Index(i), slice.Index(i), elemPath); err != nil {
			return err
		}
	}
	rv.Set(slice)
	return nil
}

func decodeMap(v js.Value, rv reflect.Value, path string) error {
	if rv.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("decode %s: map key must be string, got %s", path, rv.Type().Key())
	}

	keys := js.Global().Get("Object").Call("keys", v)
	n := keys.Length()
	m := reflect.MakeMapWithSize(rv.Type(), n)
	elemType := rv.Type().Elem()

	for i := 0; i < n; i++ {
		key := keys.Index(i).String()
		elem := reflect.New(elemType).Elem()
		keyPath := joinPath(path, key)
		if err := decodeValue(v.Get(key), elem, keyPath); err != nil {
			return err
		}
		m.SetMapIndex(reflect.ValueOf(key), elem)
	}
	rv.Set(m)
	return nil
}

// jsToGo converts a js.Value to the closest Go native type.
// Used for interface{} / any fields.
func jsToGo(v js.Value) any {
	switch v.Type() {
	case js.TypeBoolean:
		return v.Bool()
	case js.TypeNumber:
		return v.Float()
	case js.TypeString:
		return v.String()
	case js.TypeNull, js.TypeUndefined:
		return nil
	default:
		return v
	}
}

func jsTypeName(v js.Value) string {
	switch v.Type() {
	case js.TypeBoolean:
		return "boolean"
	case js.TypeNumber:
		return "number"
	case js.TypeString:
		return "string"
	case js.TypeObject:
		return "object"
	case js.TypeFunction:
		return "function"
	case js.TypeNull:
		return "null"
	case js.TypeUndefined:
		return "undefined"
	case js.TypeSymbol:
		return "symbol"
	default:
		return "unknown"
	}
}

func joinPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}
