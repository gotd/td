package pool

import (
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/rpc"
	"github.com/gotd/td/transport"
)

func TestErrRetryableOnNewConn_AckedCloseNotRetryable(t *testing.T) {
	// An acknowledged request may already have been processed by the
	// server: errRetryableOnNewConn must NOT classify it as safe to retry
	// on a new connection, or a transparent resend could duplicate the RPC.
	// This holds regardless of the write-failure opt-in.
	require.False(t, errRetryableOnNewConn(rpc.ErrEngineClosedAfterAck, false))
	require.False(t, errRetryableOnNewConn(rpc.ErrEngineClosedAfterAck, true))
}

func TestErrRetryableOnNewConn_WriteFailedIsOptIn(t *testing.T) {
	err := errors.Wrap(transport.ErrWriteFailed, "write")

	require.False(t, errRetryableOnNewConn(err, false),
		"a failed transport send must surface to the caller unless retrying was requested")
	require.True(t, errRetryableOnNewConn(err, true))
}
