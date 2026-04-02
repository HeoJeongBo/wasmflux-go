import { useWasmInit, WasmFluxProvider } from "./hooks/useWasm"
import { Counter } from "./components/Counter"
import { FPSMonitor } from "./components/FPSMonitor"
import { SignalMonitor } from "./components/SignalMonitor"
import { ComputePanel } from "./components/ComputePanel"
import { StateInspector } from "./components/StateInspector"

export function App() {
  const { status, error, sdk } = useWasmInit("/app.wasm")

  if (status === "loading") {
    return <div style={styles.status}>Loading WASM...</div>
  }
  if (status === "error") {
    return <div style={{ ...styles.status, color: "#f38ba8" }}>Error: {error?.message}</div>
  }

  return (
    <WasmFluxProvider value={sdk}>
      <div style={styles.app}>
        <h1 style={styles.header}>wasmflux-go</h1>
        <p style={styles.sub}>Module DI · Signal Processing · Goroutine Compute · SDK Pattern</p>
        <div style={styles.grid}>
          <Counter />
          <FPSMonitor />
        </div>
        <SignalMonitor />
        <ComputePanel />
        <StateInspector />
      </div>
    </WasmFluxProvider>
  )
}

const styles: Record<string, React.CSSProperties> = {
  app: { maxWidth: 700, margin: "0 auto", padding: 24, fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif" },
  header: { color: "#89b4fa", margin: 0, fontSize: 28 },
  sub: { color: "#6c7086", margin: "4px 0 24px", fontSize: 14 },
  grid: { display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16 },
  status: { display: "flex", alignItems: "center", justifyContent: "center", height: "100vh", fontSize: 18, color: "#cdd6f4", fontFamily: "monospace" },
}
