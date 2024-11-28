package telegram

import (
	"context"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/exchange"
)

type fingerprintNotFoundConn struct{}

func (m fingerprintNotFoundConn) Run(context.Context) error {
	return exchange.ErrKeyFingerprintNotFound
}

func (m fingerprintNotFoundConn) Invoke(context.Context, bin.Encoder, bin.Decoder) error {
	return nil
}

func (m fingerprintNotFoundConn) Ping(context.Context) error {
	return nil
}

func (m fingerprintNotFoundConn) Ready() <-chan struct{} {
	return nil
}

func TestClient_reconnectUntilClosed(t *testing.T) {
	client := Client{
		newConnBackoff: func() backoff.BackOff {
			return backoff.NewConstantBackOff(time.Nanosecond)
		},
		log: zap.NewNop(),
	}
	client.init()
	client.conn = fingerprintNotFoundConn{}

	ctx := context.Background()
	require.ErrorIs(t, client.reconnectUntilClosed(ctx), exchange.ErrKeyFingerprintNotFound)
}
