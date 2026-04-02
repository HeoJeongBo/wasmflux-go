package bridge

import (
	"reflect"
	"testing"
)

type basicStruct struct {
	Name  string `js:"name"`
	Count int    `js:"count"`
}

type jsonFallbackStruct struct {
	Name  string `json:"user_name"`
	Count int    `json:"total_count"`
}

type noTagStruct struct {
	Name  string
	Count int
}

type mixedTagStruct struct {
	JSField   string `js:"js_field"`
	JSONField string `json:"json_field"`
	NoTag     string
}

type skipFieldStruct struct {
	Visible string `js:"visible"`
	Hidden  string `js:"-"`
	Skipped string `json:"-"`
}

type omitEmptyStruct struct {
	Name  string `js:"name,omitempty"`
	Count int    `json:"count,omitempty"`
}

type nestedStruct struct {
	Inner basicStruct `js:"inner"`
	Value int         `js:"value"`
}

type embeddedStruct struct {
	basicStruct
	Extra string `js:"extra"`
}

type pointerFieldStruct struct {
	Name  *string `js:"name"`
	Inner *basicStruct `js:"inner"`
}

type unexportedStruct struct {
	Public  string `js:"public"`
	private string //nolint
}

func TestGetStructInfo_JSTags(t *testing.T) {
	fields := getStructInfo(reflect.TypeOf(basicStruct{}))
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}
	if fields[0].jsName != "name" {
		t.Errorf("field 0: got %q, want %q", fields[0].jsName, "name")
	}
	if fields[1].jsName != "count" {
		t.Errorf("field 1: got %q, want %q", fields[1].jsName, "count")
	}
}

func TestGetStructInfo_JSONFallback(t *testing.T) {
	fields := getStructInfo(reflect.TypeOf(jsonFallbackStruct{}))
	if fields[0].jsName != "user_name" {
		t.Errorf("got %q, want %q", fields[0].jsName, "user_name")
	}
	if fields[1].jsName != "total_count" {
		t.Errorf("got %q, want %q", fields[1].jsName, "total_count")
	}
}

func TestGetStructInfo_NoTag(t *testing.T) {
	fields := getStructInfo(reflect.TypeOf(noTagStruct{}))
	if fields[0].jsName != "name" {
		t.Errorf("got %q, want %q", fields[0].jsName, "name")
	}
	if fields[1].jsName != "count" {
		t.Errorf("got %q, want %q", fields[1].jsName, "count")
	}
}

func TestGetStructInfo_MixedTags(t *testing.T) {
	fields := getStructInfo(reflect.TypeOf(mixedTagStruct{}))
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(fields))
	}
	if fields[0].jsName != "js_field" {
		t.Errorf("field 0: got %q, want %q", fields[0].jsName, "js_field")
	}
	if fields[1].jsName != "json_field" {
		t.Errorf("field 1: got %q, want %q", fields[1].jsName, "json_field")
	}
	if fields[2].jsName != "notag" {
		t.Errorf("field 2: got %q, want %q", fields[2].jsName, "notag")
	}
}

func TestGetStructInfo_SkipFields(t *testing.T) {
	fields := getStructInfo(reflect.TypeOf(skipFieldStruct{}))
	if len(fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(fields))
	}
	if fields[0].jsName != "visible" {
		t.Errorf("got %q, want %q", fields[0].jsName, "visible")
	}
}

func TestGetStructInfo_OmitEmpty(t *testing.T) {
	fields := getStructInfo(reflect.TypeOf(omitEmptyStruct{}))
	if !fields[0].omitEmpty {
		t.Error("field 0 should have omitEmpty=true")
	}
	if !fields[1].omitEmpty {
		t.Error("field 1 should have omitEmpty=true")
	}
}

func TestGetStructInfo_Embedded(t *testing.T) {
	fields := getStructInfo(reflect.TypeOf(embeddedStruct{}))
	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = f.jsName
	}
	// Should flatten embedded basicStruct fields + Extra.
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d: %v", len(fields), names)
	}
	if fields[0].jsName != "name" {
		t.Errorf("field 0: got %q, want %q", fields[0].jsName, "name")
	}
	if fields[1].jsName != "count" {
		t.Errorf("field 1: got %q, want %q", fields[1].jsName, "count")
	}
	if fields[2].jsName != "extra" {
		t.Errorf("field 2: got %q, want %q", fields[2].jsName, "extra")
	}
}

func TestGetStructInfo_UnexportedSkipped(t *testing.T) {
	fields := getStructInfo(reflect.TypeOf(unexportedStruct{}))
	if len(fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(fields))
	}
	if fields[0].jsName != "public" {
		t.Errorf("got %q, want %q", fields[0].jsName, "public")
	}
}

func TestGetStructInfo_Cache(t *testing.T) {
	typ := reflect.TypeOf(basicStruct{})

	// Clear any prior cache entry for deterministic test.
	structInfoCache.Delete(typ)

	fields1 := getStructInfo(typ)
	fields2 := getStructInfo(typ)

	// Should be the same slice from cache.
	if &fields1[0] != &fields2[0] {
		t.Error("expected same cached slice on second call")
	}
}

func TestGetStructInfo_Pointer(t *testing.T) {
	fields := getStructInfo(reflect.TypeOf((*basicStruct)(nil)))
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields from pointer type, got %d", len(fields))
	}
}

func TestResolveTag_Priority(t *testing.T) {
	type both struct {
		Field string `js:"js_name" json:"json_name"`
	}
	fields := getStructInfo(reflect.TypeOf(both{}))
	if fields[0].jsName != "js_name" {
		t.Errorf("got %q, want %q (js should win)", fields[0].jsName, "js_name")
	}
}

func TestGetStructInfo_DeepNested(t *testing.T) {
	type inner struct {
		X int `js:"x"`
	}
	type middle struct {
		inner
		Y int `js:"y"`
	}
	type outer struct {
		middle
		Z int `js:"z"`
	}

	fields := getStructInfo(reflect.TypeOf(outer{}))
	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = f.jsName
	}
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d: %v", len(fields), names)
	}
	if names[0] != "x" || names[1] != "y" || names[2] != "z" {
		t.Errorf("names: %v, want [x y z]", names)
	}
}

func TestGetStructInfo_ConcurrentAccess(t *testing.T) {
	type concStruct struct {
		A string `js:"a"`
		B int    `js:"b"`
	}
	typ := reflect.TypeOf(concStruct{})
	structInfoCache.Delete(typ)

	done := make(chan []fieldInfo, 100)
	for i := 0; i < 100; i++ {
		go func() {
			done <- getStructInfo(typ)
		}()
	}
	var first []fieldInfo
	for i := 0; i < 100; i++ {
		result := <-done
		if first == nil {
			first = result
		}
		if len(result) != 2 {
			t.Errorf("expected 2 fields, got %d", len(result))
		}
	}
}

func TestGetStructInfo_NonStruct(t *testing.T) {
	fields := getStructInfo(reflect.TypeOf(42))
	if fields != nil {
		t.Errorf("expected nil for non-struct, got %v", fields)
	}
}
