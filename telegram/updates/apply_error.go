package updates

import "github.com/go-faster/errors"

// applyError distinguishes failed local effects from recoverable RPC errors.
type applyError struct{ error }

func (e *applyError) Unwrap() error { return e.error }

func isApplyError(err error) bool {
	var apply *applyError
	return errors.As(err, &apply)
}
