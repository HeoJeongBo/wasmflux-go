import { useState, useEffect, useRef, useContext, createContext, useCallback } from "react"
import {
  WasmFlux,
  type CounterState,
  type SignalState,
  type ComputeResult,
} from "../lib/wasmflux"

// ---------------------------------------------------------------------------
// Context — single SDK instance shared across components
// ---------------------------------------------------------------------------

const WasmFluxContext = createContext<WasmFlux | null>(null)

export const WasmFluxProvider = WasmFluxContext.Provider

export function useWasmFlux(): WasmFlux {
  const sdk = useContext(WasmFluxContext)
  if (!sdk) throw new Error("useWasmFlux must be used within WasmFluxProvider")
  return sdk
}

// ---------------------------------------------------------------------------
// useWasmInit — loads WASM and returns the SDK instance
// ---------------------------------------------------------------------------

type InitStatus = "loading" | "ready" | "error"

export function useWasmInit(wasmUrl: string) {
  const [status, setStatus] = useState<InitStatus>("loading")
  const [error, setError] = useState<Error | null>(null)
  const [sdk, setSdk] = useState<WasmFlux | null>(null)

  useEffect(() => {
    let cancelled = false

    WasmFlux.init(wasmUrl)
      .then((instance) => {
        if (!cancelled) {
          setSdk(instance)
          setStatus("ready")
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err : new Error(String(err)))
          setStatus("error")
        }
      })

    return () => {
      cancelled = true
    }
  }, [wasmUrl])

  return { status, error, sdk }
}

// ---------------------------------------------------------------------------
// Domain hooks — subscribe to SDK events
// ---------------------------------------------------------------------------

export function useCounterState() {
  const sdk = useWasmFlux()
  const [state, setState] = useState<CounterState>({ count: 0, fps: 0 })

  useEffect(() => {
    return sdk.on("counter", setState)
  }, [sdk])

  const increment = useCallback(() => sdk.increment(), [sdk])
  const decrement = useCallback(() => sdk.decrement(), [sdk])
  const add = useCallback((n: number) => sdk.add(n), [sdk])

  return { state, increment, decrement, add }
}

export function useSignalState() {
  const sdk = useWasmFlux()
  const [state, setState] = useState<SignalState>({
    mean: 0, stdDev: 0, min: 0, max: 0, count: 0, recent: [],
  })

  useEffect(() => {
    return sdk.on("signal", setState)
  }, [sdk])

  const pushSignal = useCallback((v: number) => sdk.pushSignal(v), [sdk])
  const pushBatch = useCallback((v: Float64Array) => sdk.pushSignalBatch(v), [sdk])
  const analyzeAsync = useCallback(() => sdk.analyzeAsync(), [sdk])

  return { state, pushSignal, pushBatch, analyzeAsync }
}

export function useComputeState() {
  const sdk = useWasmFlux()
  const [result, setResult] = useState<ComputeResult | null>(null)
  const [computing, setComputing] = useState(false)

  useEffect(() => {
    return sdk.on("compute", setResult)
  }, [sdk])

  const compute = useCallback(async (n: number) => {
    setComputing(true)
    try {
      const res = await sdk.computeAsync(n)
      setResult(res)
      return res
    } finally {
      setComputing(false)
    }
  }, [sdk])

  return { result, computing, compute }
}
