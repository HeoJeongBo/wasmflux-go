package log

import "strconv"

// Formatter formats a log entry into a string.
type Formatter interface {
	Format(level Level, msg string, fields []Field) string
}

// TextFormatter formats logs as key=value text (default).
type TextFormatter struct {
	// WithTimestamp adds a timestamp prefix if true.
	WithTimestamp bool
	// TimestampFunc returns the current timestamp string.
	// If nil, uses a simple incrementing counter.
	TimestampFunc func() string
}

func (f *TextFormatter) Format(level Level, msg string, fields []Field) string {
	buf := make([]byte, 0, 128)

	if f.WithTimestamp {
		buf = append(buf, '[')
		if f.TimestampFunc != nil {
			buf = append(buf, f.TimestampFunc()...)
		}
		buf = append(buf, "] "...)
	}

	buf = append(buf, '[')
	buf = append(buf, level.String()...)
	buf = append(buf, "] "...)
	buf = append(buf, msg...)

	if len(fields) > 0 {
		buf = append(buf, ' ')
		for i := range fields {
			if i > 0 {
				buf = append(buf, ' ')
			}
			buf = append(buf, fields[i].Key...)
			buf = append(buf, '=')
			buf = fields[i].appendValue(buf)
		}
	}

	return string(buf)
}

// JSONFormatter formats logs as JSON objects.
type JSONFormatter struct {
	// WithTimestamp adds a "ts" field if true.
	WithTimestamp bool
	// TimestampFunc returns the current timestamp string.
	TimestampFunc func() string
}

func (f *JSONFormatter) Format(level Level, msg string, fields []Field) string {
	buf := make([]byte, 0, 256)
	buf = append(buf, `{"level":"`...)
	buf = append(buf, level.LowerString()...)
	buf = append(buf, `","msg":`...)
	buf = appendJSONString(buf, msg)

	if f.WithTimestamp && f.TimestampFunc != nil {
		buf = append(buf, `,"ts":"`...)
		buf = append(buf, f.TimestampFunc()...)
		buf = append(buf, '"')
	}

	for i := range fields {
		buf = append(buf, ',')
		buf = appendJSONString(buf, fields[i].Key)
		buf = append(buf, ':')
		switch fields[i].Kind {
		case FieldString:
			buf = appendJSONString(buf, fields[i].Str)
		case FieldInt:
			buf = strconv.AppendInt(buf, fields[i].Int, 10)
		case FieldFloat:
			buf = strconv.AppendFloat(buf, fields[i].Float, 'f', -1, 64)
		case FieldBool:
			if fields[i].Int == 1 {
				buf = append(buf, "true"...)
			} else {
				buf = append(buf, "false"...)
			}
		}
	}

	buf = append(buf, '}')
	return string(buf)
}

func appendJSONString(buf []byte, s string) []byte {
	buf = append(buf, '"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			buf = append(buf, '\\', '"')
		case '\\':
			buf = append(buf, '\\', '\\')
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		default:
			if c < 0x20 {
				buf = append(buf, '\\', 'u', '0', '0')
				buf = append(buf, "0123456789abcdef"[c>>4])
				buf = append(buf, "0123456789abcdef"[c&0xf])
			} else {
				buf = append(buf, c)
			}
		}
	}
	buf = append(buf, '"')
	return buf
}
