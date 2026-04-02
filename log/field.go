package log

import "strconv"

// FieldKind discriminates the value type stored in a Field.
// Using a kind tag instead of interface{} avoids boxing on hot paths.
type FieldKind uint8

const (
	FieldString FieldKind = iota
	FieldInt
	FieldFloat
	FieldBool
)

// Field is a pre-allocated key-value pair for structured logging.
// Uses a union-style struct to avoid interface{} allocation.
type Field struct {
	Key   string
	Str   string
	Int   int64
	Float float64
	Kind  FieldKind
}

// String creates a string field.
func String(key, val string) Field {
	return Field{Key: key, Str: val, Kind: FieldString}
}

// Int creates an int64 field.
func Int(key string, val int64) Field {
	return Field{Key: key, Int: val, Kind: FieldInt}
}

// Float creates a float64 field.
func Float(key string, val float64) Field {
	return Field{Key: key, Float: val, Kind: FieldFloat}
}

// Bool creates a bool field.
func Bool(key string, val bool) Field {
	f := Field{Key: key, Kind: FieldBool}
	if val {
		f.Int = 1
	}
	return f
}

// Err creates a field from an error.
func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Str: "<nil>", Kind: FieldString}
	}
	return Field{Key: "error", Str: err.Error(), Kind: FieldString}
}

// appendValue appends the field's value to a byte slice without allocation.
func (f *Field) appendValue(dst []byte) []byte {
	switch f.Kind {
	case FieldString:
		dst = append(dst, f.Str...)
	case FieldInt:
		dst = strconv.AppendInt(dst, f.Int, 10)
	case FieldFloat:
		dst = strconv.AppendFloat(dst, f.Float, 'f', -1, 64)
	case FieldBool:
		if f.Int == 1 {
			dst = append(dst, "true"...)
		} else {
			dst = append(dst, "false"...)
		}
	}
	return dst
}
