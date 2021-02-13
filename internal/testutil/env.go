package testutil

import (
	"os"
	"testing"
)

// SetEnv sets environment variable and sets delete callback
// to unset variable after test.
func SetEnv(t *testing.T, k, v string) {
	// Set envs.
	if err := os.Setenv(k, v); err != nil {
		t.Fatalf("Setting env %q failed: %s", k, err)
	}
	// Set cleanup callback.
	t.Cleanup(func() {
		_ = os.Unsetenv(k)
	})
}
