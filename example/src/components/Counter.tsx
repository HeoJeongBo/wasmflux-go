import { useState } from "react"
import { useCounterState } from "../hooks/useWasm"

export function Counter() {
  const { state, increment, decrement, add } = useCounterState()
  const [addValue, setAddValue] = useState(10)

  return (
    <div style={styles.card}>
      <h2 style={styles.title}>Counter</h2>
      <div style={styles.count}>{state.count}</div>
      <div style={styles.row}>
        <button style={styles.btn} onClick={decrement}>-</button>
        <button style={styles.btn} onClick={increment}>+</button>
      </div>
      <div style={styles.row}>
        <input
          type="number"
          value={addValue}
          onChange={(e) => setAddValue(Number(e.target.value))}
          style={styles.input}
        />
        <button style={styles.btn} onClick={() => add(addValue)}>Add</button>
      </div>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  card: { background: "#1e1e2e", borderRadius: 8, padding: 20 },
  title: { margin: "0 0 12px", color: "#cdd6f4", fontSize: 18 },
  count: { fontSize: 48, fontWeight: "bold", color: "#89b4fa", textAlign: "center", margin: "12px 0" },
  row: { display: "flex", gap: 8, justifyContent: "center", marginTop: 8 },
  btn: { padding: "8px 20px", fontSize: 16, background: "#313244", color: "#cdd6f4", border: "1px solid #45475a", borderRadius: 6, cursor: "pointer" },
  input: { width: 80, padding: "8px 12px", fontSize: 16, background: "#313244", color: "#cdd6f4", border: "1px solid #45475a", borderRadius: 6, textAlign: "center" as const },
}
