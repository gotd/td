package testutil

import (
	"errors"
)

var errTestError = errors.New("test error")

// TestError returns error for testing error pass.
func TestError() error {
	return errTestError
}
