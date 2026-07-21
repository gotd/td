package pool

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/rpc"
)

func TestErrRetryableOnNewConn_AckedCloseNotRetryable(t *testing.T) {
	// An acknowledged request may already have been processed by the
	// server: errRetryableOnNewConn must NOT classify it as safe to retry
	// on a new connection, or a transparent resend could duplicate the RPC.
	require.False(t, errRetryableOnNewConn(rpc.ErrEngineClosedAfterAck))
}
