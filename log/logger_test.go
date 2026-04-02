package log

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

type mockWriter struct {
	messages []struct {
		level Level
		msg   string
	}
}

func (w *mockWriter) WriteLog(level Level, msg string) {
	w.messages = append(w.messages, struct {
		level Level
		msg   string
	}{level, msg})
}

func TestLogger_Levels(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelInfo)

	l.Debug("should be skipped")
	l.Info("info message")
	l.Warn("warn message")
	l.Error("error message")

	if len(w.messages) != 3 {
		t.Fatalf("message count: %d, want 3", len(w.messages))
	}
	if w.messages[0].level != LevelInfo {
		t.Errorf("first level: %v, want Info", w.messages[0].level)
	}
}

func TestLogger_Fields(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelDebug)

	l.Info("test", String("key", "val"), Int("n", 42))

	if len(w.messages) != 1 {
		t.Fatalf("message count: %d, want 1", len(w.messages))
	}
	want := "[INFO] test key=val n=42"
	if w.messages[0].msg != want {
		t.Errorf("msg: %q, want %q", w.messages[0].msg, want)
	}
}

func TestLogger_With(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelDebug)
	child := l.With(String("module", "test"))

	child.Info("hello")

	want := "[INFO] hello module=test"
	if w.messages[0].msg != want {
		t.Errorf("msg: %q, want %q", w.messages[0].msg, want)
	}
}

func TestLogger_ErrField(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelDebug)

	l.Error("failed", Err(nil))
	l.Error("failed", Err(errors.New("boom")))

	if w.messages[0].msg != "[ERROR] failed error=<nil>" {
		t.Errorf("msg: %q", w.messages[0].msg)
	}
	if w.messages[1].msg != "[ERROR] failed error=boom" {
		t.Errorf("msg: %q", w.messages[1].msg)
	}
}

func TestLogger_BoolField(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelDebug)

	l.Info("test", Bool("flag", true))

	want := "[INFO] test flag=true"
	if w.messages[0].msg != want {
		t.Errorf("msg: %q, want %q", w.messages[0].msg, want)
	}
}

func TestLogger_JSONFormatter(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelDebug, WithFormatter(&JSONFormatter{}))

	l.Info("request", String("method", "GET"), Int("status", 200))

	want := `{"level":"info","msg":"request","method":"GET","status":200}`
	if w.messages[0].msg != want {
		t.Errorf("msg: %q, want %q", w.messages[0].msg, want)
	}
}

func TestLogger_RateLimit(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelDebug, WithRateLimit(2, 1<<30)) // 2 per huge window

	for i := 0; i < 10; i++ {
		l.Info("spam")
	}

	if len(w.messages) != 2 {
		t.Errorf("message count: %d, want 2", len(w.messages))
	}
}

func TestLogger_WithCaller(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelDebug, WithCaller())

	l.Info("test")

	if len(w.messages) != 1 {
		t.Fatalf("message count: %d", len(w.messages))
	}
	msg := w.messages[0].msg
	// Should contain caller=<file>:<line>.
	if !contains(msg, "caller=") {
		t.Errorf("expected caller info, got %q", msg)
	}
	if !contains(msg, "logger_test.go") {
		t.Errorf("expected test file in caller, got %q", msg)
	}
}

func TestLogger_TextFormatterWithTimestamp(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelDebug, WithFormatter(&TextFormatter{
		WithTimestamp: true,
		TimestampFunc: func() string { return "12:00:00" },
	}))

	l.Info("hello")

	msg := w.messages[0].msg
	if !contains(msg, "[12:00:00]") {
		t.Errorf("expected timestamp, got %q", msg)
	}
}

func TestLogger_JSONFormatterEscaping(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelDebug, WithFormatter(&JSONFormatter{}))

	l.Info("test", String("val", `he said "hello"`))

	msg := w.messages[0].msg
	if !contains(msg, `\"hello\"`) {
		t.Errorf("expected escaped quotes, got %q", msg)
	}
}

