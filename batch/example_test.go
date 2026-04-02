package batch_test

import (
	"fmt"

	"github.com/heojeongbo/wasmflux-go/batch"
)

func ExampleProcessor() {
	p := batch.NewProcessor(3, func(items []int) {
		fmt.Println("flush:", items)
	})

	p.Push(1)
	p.Push(2)
	p.Push(3) // auto-flush at capacity
	p.Push(4)
	p.Push(5)
	p.Flush() // manual flush of remaining
	// Output:
	// flush: [1 2 3]
	// flush: [4 5]
}
