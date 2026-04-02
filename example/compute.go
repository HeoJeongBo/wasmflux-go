//go:build js && wasm

package main

import (
	"math/big"
	"syscall/js"
	"time"

	wasmflux "github.com/heojeongbo/wasmflux-go"
	"github.com/heojeongbo/wasmflux-go/bridge"
	"github.com/heojeongbo/wasmflux-go/event"
	"github.com/heojeongbo/wasmflux-go/flux"
	"github.com/heojeongbo/wasmflux-go/internal/jsutil"
	"github.com/heojeongbo/wasmflux-go/log"
)

// computeResult holds the output of a heavy computation.
type computeResult struct {
	Input      int     `js:"input"`
	Fibonacci  int     `js:"fibonacci"`
	PrimeCount int     `js:"primeCount"`
	Elapsed    float64 `js:"elapsed"`
}

// computeModule runs heavy CPU work in goroutines, returning results as Promises.
// Injects: "counter.store" — demonstrates cross-module dependency.
type computeModule struct {
	ctx    wasmflux.ModuleContext
	logger *log.Logger
}

func (m *computeModule) Name() string { return "compute" }

func (m *computeModule) Init(ctx wasmflux.ModuleContext) error {
	m.ctx = ctx
	m.logger = ctx.Logger.With(log.String("module", "compute"))

	// Bridge: computeAsync — goroutine computation, returns Promise.
	ctx.Bridge.Expose("computeAsync", func(_ js.Value, args []js.Value) any {
		n, err := bridge.ArgInt(args, 0)
		if err != nil {
			m.logger.Error("computeAsync: invalid arg", log.Err(err))
			return js.Undefined()
		}

		return bridge.NewPromise(func(resolve, reject func(any)) {
			go func() {
				start := jsutil.Now()
				result := computeResult{
					Input:      n,
					Fibonacci:  fibonacci(n),
					PrimeCount: countPrimes(n),
					Elapsed:    jsutil.Now() - start,
				}

				m.logger.Info("computation done",
					log.Int("input", int64(n)),
					log.Float("elapsed_ms", result.Elapsed),
				)

				m.ctx.Bus.Emit("compute:done", result)

				v, err := bridge.Encode(result)
				if err != nil {
					reject(err.Error())
					return
				}
				resolve(v)
			}()
		})
	})

	// Bridge: subscribe to computation results.
	ctx.Bridge.Expose("subscribeCompute", func(_ js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		cb := args[0]
		ctx.Bus.On("compute:done", func(e event.Event) {
			r := e.Data.(computeResult)
			v, _ := bridge.Encode(r)
			cb.Invoke(v)
		})
		return nil
	})

	return nil
}

func (m *computeModule) Start() error {
	// Inject counter store — cross-module DI.
	counterStore, ok := wasmflux.InjectAs[*flux.Store[counterState]](m.ctx, "counter.store")
	if ok {
		m.logger.Info("injected counter.store",
			log.Int("current_count", int64(counterStore.GetState().Count)),
		)
	}
	m.logger.Info("compute module started")
	return nil
}

func (m *computeModule) Stop() error { return nil }

// fibonacci computes the n-th fibonacci number using big.Int (CPU-heavy for demo).
func fibonacci(n int) int {
	if n <= 0 {
		return 0
	}
	a := big.NewInt(0)
	b := big.NewInt(1)
	for i := 1; i < n; i++ {
		a.Add(a, b)
		a, b = b, a
	}
	if b.IsInt64() {
		return int(b.Int64())
	}
	return int(b.Uint64())
}

// countPrimes counts primes up to n*100 using a sieve (CPU-heavy for demo).
func countPrimes(n int) int {
	if n < 2 {
		return 0
	}
	limit := n * 100
	if limit > 1_000_000 {
		limit = 1_000_000
	}
	_ = time.Now()
	sieve := make([]bool, limit+1)
	for i := 2; i*i <= limit; i++ {
		if !sieve[i] {
			for j := i * i; j <= limit; j += i {
				sieve[j] = true
			}
		}
	}
	count := 0
	for i := 2; i <= limit; i++ {
		if !sieve[i] {
			count++
		}
	}
	return count
}