func TestLogger_JSONFormatterTimestamp(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelDebug, WithFormatter(&JSONFormatter{
		WithTimestamp: true,
		TimestampFunc: func() string { return "2026-01-01T00:00:00Z" },
	}))

	l.Info("test")

	msg := w.messages[0].msg
	if !contains(msg, `"ts":"2026-01-01T00:00:00Z"`) {
		t.Errorf("expected timestamp in JSON, got %q", msg)
	}
}

func TestLogger_FloatField(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelDebug)

	l.Info("test", Float("pi", 3.14))

	want := "[INFO] test pi=3.14"
	if w.messages[0].msg != want {
		t.Errorf("got %q, want %q", w.messages[0].msg, want)
	}
}

func TestLogger_StdWriter(t *testing.T) {
	var buf bytes.Buffer
	w := &StdWriter{Out: &buf}
	w.WriteLog(LevelInfo, "hello")
	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("StdWriter output: %q", buf.String())
	}
}

func TestLogger_StdWriter_NilOut(t *testing.T) {
	w := &StdWriter{}
	// Should write to stderr without panic.
	w.WriteLog(LevelInfo, "test")
}

func TestNewStdLogger(t *testing.T) {
	l := NewStdLogger(LevelWarn)
	if l.GetLevel() != LevelWarn {
		t.Errorf("level: %v, want Warn", l.GetLevel())
	}
}

func TestLogger_SetLevel_GetLevel(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelInfo)

	if l.GetLevel() != LevelInfo {
		t.Errorf("initial level: %v", l.GetLevel())
	}
	l.SetLevel(LevelError)
	if l.GetLevel() != LevelError {
		t.Errorf("after set: %v", l.GetLevel())
	}

	l.Info("should be skipped")
	l.Warn("should be skipped")
	l.Error("should appear")

	if len(w.messages) != 1 {
		t.Errorf("message count: %d, want 1", len(w.messages))
	}
}

func TestLogger_Warn(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelDebug)

	l.Warn("warning")
	if len(w.messages) != 1 {
		t.Fatal("expected 1 message")
	}
	if w.messages[0].level != LevelWarn {
		t.Errorf("level: %v, want Warn", w.messages[0].level)
	}
}

func TestLogger_WarnSkipped(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelError)

	l.Warn("skipped")
	if len(w.messages) != 0 {
		t.Error("Warn should be skipped at Error level")
	}
}

