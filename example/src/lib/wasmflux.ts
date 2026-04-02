// ---------------------------------------------------------------------------
// Types — mirror Go structs
// ---------------------------------------------------------------------------

export interface CounterState {
  count: number
  fps: number
}

export interface SignalState {
  mean: number
  stdDev: number
  min: number
  max: number
  count: number
  recent: number[]
}

export interface ComputeResult {
  input: number
  fibonacci: number
  primeCount: number
  elapsed: number
}

// ---------------------------------------------------------------------------
// Internal bridge type (what Go exposes on the namespace)
// ---------------------------------------------------------------------------

interface Bridge {
  // Counter
  increment(): void
  decrement(): void
  add(n: number): void
  getState(): CounterState
  subscribeCounter(cb: (state: CounterState) => void): void

  // Signal
  pushSignal(value: number): void
  pushSignalBatch(values: Float64Array): void
  analyzeAsync(): Promise<SignalState>
  getSignalState(): SignalState
  subscribeSignal(cb: (state: SignalState) => void): void

  // Compute
  computeAsync(n: number): Promise<ComputeResult>
  subscribeCompute(cb: (state: ComputeResult) => void): void

  // App
  shutdown(): void
  register(name: string, fn: (...args: unknown[]) => unknown): void
  call(name: string, ...args: unknown[]): unknown
  appendLog(level: string, msg: string): void
}

// Globals from wasm_exec.js / glue.js (internal only, not exported)
declare global {
  class Go {
    importObject: WebAssembly.Imports
    run(instance: WebAssembly.Instance): Promise<void>
  }
  interface Window {
    wasmflux: Bridge
  }
}

// ---------------------------------------------------------------------------
// Event emitter (replaces global subscribe pattern)
// ---------------------------------------------------------------------------

type Listener<T> = (value: T) => void
type Unsubscribe = () => void

class Emitter<Events extends { [key: string]: unknown }> {
  private listeners = new Map<keyof Events, Set<Listener<never>>>()

  on<K extends keyof Events>(event: K, fn: Listener<Events[K]>): Unsubscribe {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }
    const set = this.listeners.get(event)!
    set.add(fn as Listener<never>)
    return () => set.delete(fn as Listener<never>)
  }

  emit<K extends keyof Events>(event: K, value: Events[K]) {
    this.listeners.get(event)?.forEach((fn) => (fn as Listener<Events[K]>)(value))
  }
}

// ---------------------------------------------------------------------------
// WasmFlux SDK
// ---------------------------------------------------------------------------

interface WasmFluxEvents {
  [key: string]: unknown
  counter: CounterState
  signal: SignalState
  compute: ComputeResult
}

export class WasmFlux {
  private bridge: Bridge
  private emitter = new Emitter<WasmFluxEvents>()

  private constructor(bridge: Bridge) {
    this.bridge = bridge

    // Wire Go push callbacks → internal emitter (one-time setup).
    this.bridge.subscribeCounter((s) => this.emitter.emit("counter", s))
    this.bridge.subscribeSignal((s) => this.emitter.emit("signal", s))
    this.bridge.subscribeCompute((s) => this.emitter.emit("compute", s))
  }

  /**
   * Initialize WASM and return a ready-to-use SDK instance.
   * This is the only way to create a WasmFlux instance.
   */
  static async init(wasmUrl: string): Promise<WasmFlux> {
    const go = new Go()
    const result = await WebAssembly.instantiateStreaming(
      fetch(wasmUrl),
      go.importObject,
    )
    // Fire-and-forget: resolves when Go main() exits.
    go.run(result.instance)
    // Yield one tick to let Go's synchronous init complete.
    await new Promise((r) => setTimeout(r, 0))

    return new WasmFlux(window.wasmflux)
  }

  // --- Counter ---

  increment() {
    this.bridge.increment()
  }

  decrement() {
    this.bridge.decrement()
  }

  add(n: number) {
    this.bridge.add(n)
  }

  getCounterState(): CounterState {
    return this.bridge.getState()
  }

  // --- Signal ---

  pushSignal(value: number) {
    this.bridge.pushSignal(value)
  }

  pushSignalBatch(values: Float64Array) {
    this.bridge.pushSignalBatch(values)
  }

  analyzeAsync(): Promise<SignalState> {
    return this.bridge.analyzeAsync()
  }

  getSignalState(): SignalState {
    return this.bridge.getSignalState()
  }

  // --- Compute ---

  computeAsync(n: number): Promise<ComputeResult> {
    return this.bridge.computeAsync(n)
  }

  // --- Subscriptions ---

  on<K extends keyof WasmFluxEvents>(
    event: K,
    fn: Listener<WasmFluxEvents[K]>,
  ): Unsubscribe {
    return this.emitter.on(event, fn)
  }

  // --- Lifecycle ---

  shutdown() {
    this.bridge.shutdown()
  }
}
