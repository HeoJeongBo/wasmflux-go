import { useState, useEffect } from "react"
import { useWasmFlux } from "../hooks/useWasm"
import type { CounterState, SignalState, ComputeResult } from "../lib/wasmflux"

interface AppState {
  counter: CounterState
  signal: Omit<SignalState, "recent"> & { samples: number }
  lastCompute: ComputeResult | null
}

export function StateInspector() {
  const sdk = useWasmFlux()

  const [state, setState] = useState<AppState>({
    counter: { count: 0, fps: 0 },
    signal: { mean: 0, stdDev: 0, min: 0, max: 0, count: 0, samples: 0 },
    lastCompute: null,
  })

  useEffect(() => {
    const unsubs = [
      sdk.on("counter", (c) =>
        setState((prev) => ({ ...prev, counter: c })),
      ),
      sdk.on("signal", (s) =>
        setState((prev) => ({
          ...prev,
          signal: {
            mean: s.mean,
            stdDev: s.stdDev,
            min: s.min,
            max: s.max,
            count: s.count,
            samples: s.recent.length,
          },
        })),
      ),
      sdk.on("compute", (c) =>
        setState((prev) => ({ ...prev, lastCompute: c })),
      ),
    ]
    return () => unsubs.forEach((fn) => fn())
  }, [sdk])

  return (
    <div style={styles.card}>
      <h2 style={styles.title}>State Inspector</h2>
      <pre style={styles.json}>{JSON.stringify(state, null, 2)}</pre>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  card: { background: "#1e1e2e", borderRadius: 8, padding: 20, marginTop: 16 },
  title: { margin: "0 0 12px", color: "#cdd6f4", fontSize: 18 },
  json: { background: "#11111b", padding: 12, borderRadius: 6, color: "#89b4fa", fontSize: 13, margin: 0, overflow: "auto", fontFamily: "monospace" },
}
