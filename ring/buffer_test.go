package ring

import (
	"slices"
	"testing"
)

func TestBuffer_Write(t *testing.T) {
	b := NewBuffer[int](3)

	b.Write(1)
	b.Write(2)
	b.Write(3)

	got := b.Values()
	want := []int{1, 2, 3}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBuffer_Overwrite(t *testing.T) {
	b := NewBuffer[int](3)

	b.Write(1)
	b.Write(2)
	b.Write(3)
	b.Write(4) // overwrites 1
	b.Write(5) // overwrites 2

	got := b.Values()
	want := []int{3, 4, 5}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBuffer_WriteBatch(t *testing.T) {
	b := NewBuffer[int](4)

	b.WriteBatch([]int{1, 2, 3})
	got := b.Values()
	want := []int{1, 2, 3}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	b.WriteBatch([]int{4, 5})
	got = b.Values()
	want = []int{2, 3, 4, 5}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBuffer_WriteBatchOverflow(t *testing.T) {
	b := NewBuffer[int](3)

	b.WriteBatch([]int{1, 2, 3, 4, 5})
	got := b.Values()
	want := []int{3, 4, 5}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBuffer_Drain(t *testing.T) {
	b := NewBuffer[int](3)
	b.Write(1)
	b.Write(2)

	dst := make([]int, 3)
	n := b.Drain(dst)
	if n != 2 {
		t.Errorf("drain returned %d, want 2", n)
	}
	got := dst[:n]
	want := []int{1, 2}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if b.Len() != 0 {
		t.Errorf("len after drain: %d, want 0", b.Len())
	}
}

func TestBuffer_Iter(t *testing.T) {
	b := NewBuffer[int](3)
	b.Write(1)
	b.Write(2)
	b.Write(3)
	b.Write(4)

	var got []int
	for v := range b.Iter() {
		got = append(got, v)
	}
	want := []int{2, 3, 4}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBuffer_BackwardIter(t *testing.T) {
	b := NewBuffer[int](3)
	b.Write(1)
	b.Write(2)
	b.Write(3)
	b.Write(4)

	var got []int
	for v := range b.BackwardIter() {
		got = append(got, v)
	}
	want := []int{4, 3, 2}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBuffer_Clear(t *testing.T) {
	b := NewBuffer[int](3)
	b.Write(1)
	b.Write(2)
	b.Clear()

	if b.Len() != 0 {
		t.Errorf("len after clear: %d, want 0", b.Len())
	}
}

func TestBuffer_Read(t *testing.T) {
	b := NewBuffer[int](4)
	b.Write(10)
	b.Write(20)
	b.Write(30)

	dst := make([]int, 4)
	n := b.Read(dst)
	if n != 3 {
		t.Errorf("Read returned %d, want 3", n)
	}
	want := []int{10, 20, 30}
	if !slices.Equal(dst[:n], want) {
		t.Errorf("got %v, want %v", dst[:n], want)
	}

	// Read again — should return same data (non-destructive).
	n2 := b.Read(dst)
	if n2 != 3 {
		t.Errorf("second Read returned %d, want 3", n2)
	}
}

func TestBuffer_LenCap(t *testing.T) {
	b := NewBuffer[int](5)
	if b.Len() != 0 {
		t.Errorf("Len: %d, want 0", b.Len())
	}
	if b.Cap() != 5 {
		t.Errorf("Cap: %d, want 5", b.Cap())
	}

	b.Write(1)
	b.Write(2)
	if b.Len() != 2 {
		t.Errorf("Len: %d, want 2", b.Len())
	}

	// Overfill.
	for i := 0; i < 10; i++ {
		b.Write(i)
	}
	if b.Len() != 5 {
		t.Errorf("Len should cap at capacity: %d", b.Len())
	}
}

func TestBuffer_Capacity1(t *testing.T) {
	b := NewBuffer[string](1)
	b.Write("a")
	b.Write("b")

	got := b.Values()
	if len(got) != 1 || got[0] != "b" {
		t.Errorf("got %v, want [b]", got)
	}
}

func TestBuffer_EmptyOperations(t *testing.T) {
	b := NewBuffer[int](3)

	// Read from empty.
	dst := make([]int, 3)
	n := b.Read(dst)
	if n != 0 {
		t.Errorf("Read empty: %d, want 0", n)
	}

	// Drain from empty.
	n = b.Drain(dst)
	if n != 0 {
		t.Errorf("Drain empty: %d, want 0", n)
	}

	// Values from empty.
	if len(b.Values()) != 0 {
		t.Error("Values should be empty")
	}

	// Iter from empty.
	count := 0
	for range b.Iter() {
		count++
	}
	if count != 0 {
		t.Error("Iter should be empty")
	}
}

func TestBuffer_NewBuffer_ZeroCapacity(t *testing.T) {
	b := NewBuffer[int](0)
	if b.Cap() != 1 {
		t.Errorf("Cap: %d, want 1 (clamped)", b.Cap())
	}
	b.Write(42)
	if b.Values()[0] != 42 {
		t.Error("should work with clamped capacity")
	}
}

func TestBuffer_NewBuffer_NegativeCapacity(t *testing.T) {
	b := NewBuffer[int](-5)
	if b.Cap() != 1 {
		t.Errorf("Cap: %d, want 1 (clamped)", b.Cap())
	}
}

func TestBuffer_View_NotFull(t *testing.T) {
	b := NewBuffer[int](5)
	b.Write(1)
	b.Write(2)

	var got []int
	b.View(func(v int) bool {
		got = append(got, v)
		return true
	})
	want := []int{1, 2}
	if !slices.Equal(got, want) {
		t.Errorf("View not full: %v, want %v", got, want)
	}
}

func TestBuffer_View_Full(t *testing.T) {
	b := NewBuffer[int](3)
	b.Write(1)
	b.Write(2)
	b.Write(3)
	b.Write(4) // overwrites 1

	var got []int
	b.View(func(v int) bool {
		got = append(got, v)
		return true
	})
	want := []int{2, 3, 4}
	if !slices.Equal(got, want) {
		t.Errorf("View full: %v, want %v", got, want)
	}
}

func TestBuffer_View_EarlyStop(t *testing.T) {
	b := NewBuffer[int](5)
	b.Write(1)
	b.Write(2)
	b.Write(3)

	var got []int
	b.View(func(v int) bool {
		got = append(got, v)
		return len(got) < 2 // stop after 2
	})
	if len(got) != 2 {
		t.Errorf("got %d items, want 2", len(got))
	}
}

func TestBuffer_View_EarlyStop_Full(t *testing.T) {
	b := NewBuffer[int](3)
	b.Write(1)
	b.Write(2)
	b.Write(3)
	b.Write(4)

	var got []int
	b.View(func(v int) bool {
		got = append(got, v)
		return len(got) < 1
	})
	if len(got) != 1 {
		t.Errorf("got %d items, want 1", len(got))
	}
}

func TestBuffer_BackwardView_NotFull(t *testing.T) {
	b := NewBuffer[int](5)
	b.Write(10)
	b.Write(20)

	var got []int
	b.BackwardView(func(v int) bool {
		got = append(got, v)
		return true
	})
	want := []int{20, 10}
	if !slices.Equal(got, want) {
		t.Errorf("BackwardView: %v, want %v", got, want)
	}
}

func TestBuffer_BackwardView_Full(t *testing.T) {
	b := NewBuffer[int](3)
	b.Write(1)
	b.Write(2)
	b.Write(3)
	b.Write(4) // overwrites 1

	var got []int
	b.BackwardView(func(v int) bool {
		got = append(got, v)
		return true
	})
	want := []int{4, 3, 2}
	if !slices.Equal(got, want) {
		t.Errorf("BackwardView full: %v, want %v", got, want)
	}
}

func TestBuffer_BackwardView_EarlyStop(t *testing.T) {
	b := NewBuffer[int](3)
	b.Write(1)
	b.Write(2)
	b.Write(3)
	b.Write(4)

	var got []int
	b.BackwardView(func(v int) bool {
		got = append(got, v)
		return len(got) < 1
	})
	if len(got) != 1 {
		t.Errorf("got %d items, want 1", len(got))
	}
}

func TestBuffer_BackwardView_EarlyStop_NotFull(t *testing.T) {
	b := NewBuffer[int](5)
	b.Write(10)
	b.Write(20)
	b.Write(30)

	var got []int
	b.BackwardView(func(v int) bool {
		got = append(got, v)
		return len(got) < 2
	})
	want := []int{30, 20}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func BenchmarkBuffer_Write(b *testing.B) {
	buf := NewBuffer[float64](1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Write(float64(i))
	}
}

func BenchmarkBuffer_WriteBatch(b *testing.B) {
	buf := NewBuffer[float64](1024)
	batch := make([]float64, 60) // 60Hz frame
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.WriteBatch(batch)
	}
}
