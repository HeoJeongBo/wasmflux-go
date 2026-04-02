package ring_test

import (
	"fmt"

	"github.com/heojeongbo/wasmflux-go/ring"
)

func ExampleBuffer() {
	buf := ring.NewBuffer[int](3)

	buf.Write(1)
	buf.Write(2)
	buf.Write(3)
	buf.Write(4) // overwrites 1

	for v := range buf.Iter() {
		fmt.Print(v, " ")
	}
	fmt.Println()
	// Output: 2 3 4
}

func ExampleBuffer_WriteBatch() {
	buf := ring.NewBuffer[string](4)
	buf.WriteBatch([]string{"a", "b", "c", "d", "e"})

	fmt.Println(buf.Values())
	// Output: [b c d e]
}

func ExampleBuffer_Drain() {
	buf := ring.NewBuffer[int](3)
	buf.Write(10)
	buf.Write(20)

	dst := make([]int, 3)
	n := buf.Drain(dst)
	fmt.Println(dst[:n], "len after drain:", buf.Len())
	// Output: [10 20] len after drain: 0
}
