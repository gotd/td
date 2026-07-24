package rpc

import (
	"fmt"

	"github.com/go-faster/errors"
)

// RetryLimitReachedErr means that server does not acknowledge request
// after multiple retries.
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

// ErrEngineClosedAfterAck means that the request was acknowledged by the
// server, but the engine was forcibly closed before a result arrived. The
// server may have already processed the request, so — unlike
// ErrEngineClosed — this error is deliberately not classified as retryable:
// a transparent resend on a new connection could duplicate the RPC (e.g. a
// second messages.sendMessage).
var ErrEngineClosedAfterAck = errors.New("engine forcibly closed after request was acknowledged")

// ackedCloseError reports that an acknowledged request lost its engine
// before a result arrived.
//
// It deliberately unwraps only to the plain cancellation cause (so callers
// checking errors.Is(err, context.Canceled) keep working, and
// errors.Is(err, ErrEngineClosed) stays false, preserving non-retryability),
// while additionally matching ErrEngineClosedAfterAck via Is for
// diagnosability.
//
// A hand-rolled type is used instead of errors.Join(cause, ErrEngineClosedAfterAck)
// to keep the rendered message on a single line; errors.Join embeds a
// newline between joined errors, which is an existing annoyance in
// transport/connection.go that this avoids repeating.
type ackedCloseError struct {
	cause error
}

func (e *ackedCloseError) Error() string {
	// Name the acknowledgement explicitly: an operator reading this line for a
	// stuck request must not mistake it for caller-initiated cancellation, since
	// the server may have executed the request already.
	return "engine forcibly closed after request was acknowledged: " + e.cause.Error()
}

func (e *ackedCloseError) Unwrap() error {
	return e.cause
}

func (e *ackedCloseError) Is(target error) bool {
	return target == ErrEngineClosedAfterAck
}

// ErrRetrySendFailed means that a request was sent, went unacknowledged for
// RetryInterval, and the retry send then failed. The original send already
// reached the wire, so the server may hold the request: unlike a failure of
// the very first send, this is deliberately not classified as retryable, since
// a transparent resend uses a fresh msg_id and would bypass the server's
// deduplication.
var ErrRetrySendFailed = errors.New("retry send failed for unacknowledged request")

// retrySendError reports that the retry send of a still-unacknowledged request
// failed.
//
// It deliberately does NOT implement Unwrap: the underlying cause routinely
// carries transport.ErrWriteFailed, which callers (telegram/invoke.go,
// pool/pool_conn.go) treat as "safe to resend on a new connection". That is
// true for the first send of a request and false here, so the chain is cut and
// the cause survives only in the rendered message. Matching ErrRetrySendFailed
// via Is keeps the case identifiable.
//
// Hand-rolled rather than errors.Join for the same reason as ackedCloseError
// above: errors.Join embeds a newline between joined errors.
type retrySendError struct {
	cause error
}

func (e *retrySendError) Error() string {
	return ErrRetrySendFailed.Error() + ": " + e.cause.Error()
}

func (e *retrySendError) Is(target error) bool {
	return target == ErrRetrySendFailed
}