func TestLogger_InfoSkipped(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelWarn)

	l.Info("skipped")
	if len(w.messages) != 0 {
		t.Error("Info should be skipped at Warn level")
	}
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestLevel_LowerString(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{LevelDebug, "debug"},
		{LevelInfo, "info"},
		{LevelWarn, "warn"},
		{LevelError, "error"},
		{Level(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.level.LowerString(); got != tt.want {
			t.Errorf("Level(%d).LowerString() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestLogger_RateLimit_WindowReset(t *testing.T) {
	w := &mockWriter{}
	l := NewLogger(w, LevelDebug, WithRateLimit(1, 50*time.Millisecond))

	l.Info("spam")
	l.Info("spam") // rate limited (same key)

	if len(w.messages) != 1 {
		t.Errorf("count: %d, want 1", len(w.messages))
	}

	// Wait for window to reset.
	time.Sleep(60 * time.Millisecond)
	l.Info("spam")
	if len(w.messages) != 2 {
		t.Errorf("count after reset: %d, want 2", len(w.messages))
	}
}

func TestLogger_NoFormatter_Fallback(t *testing.T) {
	w := &mockWriter{}
	l := &Logger{
		writer: w,
		level:  LevelDebug,
		buf:    make([]byte, 0, 256),
	}

	l.Info("direct", String("k", "v"))
	if len(w.messages) != 1 {
		t.Fatal("expected 1 message")
	}
	want := "direct k=v"
	if w.messages[0].msg != want {
		t.Errorf("got %q, want %q", w.messages[0].msg, want)
	}
}

func TestJSONFormatter_ControlCharEscaping(t *testing.T) {
	f := &JSONFormatter{}
	msg := f.Format(LevelInfo, "test", []Field{String("val", "line1\nline2\ttab")})
	if !strings.Contains(msg, `\n`) || !strings.Contains(msg, `\t`) {
		t.Errorf("expected escaped control chars, got %q", msg)
	}
}

func TestJSONFormatter_WithTimestamp(t *testing.T) {
	f := &JSONFormatter{
		WithTimestamp: true,
		TimestampFunc: func() string { return "NOW" },
	}
	msg := f.Format(LevelDebug, "x", nil)
	if !strings.Contains(msg, `"ts":"NOW"`) {
		t.Errorf("expected timestamp, got %q", msg)
	}
}

func TestTextFormatter_NoTimestamp(t *testing.T) {
	f := &TextFormatter{}
	msg := f.Format(LevelInfo, "hello", nil)
	want := "[INFO] hello"
	if msg != want {
		t.Errorf("got %q, want %q", msg, want)
	}
}

func TestField_AppendValue_AllTypes(t *testing.T) {
	tests := []struct {
		field Field
		want  string
	}{
		{String("k", "hello"), "hello"},
		{Int("k", 42), "42"},
		{Float("k", 3.14), "3.14"},
		{Bool("k", true), "true"},
		{Bool("k", false), "false"},
	}
	for _, tt := range tests {
		buf := tt.field.appendValue(nil)
		if string(buf) != tt.want {
			t.Errorf("%s: got %q, want %q", tt.field.Key, string(buf), tt.want)
		}
	}
}

func TestWriterFunc(t *testing.T) {
	var captured string
	fn := WriterFunc(func(_ Level, msg string) {
		captured = msg
	})
	fn.WriteLog(LevelInfo, "via func")
	if captured != "via func" {
		t.Errorf("got %q", captured)
	}
}

func TestStackTrace(t *testing.T) {
	s := stackTrace(1)
	if s == "" {
		t.Error("stackTrace should not be empty")
	}
	if !strings.Contains(s, "TestStackTrace") {
		t.Errorf("should contain caller, got %q", s)
	}
}

func TestCallerInfo_ShortPath(t *testing.T) {
	// callerInfo with a valid skip should return file:line.
	info := callerInfo(1)
	if info == "" || info == "???:0" {
		t.Errorf("unexpected caller info: %q", info)
	}
}

func TestJSONFormatter_AllFieldTypes(t *testing.T) {
	f := &JSONFormatter{}

	msg := f.Format(LevelWarn, "test", []Field{
		String("s", "hello"),
		Int("i", 42),
		Float("f", 3.14),
		Bool("b", true),
		Bool("b2", false),
	})

	for _, want := range []string{`"s":"hello"`, `"i":42`, `"f":3.14`, `"b":true`, `"b2":false`} {
		if !strings.Contains(msg, want) {
			t.Errorf("missing %s in %q", want, msg)
		}
	}
}

func TestAppendJSONString_ControlChars(t *testing.T) {
	f := &JSONFormatter{}
	msg := f.Format(LevelInfo, "test", []Field{String("v", "a\rb\x01c")})
	if !strings.Contains(msg, `\r`) {
		t.Errorf("should escape \\r, got %q", msg)
	}
	if !strings.Contains(msg, `\u0001`) {
		t.Errorf("should escape control char, got %q", msg)
	}
}

func TestLastNSlash_Short(t *testing.T) {
	// Path with fewer slashes than requested.
	idx := lastNSlash("file.go", 2)
	if idx != -1 {
		t.Errorf("expected -1 for no slashes, got %d", idx)
	}

	idx = lastNSlash("a/file.go", 2)
	if idx != -1 {
		t.Errorf("expected -1 for 1 slash with n=2, got %d", idx)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func BenchmarkLogger_Info(b *testing.B) {
	w := &mockWriter{}
	l := NewLogger(w, LevelInfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("benchmark", String("key", "val"), Int("n", 42))
	}
}

func BenchmarkLogger_Debug_Skipped(b *testing.B) {
	w := &mockWriter{}
	l := NewLogger(w, LevelInfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("skipped")
	}
}
