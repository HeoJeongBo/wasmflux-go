import { useState } from "react"
import { useComputeState } from "../hooks/useWasm"

export function ComputePanel() {
  const { result, computing, compute } = useComputeState()
  const [input, setInput] = useState(40)

  return (
    <div style={styles.card}>
      <h2 style={styles.title}>Compute (Goroutine)</h2>
      <p style={styles.desc}>
        Heavy CPU work runs in a Go goroutine. JS gets a Promise back.
      </p>

      <div style={styles.row}>
        <input
          type="number"
          value={input}
          onChange={(e) => setInput(Number(e.target.value))}
          style={styles.input}
        />
        <button
          style={{
            ...styles.btn,
            opacity: computing ? 0.5 : 1,
            cursor: computing ? "wait" : "pointer",
          }}
          onClick={() => compute(input)}
          disabled={computing}
        >
          {computing ? "Computing..." : "Run"}
        </button>
      </div>

      {result && (
        <div style={styles.result}>
          <Row label="Input" value={String(result.input)} />
          <Row label="Fibonacci" value={String(result.fibonacci)} />
          <Row label="Primes up to N×100" value={String(result.primeCount)} />
          <Row label="Elapsed" value={`${result.elapsed.toFixed(2)} ms`} />
        </div>
      )}
    </div>
  )
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div style={styles.resultRow}>
      <span style={styles.resultLabel}>{label}</span>
      <span style={styles.resultValue}>{value}</span>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  card: { background: "#1e1e2e", borderRadius: 8, padding: 20, marginTop: 16 },
  title: { margin: "0 0 4px", color: "#cdd6f4", fontSize: 18 },
  desc: { margin: "0 0 16px", color: "#6c7086", fontSize: 13 },
  row: { display: "flex", gap: 8, marginBottom: 16 },
  input: { width: 100, padding: "8px 12px", fontSize: 16, background: "#313244", color: "#cdd6f4", border: "1px solid #45475a", borderRadius: 6, textAlign: "center" as const },
  btn: { padding: "8px 20px", fontSize: 14, background: "#313244", color: "#cdd6f4", border: "1px solid #45475a", borderRadius: 6 },
  result: { background: "#11111b", borderRadius: 6, padding: 12 },
  resultRow: { display: "flex", justifyContent: "space-between", padding: "4px 0", fontSize: 14 },
  resultLabel: { color: "#6c7086" },
  resultValue: { color: "#cdd6f4", fontFamily: "monospace" },
}
