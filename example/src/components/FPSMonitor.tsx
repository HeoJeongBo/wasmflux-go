import { useCounterState } from "../hooks/useWasm"

export function FPSMonitor() {
  const { state } = useCounterState()
  const fps = state.fps

  return (
    <div style={styles.card}>
      <h2 style={styles.title}>
        FPS:{" "}
        <span style={{ color: fps >= 55 ? "#a6e3a1" : fps >= 30 ? "#f9e2af" : "#f38ba8" }}>
          {fps.toFixed(1)}
        </span>
      </h2>
      <div style={styles.bar}>
        <div
          style={{
            ...styles.fill,
            width: `${Math.min((fps / 120) * 100, 100)}%`,
            background: fps >= 55 ? "#a6e3a1" : fps >= 30 ? "#f9e2af" : "#f38ba8",
          }}
        />
      </div>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  card: { background: "#1e1e2e", borderRadius: 8, padding: 20 },
  title: { margin: "0 0 12px", color: "#cdd6f4", fontSize: 18 },
  bar: { height: 12, background: "#11111b", borderRadius: 6, overflow: "hidden" },
  fill: { height: "100%", borderRadius: 6, transition: "width 0.3s" },
}
