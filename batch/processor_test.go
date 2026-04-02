package batch

import (
	"slices"
	"testing"
)

func TestProcessor_AutoFlush(t *testing.T) {
	var flushed [][]int

	p := NewProcessor(3, func(items []int) {
		flushed = append(flushed, items)
	})

	p.Push(1)
	p.Push(2)
	p.Push(3) // triggers flush

	if len(flushed) != 1 {
		t.Fatalf("flush count: %d, want 1", len(flushed))
	}
	if !slices.Equal(flushed[0], []int{1, 2, 3}) {
		t.Errorf("flushed: %v, want [1 2 3]", flushed[0])
	}
}

func TestProcessor_ManualFlush(t *testing.T) {
	var flushed [][]int

	p := NewProcessor[int](10, func(items []int) {
		flushed = append(flushed, items)
	})

	p.Push(1)
	p.Push(2)
	p.Flush()

	if len(flushed) != 1 {
		t.Fatalf("flush count: %d, want 1", len(flushed))
	}
	if !slices.Equal(flushed[0], []int{1, 2}) {
		t.Errorf("flushed: %v, want [1 2]", flushed[0])
	}
}

func TestProcessor_EmptyFlush(t *testing.T) {
	flushCount := 0
	p := NewProcessor[int](10, func(_ []int) {
		flushCount++
	})

	p.Flush() // should be no-op

	if flushCount != 0 {
		t.Errorf("flush count: %d, want 0", flushCount)
	}
}

func TestProcessor_PushBatch(t *testing.T) {
	var flushed [][]int

	p := NewProcessor(3, func(items []int) {
		flushed = append(flushed, items)
	})

	p.PushBatch([]int{1, 2, 3, 4, 5})

	if len(flushed) != 1 {
		t.Fatalf("flush count: %d, want 1", len(flushed))
	}

	// Remaining 2 items should still be in the buffer.
	if p.Len() != 2 {
		t.Errorf("remaining: %d, want 2", p.Len())
	}
}

func TestProcessor_Reset(t *testing.T) {
	p := NewProcessor[int](10, func(_ []int) {})

	p.Push(1)
	p.Push(2)
	p.Reset()

	if p.Len() != 0 {
		t.Errorf("len after reset: %d, want 0", p.Len())
	}
}

func BenchmarkProcessor_Push(b *testing.B) {
	p := NewProcessor(64, func(_ []float64) {})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Push(float64(i))
	}
}
