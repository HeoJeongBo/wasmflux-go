package log_test

import (
	"fmt"

	"github.com/heojeongbo/wasmflux-go/log"
)

type printWriter struct{}

func (w *printWriter) WriteLog(_ log.Level, msg string) {
	fmt.Println(msg)
}

func ExampleLogger() {
	l := log.NewLogger(&printWriter{}, log.LevelInfo)

	l.Info("server started", log.String("addr", ":8080"), log.Int("workers", 4))
	l.Debug("this is skipped") // below LevelInfo
	// Output: [INFO] server started addr=:8080 workers=4
}

func ExampleLogger_With() {
	l := log.NewLogger(&printWriter{}, log.LevelDebug)
	child := l.With(log.String("module", "auth"))

	child.Debug("checking token")
	// Output: [DEBUG] checking token module=auth
}

func ExampleLogger_json() {
	l := log.NewLogger(&printWriter{}, log.LevelInfo, log.WithFormatter(&log.JSONFormatter{}))

	l.Info("request", log.String("method", "GET"), log.Int("status", 200))
	// Output: {"level":"info","msg":"request","method":"GET","status":200}
}
