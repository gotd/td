package testutil

import (
	"testing"
)

// SetEnv sets environment variable and sets delete callback
// to unset variable after test.
func SetEnv(t *testing.T, k, v string) {
	// Set envs.
	t.Setenv(k, v)
}
