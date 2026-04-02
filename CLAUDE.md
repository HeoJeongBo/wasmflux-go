# wasmflux-go

High-performance Go + WebAssembly framework for real-time data streams at 60Hz+.

## Structure

- `app.go`, `module.go`, `app_options.go`, `registry.go` — App lifecycle, module DI
- `bridge/` — Go↔JS interop, Encode/Decode (reflection), typed arg validation, callback pooling, Promise
- `event/` — Topic-based event bus (`On`, `Once`, `Topics`)
- `flux/` — Generic Flux state management, delta subscriptions, middleware
- `log/` — Structured logger (Text/JSON formatters, caller info, rate limiting)
- `pool/` — Generic object pool (sync.Pool)
- `ring/` — Ring buffer for high-frequency streaming data
- `batch/` — Batch processor
- `tick/` — RAF loop, interval/timeout scheduler
- `errors/` — WASM context error handling, panic recovery with stack traces
- `util/` — Debounce, Throttle, Retry, RateLimiter, Goroutine Group
- `internal/jsutil/` — DOM manipulation, Fetch/HTTP, LocalStorage/SessionStorage, TypedArray
- `internal/debug/` — Debug-only tools (build tag: debug)
- `example/` — Go modules + Vite React example (counter, signal, compute)
- `cmd/devserver/` — Dev HTTP server

## Build

```bash
make setup         # Copy wasm_exec.js
make build         # Build WASM
make build-debug   # Build with debug tag
make serve         # Start dev server
make test          # Run tests (race detector)
make bench         # Run benchmarks
make coverage      # Generate coverage report
make check         # fmt + vet + lint + test
make watch         # Auto-rebuild on file changes
make size          # Print WASM binary size
make setup-example # Copy wasm_exec.js + glue.js to example/
make build-example # Build WASM for example
make example       # Build + start Vite dev server
```

## Module DI (Dependency Injection)

```go
// Register a service during Init
ctx.Provide("counter.store", m.store)

// Inject a service during Start (after all Init calls complete)
store, ok := wasmflux.InjectAs[*flux.Store[State]](m.ctx, "counter.store")
```

- `Provide()` — called during Init
- `Inject()` / `InjectAs[T]()` — called during Start
- DI is only invoked at Init/Start, zero impact on the 60Hz hot path

## Conventions

- WASM-specific code: `//go:build js && wasm` build tag
- Pure Go packages (event, flux, log, pool, ring, batch, util, registry): no build tag, natively testable
- Functional options pattern (`WithXxx`)
- Generic types (`Pool[T]`, `Store[S]`, `Buffer[T]`)
- Minimize allocations on hot paths
- `internal/` packages are not exported
- Debug features: stripped in release via `//go:build debug`
- Godoc: exported type/function comments start with the name
- Logger fields: `log.String()`, `log.Int()`, `log.Float()`
- Bridge args: `bridge.ArgString()`, `bridge.ArgInt()` for type-safe extraction
- Bridge conversion: `bridge.Encode(struct)` / `bridge.Decode(jsValue, &struct)` — no JSON round-trip
- Module DI: `Provide` in Init, `Inject` in Start (registration order = DI order)
