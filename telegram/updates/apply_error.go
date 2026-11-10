package updates

import "github.com/go-faster/errors"

// applyError marks a Handler or StateStorage error of the apply path with
// Config.StopOnApplyError, distinguishing failed local effects from
// recoverable RPC errors. It stops the manager: Run returns it.
type applyError struct{ error }

func (e *applyError) Unwrap() error { return e.error }

func isApplyError(err error) bool {
	var apply *applyError
	return errors.As(err, &apply)
}
