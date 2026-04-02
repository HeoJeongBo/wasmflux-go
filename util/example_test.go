package util_test

import (
	"errors"
	"fmt"

	"github.com/heojeongbo/wasmflux-go/util"
)

func ExampleRetry() {
	attempts := 0
	err := util.Retry(func() error {
		attempts++
		if attempts < 3 {
			return errors.New("not ready")
		}
		return nil
	}, util.WithMaxAttempts(5), util.WithBaseDelay(0))

	fmt.Println("succeeded after", attempts, "attempts, err:", err)
	// Output: succeeded after 3 attempts, err: <nil>
}

func ExampleRateLimiter() {
	rl := util.NewRateLimiter(1000, 5) // 1000/sec, burst 5

	allowed := 0
	for i := 0; i < 10; i++ {
		if rl.Allow() {
			allowed++
		}
	}
	fmt.Println("allowed:", allowed)
	// Output: allowed: 5
}
