package testutil

import (
	"fmt"
	"testing"
)

// Payloads returns a payload sizes list.
// Helpful for benchmarks.
func Payloads() []int {
	return []int{
		16,
		128,
		1024,
		8192,
		64 * 1024,
		512 * 1024,
	}
}

// RunPayloads runs given benchmark runner for every payload.
func RunPayloads(b *testing.B, runner func(size int) func(b *testing.B)) {
	for _, size := range Payloads() {
		b.Run(fmt.Sprintf("%db", size), runner(size))
	}
}
