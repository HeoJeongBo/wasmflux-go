package bridge

import (
	"reflect"
	"strings"
	"sync"
)

// fieldInfo holds cached metadata about a struct field for JS ↔ Go mapping.
type fieldInfo struct {
	index     []int        // reflect field index (supports embedded structs)
	jsName    string       // property name on the JS side
	fieldType reflect.Type // Go type of the field
	omitEmpty bool         // skip zero values during encoding
}

// structInfoCache maps reflect.Type → []fieldInfo to avoid repeated reflection.
var structInfoCache sync.Map

// getStructInfo returns cached field metadata for the given struct type.
// Tag resolution priority: js:"name" > json:"name" > lowercase(GoFieldName).
func getStructInfo(t reflect.Type) []fieldInfo {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	if cached, ok := structInfoCache.Load(t); ok {
		return cached.([]fieldInfo)
	}

	fields := collectFields(t, nil)
	structInfoCache.Store(t, fields)
	return fields
}

// collectFields recursively collects field info, flattening embedded structs.
func collectFields(t reflect.Type, parentIndex []int) []fieldInfo {
	var fields []fieldInfo

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		index := make([]int, len(parentIndex)+1)
		copy(index, parentIndex)
		index[len(parentIndex)] = i

		// Handle embedded (anonymous) struct fields — flatten them.
		// Anonymous fields are promoted even if the type name is unexported.
		if f.Anonymous {
			ft := f.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				fields = append(fields, collectFields(ft, index)...)
				continue
			}
		}

		// Skip unexported non-anonymous fields.
		if !f.IsExported() {
			continue
		}

		name, omitEmpty, skip := resolveTag(f)
		if skip {
			continue
		}

		fields = append(fields, fieldInfo{
			index:     index,
			jsName:    name,
			fieldType: f.Type,
			omitEmpty: omitEmpty,
		})
	}

	return fields
}

// resolveTag extracts the JS property name from struct tags.
// Priority: js tag > json tag > lowercase field name.
// Returns (name, omitEmpty, skip).
func resolveTag(f reflect.StructField) (string, bool, bool) {
	// Check js tag first.
	if tag, ok := f.Tag.Lookup("js"); ok {
		name, opts := parseTag(tag)
		if name == "-" {
			return "", false, true
		}
		if name != "" {
			return name, containsOmitEmpty(opts), false
		}
	}

	// Fall back to json tag.
	if tag, ok := f.Tag.Lookup("json"); ok {
		name, opts := parseTag(tag)
		if name == "-" {
			return "", false, true
		}
		if name != "" {
			return name, containsOmitEmpty(opts), false
		}
	}

	// Fall back to lowercase field name.
	return strings.ToLower(f.Name), false, false
}

// parseTag splits a struct tag value into name and remaining options.
func parseTag(tag string) (string, string) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tag[idx+1:]
	}
	return tag, ""
}

func containsOmitEmpty(opts string) bool {
	for opts != "" {
		var name string
		if i := strings.Index(opts, ","); i >= 0 {
			name, opts = opts[:i], opts[i+1:]
		} else {
			name, opts = opts, ""
		}
		if name == "omitempty" {
			return true
		}
	}
	return false
}
