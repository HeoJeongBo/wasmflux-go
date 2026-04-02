//go:build js && wasm

package main

import (
	"syscall/js"

	wasmflux "github.com/heojeongbo/wasmflux-go"
	"github.com/heojeongbo/wasmflux-go/bridge"
	"github.com/heojeongbo/wasmflux-go/flux"
	"github.com/heojeongbo/wasmflux-go/log"
	"github.com/heojeongbo/wasmflux-go/tick"
)

// counterState is the flux store state for the counter module.
type counterState struct {
	Count int     `js:"count"`
	FPS   float64 `js:"fps"`
}

// counterModule manages a counter and FPS tracking.
// Provides: "counter.store" — other modules can inject this to read/subscribe to counter state.
type counterModule struct {
	ctx      wasmflux.ModuleContext
	logger   *log.Logger
	store    *flux.Store[counterState]
	raf      *tick.RAFLoop
	frames   int
	fpsAccum float64
}

func (m *counterModule) Name() string { return "counter" }

func (m *counterModule) Init(ctx wasmflux.ModuleContext) error {
	m.ctx = ctx
	m.logger = ctx.Logger.With(log.String("module", "counter"))

	m.store = flux.NewStore(
		counterState{},
		func(state counterState, action flux.Action) counterState {
			switch action.Type {
			case "increment":
				state.Count++
			case "decrement":
				state.Count--
			case "add":
				state.Count += action.Payload.(int)
			case "set_fps":
				state.FPS = action.Payload.(float64)
			}
			return state
		},
		flux.LoggerMiddleware(m.logger),
	)

	// Provide store so signal/compute modules can inject it.
	ctx.Provide("counter.store", m.store)

	// Bridge: actions.
	ctx.Bridge.ExposeSimple("increment", func(_ []js.Value) {
		m.store.Dispatch(flux.NewAction("increment", nil))
	})
	ctx.Bridge.ExposeSimple("decrement", func(_ []js.Value) {
		m.store.Dispatch(flux.NewAction("decrement", nil))
	})
	ctx.Bridge.Expose("add", func(_ js.Value, args []js.Value) any {
		n, err := bridge.ArgInt(args, 0)
		if err != nil {
			m.logger.Error("add: invalid arg", log.Err(err))
			return nil
		}
		m.store.Dispatch(flux.NewAction("add", n))
		return nil
	})

	// Bridge: state access.
	ctx.Bridge.Expose("getState", func(_ js.Value, _ []js.Value) any {
		v, _ := bridge.Encode(m.store.GetState())
		return v
	})
	ctx.Bridge.Expose("subscribeCounter", func(_ js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		cb := args[0]
		m.store.Subscribe(func(s counterState) {
			v, _ := bridge.Encode(s)
			cb.Invoke(v)
		})
		return nil
	})

	return nil
}

func (m *counterModule) Start() error {
	m.raf = tick.NewRAFLoop(func(dt float64) {
		m.frames++
		m.fpsAccum += dt
		if m.fpsAccum >= 1000 {
			fps := float64(m.frames) / (m.fpsAccum / 1000)
			m.store.Dispatch(flux.NewAction("set_fps", fps))
			m.frames = 0
			m.fpsAccum = 0
		}
	})
	m.raf.Start()
	m.logger.Info("counter module started")
	return nil
}

func (m *counterModule) Stop() error {
	if m.raf != nil {
		m.raf.Release()
	}
	return nil
}
