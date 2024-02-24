package testutil

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// SkipExternal skips current test if GOTD_TEST_EXTERNAL is not 1.
//
// Caller should be high-level test function with TestExternalE2E prefix,
// like TestExternalE2EConnect.
//
// Run all tests with following command in module root:
//
//	GOTD_TEST_EXTERNAL=1 go test -v -run ^TestExternalE2E ./...
func SkipExternal(tb testing.TB) {
	const env = "GOTD_TEST_EXTERNAL"

	tb.Helper()

	{
		// Checking caller prefix.
		const expectedPrefix = "TestExternalE2E"
		pc, _, _, _ := runtime.Caller(1)
		details := runtime.FuncForPC(pc)
		name := details.Name()[strings.LastIndex(details.Name(), ".")+1:]
		if !strings.HasPrefix(name, expectedPrefix) {
			tb.Fatalf("Test function %s should have prefix %s.", name, expectedPrefix)
		}
	}

	if ok, _ := strconv.ParseBool(os.Getenv(env)); !ok {
		tb.Skipf("Skipped. Set %s=1 to enable external e2e test.", env)
	}
}
