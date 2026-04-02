import { useSignalState } from "../hooks/useWasm"

export function SignalMonitor() {
  const { state, pushBatch, analyzeAsync } = useSignalState()
  const recent = state.recent || []
  const maxAbs = Math.max(1, ...recent.map(Math.abs))

  return (
    <div style={styles.card}>
      <h2 style={styles.title}>Signal Processing</h2>

      <div style={styles.stats}>
        <Stat label="Mean" value={state.mean} />
        <Stat label="StdDev" value={state.stdDev} />
        <Stat label="Min" value={state.min} />
        <Stat label="Max" value={state.max} />
        <Stat label="Samples" value={state.count} />
      </div>

      <div style={styles.graph}>
        {recent.map((val, i) => {
          const pct = (val / maxAbs) * 50
          return (
            <div key={i} style={styles.barCol}>
              <div
                style={{
                  ...styles.barFill,
                  height: `${Math.abs(pct)}%`,
                  marginTop: val >= 0 ? `${50 - pct}%` : "50%",
                  background: val >= 0 ? "#89b4fa" : "#f38ba8",
                }}
              />
            </div>
          )
        })}
        <div style={styles.zeroLine} />
      </div>

      <div style={styles.actions}>
        <button
          style={styles.btn}
          onClick={() => {
            const burst = new Float64Array(64)
            for (let i = 0; i < 64; i++) {
              burst[i] = Math.sin(i * 0.3) * (1 + Math.random() * 0.5)
            }
            pushBatch(burst)
          }}
        >
          Push 64 Samples
        </button>
        <button
          style={styles.btn}
          onClick={async () => {
            const result = await analyzeAsync()
            console.log("Async analysis:", result)
          }}
        >
          Analyze (Async)
        </button>
      </div>
    </div>
  )
}

function Stat({ label, value }: { label: string; value: number }) {
  return (
    <div style={styles.stat}>
      <div style={styles.statLabel}>{label}</div>
      <div style={styles.statValue}>{value.toFixed(3)}</div>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  card: { background: "#1e1e2e", borderRadius: 8, padding: 20, marginTop: 16 },
  title: { margin: "0 0 16px", color: "#cdd6f4", fontSize: 18 },
  stats: { display: "flex", gap: 12, flexWrap: "wrap", marginBottom: 16 },
  stat: { background: "#11111b", borderRadius: 6, padding: "8px 14px", minWidth: 80 },
  statLabel: { fontSize: 11, color: "#6c7086", textTransform: "uppercase" as const, marginBottom: 2 },
  statValue: { fontSize: 18, color: "#cdd6f4", fontFamily: "monospace" },
  graph: { position: "relative" as const, display: "flex", alignItems: "stretch", gap: 1, height: 100, background: "#11111b", borderRadius: 6, padding: 4, marginBottom: 16 },
  barCol: { flex: 1, position: "relative" as const },
  barFill: { position: "absolute" as const, left: 0, right: 0, borderRadius: 1 },
  zeroLine: { position: "absolute" as const, left: 4, right: 4, top: "50%", height: 1, background: "#45475a" },
  actions: { display: "flex", gap: 8 },
  btn: { padding: "8px 16px", fontSize: 14, background: "#313244", color: "#cdd6f4", border: "1px solid #45475a", borderRadius: 6, cursor: "pointer" },
}
