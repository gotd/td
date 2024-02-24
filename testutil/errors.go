package testutil

import (
	"github.com/go-faster/errors"
)

var errTestError = errors.New("test error")

// TestError returns error for testing error pass.
func TestError() error {
	return errTestError
}
