package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Writer is the interface for log output destinations.
type Writer interface {
	WriteLog(level Level, msg string)
}

// WriterFunc adapts a function to the Writer interface.
type WriterFunc func(level Level, msg string)

func (f WriterFunc) WriteLog(level Level, msg string) { f(level, msg) }

// StdWriter writes logs to an io.Writer (defaults to os.Stderr).
// Used for native Go testing and non-WASM environments.
type StdWriter struct {
	Out io.Writer
}

func (w *StdWriter) WriteLog(level Level, msg string) {
	out := w.Out
	if out == nil {
		out = os.Stderr
	}
	fmt.Fprintf(out, "%s\n", msg)
}

// Logger provides structured logging with zero-alloc hot paths.
// It formats messages into a pre-allocated buffer and outputs
// through a Writer (console.log bridge in WASM, stderr in native).
type Logger struct {
	writer    Writer
	level     Level
	prefix    []Field // inherited fields from With()
	buf       []byte  // scratch buffer, reused across calls
	formatter Formatter
	addCaller bool

	// Rate limiting fields.
	limiter   *logRateLimiter
}

// logRateLimiter controls per-message-key rate limiting.
type logRateLimiter struct {
	mu      sync.Mutex
	counts  map[string]int
	max     int
	window  time.Duration
	resetAt time.Time
}

func newLogRateLimiter(max int, window time.Duration) *logRateLimiter {
	return &logRateLimiter{
		counts:  make(map[string]int),
		max:     max,
		window:  window,
		resetAt: time.Now().Add(window),
	}
}

func (r *logRateLimiter) allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	if now.After(r.resetAt) {
		r.counts = make(map[string]int)
		r.resetAt = now.Add(r.window)
	}
	r.counts[key]++
	return r.counts[key] <= r.max
}

// LoggerOption configures a Logger.
type LoggerOption func(*Logger)

// WithFormatter sets a custom formatter for the logger.
func WithFormatter(f Formatter) LoggerOption {
	return func(l *Logger) { l.formatter = f }
}

// WithCaller adds caller information (file:line) to log messages.
func WithCaller() LoggerOption {
	return func(l *Logger) { l.addCaller = true }
}

// WithRateLimit limits logs to max entries per message per window.
func WithRateLimit(max int, window time.Duration) LoggerOption {
	return func(l *Logger) { l.limiter = newLogRateLimiter(max, window) }
}

// NewLogger creates a logger with the given writer and minimum level.
func NewLogger(w Writer, level Level, opts ...LoggerOption) *Logger {
	l := &Logger{
		writer:    w,
		level:     level,
		buf:       make([]byte, 0, 256),
		formatter: &TextFormatter{},
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// NewStdLogger creates a logger that writes to stderr.
func NewStdLogger(level Level, opts ...LoggerOption) *Logger {
	return NewLogger(&StdWriter{Out: os.Stderr}, level, opts...)
}

// With returns a child logger that includes the given fields in every message.
func (l *Logger) With(fields ...Field) *Logger {
	child := &Logger{
		writer:    l.writer,
		level:     l.level,
		prefix:    make([]Field, 0, len(l.prefix)+len(fields)),
		buf:       make([]byte, 0, 256),
		formatter: l.formatter,
		addCaller: l.addCaller,
		limiter:   l.limiter,
	}
	child.prefix = append(child.prefix, l.prefix...)
	child.prefix = append(child.prefix, fields...)
	return child
}

// SetLevel changes the minimum log level.
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// GetLevel returns the current minimum log level.
func (l *Logger) GetLevel() Level {
	return l.level
}

// Debug logs at debug level.
func (l *Logger) Debug(msg string, fields ...Field) {
	if l.level > LevelDebug {
		return
	}
	l.log(LevelDebug, msg, fields)
}

// Info logs at info level.
func (l *Logger) Info(msg string, fields ...Field) {
	if l.level > LevelInfo {
		return
	}
	l.log(LevelInfo, msg, fields)
}

// Warn logs at warn level.
func (l *Logger) Warn(msg string, fields ...Field) {
	if l.level > LevelWarn {
		return
	}
	l.log(LevelWarn, msg, fields)
}

// Error logs at error level.
func (l *Logger) Error(msg string, fields ...Field) {
	l.log(LevelError, msg, fields)
}

// log formats and writes a log message.
func (l *Logger) log(level Level, msg string, fields []Field) {
	if l.limiter != nil && !l.limiter.allow(msg) {
		return
	}

	allFields := l.prefix
	if len(fields) > 0 {
		allFields = append(l.prefix, fields...)
	}

	if l.addCaller {
		allFields = append(allFields, String("caller", callerInfo(3)))
	}

	if l.formatter != nil {
		formatted := l.formatter.Format(level, msg, allFields)
		l.writer.WriteLog(level, formatted)
		return
	}

	// Fallback: direct formatting using scratch buffer.
	l.buf = l.buf[:0]
	l.buf = append(l.buf, msg...)

	if len(allFields) > 0 {
		l.buf = append(l.buf, ' ')
		for i := range allFields {
			if i > 0 {
				l.buf = append(l.buf, ' ')
			}
			l.buf = append(l.buf, allFields[i].Key...)
			l.buf = append(l.buf, '=')
			l.buf = allFields[i].appendValue(l.buf)
		}
	}

	l.writer.WriteLog(level, string(l.buf))
}
