package testutil

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// RequireErr asserts that errors.Is(actual, expected) is true.
func RequireErr(t testing.TB, expected, actual error, msgAndArgs ...interface{}) {
	t.Helper()
	if !errors.Is(actual, expected) {
		require.Fail(t, fmt.Sprintf("Error chain does not match target:\n"+
			"expected: %q\n"+
			"actual  : %q", expected, actual), msgAndArgs...)
	}
}
