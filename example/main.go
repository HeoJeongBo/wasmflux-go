//go:build js && wasm

package main

import (
	"fmt"

	wasmflux "github.com/heojeongbo/wasmflux-go"
	"github.com/heojeongbo/wasmflux-go/log"
)

func main() {
	fmt.Println("wasmflux-go example starting...")

	app := wasmflux.New(
		wasmflux.WithLogLevel(log.LevelDebug),
	)

	// Registration order = DI order: counter → signal → compute.
	app.Register(&counterModule{})
	app.Register(&signalModule{})
	app.Register(&computeModule{})

	app.Run()
}
