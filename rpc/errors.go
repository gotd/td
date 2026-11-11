package rpc

import (
	"fmt"

	"github.com/go-faster/errors"
)

// RetryLimitReachedErr means that server does not acknowledge request
// after multiple retries.
//
// Engine.Do no longer returns it: after the last resend it keeps waiting for
// the result until the context is done or the engine is closed. The type is
// kept for compatibility.
type RetryLimitReachedErr struct {
	Retries int
}

func (r *RetryLimitReachedErr) Error() string {
	return fmt.Sprintf("retry limit reached after %d attempts", r.Retries)
}

// Is reports whether err is RetryLimitReachedErr.
func (r *RetryLimitReachedErr) Is(err error) bool {
	_, ok := err.(*RetryLimitReachedErr)
	return ok
}

// ErrEngineClosed means that engine was closed.
var ErrEngineClosed = errors.New("engine was closed")
