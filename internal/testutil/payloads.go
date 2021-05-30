package testutil

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
