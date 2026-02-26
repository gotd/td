package telegram

import (
	"context"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/exchange"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/telegram/internal/manager"
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

type pfsDropConn struct{}

func (pfsDropConn) Run(context.Context) error {
	return mtproto.ErrPFSDropKeysRequired
}

func (pfsDropConn) Invoke(context.Context, bin.Encoder, bin.Decoder) error {
	return nil
}

func (pfsDropConn) Ping(context.Context) error {
	return nil
}

func (pfsDropConn) Ready() <-chan struct{} {
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

func TestClient_reconnectUntilClosedPFSDropResetsStoredKey(t *testing.T) {
	key := crypto.Key{1}.WithID()
	dcID := 2
	client := Client{
		newConnBackoff: func() backoff.BackOff {
			return backoff.NewConstantBackOff(time.Nanosecond)
		},
		log: zap.NewNop(),
	}
	client.init()
	client.session = pool.NewSyncSession(pool.Session{DC: dcID})
	client.create = func(
		mtproto.Dialer,
		manager.ConnMode,
		int,
		mtproto.Options,
		manager.ConnOptions,
	) pool.Conn {
		return pfsDropConn{}
	}
	client.session.Store(pool.Session{
		DC:      dcID,
		AuthKey: key,
		Salt:    42,
	})
	client.sessions[dcID] = pool.NewSyncSession(pool.Session{
		DC:      dcID,
		AuthKey: key,
		Salt:    42,
	})
	client.conn = pfsDropConn{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	require.Error(t, client.reconnectUntilClosed(ctx))

	// Both primary and per-DC cached sessions should be wiped for clean restart.
	primary := client.session.Load()
	require.True(t, primary.AuthKey.Zero())
	require.Zero(t, primary.Salt)

	stored := client.sessions[dcID].Load()
	require.True(t, stored.AuthKey.Zero())
	require.Zero(t, stored.Salt)
}
