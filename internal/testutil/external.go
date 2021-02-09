package testutil

import (
	"os"
	"strconv"
	"testing"
)

// SkipExternal skips current test if GOTD_TEST_EXTERNAL is not 1.
func SkipExternal(tb testing.TB) {
	const env = "GOTD_TEST_EXTERNAL"

	tb.Helper()

	// TODO(ar): We can check test function for TestExternalE2E* prefix here.

	if ok, _ := strconv.ParseBool(os.Getenv(env)); !ok {
		tb.Skipf("Skipped. Set %s=1 to enable external e2e test.", env)
	}
}
