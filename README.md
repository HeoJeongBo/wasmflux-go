# wasmflux-go

A high-performance Go + WebAssembly framework designed for real-time data streams at 60Hz+.

**Features:**
- Module system with dependency injection (Provide/Inject)
- Type-safe Go↔JS bridge with reflection-based Encode/Decode (`js` struct tags)
- Flux-style generic state management with middleware
- Zero-alloc ring buffers, object pools, batch processors
- Structured logger with JSON/Text formatters, caller info, rate limiting
- Event bus, RAF loops, debounce/throttle, retry, goroutine groups
- TypeScript SDK class (no `window` global dependency)
- Vite + React example with full integration

## Quick Start

```bash
# Setup
git clone https://github.com/heojeongbo/wasmflux-go.git
cd wasmflux-go
make setup

# Build & run
make build
make serve
# Open http://localhost:8080

# Run the React example
cd example && npm install && cd ..
make setup-example
make example
# Open http://localhost:5173
```

## Architecture

```
wasmflux-go/
├── app.go, module.go, registry.go    App lifecycle, module DI
├── bridge/                           Go↔JS interop, Encode/Decode, Promise
├── event/                            Topic-based event bus
├── flux/                             Generic Flux state management
├── log/                              Structured logger (Text/JSON)
├── pool/                             Generic object pool (sync.Pool)
├── ring/                             Ring buffer for streaming data
├── batch/                            Batch processor
├── tick/                             RAF loop, interval/timeout scheduler
├── errors/                           Error handling with stack traces
├── util/                             Debounce, Throttle, Retry, RateLimiter, Group
├── internal/jsutil/                  DOM, Fetch, Storage, TypedArray helpers
├── internal/debug/                   Debug tools (build tag: debug)
├── example/                          Go modules + Vite React example
└── cmd/devserver/                    Dev HTTP server
```

## Module System

Modules are the building blocks of a wasmflux app. Each module has a lifecycle (`Init` → `Start` → `Stop`) and can share services via dependency injection.

```go
type counterModule struct {
    ctx   wasmflux.ModuleContext
    store *flux.Store[counterState]
}

func (m *counterModule) Name() string { return "counter" }

func (m *counterModule) Init(ctx wasmflux.ModuleContext) error {
    m.ctx = ctx
    m.store = flux.NewStore(counterState{}, reducer)

    // Register service for other modules to inject.
    ctx.Provide("counter.store", m.store)

    // Expose function to JS.
    ctx.Bridge.Expose("getState", func(_ js.Value, _ []js.Value) any {
        v, _ := bridge.Encode(m.store.GetState())
        return v
    })
    return nil
}

func (m *counterModule) Start() error {
    // Inject services from other modules (all Init calls completed).
    signalStore, ok := wasmflux.InjectAs[*flux.Store[signalState]](m.ctx, "signal.store")
    // ...
    return nil
}

func (m *counterModule) Stop() error { return nil }
```

Register modules in order (DI dependencies must be registered first):

```go
func main() {
    app := wasmflux.New(wasmflux.WithLogLevel(log.LevelDebug))
    app.Register(&counterModule{})
    app.Register(&signalModule{})
    app.Register(&computeModule{})
    app.Run()
}
```

## Bridge: Go↔JS Type Conversion

Reflection-based Encode/Decode with `js` struct tags. No JSON round-trip.

```go
type State struct {
    Count int       `js:"count"`
    Name  string    `js:"name"`
    Items []float64 `js:"items"`
}

// Go → JS (direct Object creation)
jsVal, err := bridge.Encode(state)

// JS → Go (direct property reading)
var state State
err := bridge.Decode(jsVal, &state)

// Type-safe argument extraction
n, err := bridge.ArgInt(args, 0)
s, err := bridge.ArgString(args, 1)

// Async computation via goroutine → JS Promise
bridge.NewPromise(func(resolve, reject func(any)) {
    go func() {
        result := heavyComputation()
        v, _ := bridge.Encode(result)
        resolve(v)
    }()
})
```

Tag priority: `js:"name"` > `json:"name"` > `lowercase(FieldName)`

## TypeScript SDK

The SDK encapsulates `window.wasmflux` internally. Components access WASM through a typed instance:

```typescript
import { WasmFlux } from "./lib/wasmflux"

// Initialize once
const sdk = await WasmFlux.init("/app.wasm")

// Type-safe method calls
sdk.increment()
sdk.add(10)
const state = sdk.getCounterState()

// Subscribe to state changes (push model, not polling)
const unsub = sdk.on("counter", (state) => {
  console.log(state.count, state.fps)
})

// Async computation
const result = await sdk.computeAsync(40)

// Cleanup
sdk.shutdown()
```

React integration with Context:

```tsx
function App() {
  const { status, sdk } = useWasmInit("/app.wasm")
  if (status !== "ready") return <div>Loading...</div>

  return (
    <WasmFluxProvider value={sdk}>
      <Counter />
    </WasmFluxProvider>
  )
}

function Counter() {
  const { state, increment, decrement } = useCounterState()
  return <div>{state.count}</div>
}
```

## Benchmarks

Measured on Apple M4 Pro, Go 1.25.4, `GOARCH=arm64`.

### Core Packages

| Package | Operation | Speed | Alloc |
|---------|-----------|-------|-------|
| ring/Buffer | Write | 2.7 ns/op | 0 |
| ring/Buffer | WriteBatch (60) | 6.7 ns/op | 0 |
| event/Bus | Emit | 7.5 ns/op | 0 |
| flux/Store | Dispatch | 9.1 ns/op | 0 |
| batch/Processor | Push | 3.1 ns/op | 0 |
| pool/ByteBuffer | Get/Put | 7.3 ns/op | 0 |
| log/Logger | Info (2 fields) | 70 ns/op | 2 allocs |
| log/Logger | Debug (skipped) | 0.23 ns/op | 0 |

### DI Overhead

| Operation | Speed | Comparison |
|-----------|-------|------------|
| Direct field access | 0.23 ns/op | baseline |
| Registry.Inject | 6.3 ns/op | ~27x (called once at Init/Start) |
| Inject + type assert | 6.3 ns/op | same as Inject |
| Inject (100 services) | 7.1 ns/op | independent of service count |
| Registry.Provide | 18 ns/op | called once per service |

> DI is only used during module Init/Start. At runtime (60Hz hot path), modules hold direct references to injected services with zero overhead.

### WASM Binary

```
WASM size: 2.8 MB (with -ldflags="-s -w")
```

## Test Coverage

All tests pass with `-race` detector enabled.

| Package | Coverage |
|---------|----------|
| wasmflux (root) | 100% |
| pool/ | 100% |
| flux/ | 100% |
| log/ | 98.3% |
| event/ | 98.1% |
| util/ | 97.9% |
| errors/ | 97.4% |
| batch/ | 96.9% |
| bridge/ | 96.2% |
| ring/ | 95.7% |

## Makefile

```bash
make build           # Build WASM binary
make build-debug     # Build with debug tag
make serve           # Build + start dev server
make test            # Run tests with race detector
make bench           # Run benchmarks
make coverage        # Generate coverage report
make check           # fmt + vet + lint + test
make watch           # Auto-rebuild on file changes
make size            # Print WASM binary size
make setup-example   # Copy wasm_exec.js + glue.js to example/
make build-example   # Build WASM for example
make example         # Build + start Vite dev server
```

## License

MIT
