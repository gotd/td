package testutil

import "testing"

const defaultAllocRuns = 10

// MaxAlloc checks that f does not allocate more than n.
func MaxAlloc(t *testing.T, n int, f func()) {
	t.Helper()
	if Race {
		t.Skip("Skipped (race detector conflicts with allocation tests)")
	}
	avg := testing.AllocsPerRun(defaultAllocRuns, f)
	if avg > float64(n) {
		t.Errorf("Allocated %f bytes per run, expected less than %d", avg, n)
	}
}

// ZeroAlloc checks that f does not allocate.
func ZeroAlloc(t *testing.T, f func()) {
	t.Helper()
	MaxAlloc(t, 0, f)
}
