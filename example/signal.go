//go:build js && wasm

package main

import (
	"math"
	"syscall/js"

	wasmflux "github.com/heojeongbo/wasmflux-go"
	"github.com/heojeongbo/wasmflux-go/batch"
	"github.com/heojeongbo/wasmflux-go/bridge"
	"github.com/heojeongbo/wasmflux-go/event"
	"github.com/heojeongbo/wasmflux-go/flux"
	"github.com/heojeongbo/wasmflux-go/log"
	"github.com/heojeongbo/wasmflux-go/ring"
	"github.com/heojeongbo/wasmflux-go/tick"
)

// signalState holds real-time signal processing results.
type signalState struct {
	Mean   float64   `js:"mean"`
	StdDev float64   `js:"stdDev"`
	Min    float64   `js:"min"`
	Max    float64   `js:"max"`
	Count  int       `js:"count"`
	Recent []float64 `js:"recent"`
}

// signalModule processes high-frequency data streams.
// Injects: "counter.store" — listens to counter changes to generate synthetic signals.
// Provides: "signal.store" — other modules can subscribe to signal analysis results.
type signalModule struct {
	ctx      wasmflux.ModuleContext
	logger   *log.Logger
	store    *flux.Store[signalState]
	buf      *ring.Buffer[float64]
	batch    *batch.Processor[float64]
	interval *tick.Interval
}

func (m *signalModule) Name() string { return "signal" }

func (m *signalModule) Init(ctx wasmflux.ModuleContext) error {
	m.ctx = ctx
	m.logger = ctx.Logger.With(log.String("module", "signal"))

	m.store = flux.NewStore(signalState{}, func(state signalState, action flux.Action) signalState {
		if action.Type == "signal:update" {
			return action.Payload.(signalState)
		}
		return state
	})
	m.buf = ring.NewBuffer[float64](256)
	m.batch = batch.NewProcessor[float64](32, func(items []float64) {
		for _, v := range items {
			m.buf.Write(v)
		}
		m.analyze()
	})

	ctx.Provide("signal.store", m.store)

	// Bridge: data input.
	ctx.Bridge.ExposeSimple("pushSignal", func(args []js.Value) {
		if len(args) > 0 {
			m.batch.Push(args[0].Float())
		}
	})
	ctx.Bridge.ExposeSimple("pushSignalBatch", func(args []js.Value) {
		if len(args) == 0 {
			return
		}
		arr := args[0]
		n := arr.Length()
		for i := 0; i < n; i++ {
			m.batch.Push(arr.Index(i).Float())
		}
	})

	// Bridge: async analysis in goroutine.
	ctx.Bridge.Expose("analyzeAsync", func(_ js.Value, _ []js.Value) any {
		return bridge.NewPromise(func(resolve, reject func(any)) {
			go func() {
				v, err := bridge.Encode(m.computeStats())
				if err != nil {
					reject(err.Error())
					return
				}
				resolve(v)
			}()
		})
	})

	// Bridge: state access.
	ctx.Bridge.Expose("getSignalState", func(_ js.Value, _ []js.Value) any {
		v, _ := bridge.Encode(m.store.GetState())
		return v
	})
	ctx.Bridge.Expose("subscribeSignal", func(_ js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		cb := args[0]
		m.store.Subscribe(func(s signalState) {
			v, _ := bridge.Encode(s)
			cb.Invoke(v)
		})
		return nil
	})

	return nil
}

func (m *signalModule) Start() error {
	// Inject counter store — cross-module DI.
	counterStore, ok := wasmflux.InjectAs[*flux.Store[counterState]](m.ctx, "counter.store")
	if ok {
		counterStore.Subscribe(func(s counterState) {
			m.ctx.Bus.Emit("counter:changed", s.Count)
		})
		m.logger.Info("injected counter.store")
	}

	// Generate synthetic signal data on counter change.
	m.ctx.Bus.On("counter:changed", func(e event.Event) {
		count := e.Data.(int)
		amplitude := float64(count) * 0.1
		for i := 0; i < 8; i++ {
			m.batch.Push(amplitude * math.Sin(float64(i)*0.5))
		}
	})

	// Auto-generate demo signal data every 100ms.
	m.interval = tick.NewInterval(func() {
		t := float64(m.buf.Len())
		for i := 0; i < 4; i++ {
			m.batch.Push(math.Sin(t*0.05) + math.Sin(t*0.13)*0.3)
		}
	}, 100)

	m.logger.Info("signal module started")
	return nil
}

func (m *signalModule) Stop() error {
	if m.interval != nil {
		m.interval.Release()
	}
	m.batch.Flush()
	return nil
}

func (m *signalModule) analyze() {
	m.store.Dispatch(flux.NewAction("signal:update", m.computeStats()))
}

func (m *signalModule) computeStats() signalState {
	data := m.buf.Values()
	n := len(data)
	if n == 0 {
		return signalState{}
	}

	sum, minVal, maxVal := 0.0, data[0], data[0]
	for _, v := range data {
		sum += v
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	mean := sum / float64(n)

	sumSq := 0.0
	for _, v := range data {
		d := v - mean
		sumSq += d * d
	}

	recent := data
	if len(recent) > 60 {
		recent = recent[len(recent)-60:]
	}

	round := func(v float64) float64 { return math.Round(v*1000) / 1000 }
	return signalState{
		Mean:   round(mean),
		StdDev: round(math.Sqrt(sumSq / float64(n))),
		Min:    round(minVal),
		Max:    round(maxVal),
		Count:  n,
		Recent: recent,
	}
}
