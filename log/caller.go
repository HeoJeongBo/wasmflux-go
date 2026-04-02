package log

import (
	"runtime"
	"strconv"
	"strings"
)

// callerInfo captures the file and line of the calling function.
// skip is the number of stack frames to skip.
func callerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "???:0"
	}
	// Shorten to last two path components for readability.
	if idx := lastNSlash(file, 2); idx >= 0 {
		file = file[idx+1:]
	}
	return file + ":" + strconv.Itoa(line)
}

func lastNSlash(s string, n int) int {
	count := 0
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' {
			count++
			if count == n {
				return i
			}
		}
	}
	return -1
}

// stackTrace captures a full stack trace, skipping the first skip frames.
func stackTrace(skip int) string {
	var sb strings.Builder
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		sb.WriteString(frame.Function)
		sb.WriteByte('\n')
		sb.WriteByte('\t')
		sb.WriteString(frame.File)
		sb.WriteByte(':')
		sb.WriteString(strconv.Itoa(frame.Line))
		sb.WriteByte('\n')
		if !more {
			break
		}
	}
	return sb.String()
}
