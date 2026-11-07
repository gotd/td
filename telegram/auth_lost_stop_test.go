package telegram

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/log"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// TestClientRunAuthLostStopAfterCallback checks that Run returns
// *AuthLostError when OnAuthLost asks to stop, even if the Run callback
// returned before the report finished.
func TestClientRunAuthLostStopAfterCallback(t *testing.T) {
	for _, tt := range []struct {
		name string
		f    func(err error) error
	}{
		{"CallbackNil", func(error) error { return nil }},
		{"CallbackError", func(err error) error { return err }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var fail atomic.Pointer[tgerr.Error]
			reason := errors.New("stop")
			c := NewClient(1, "hash", Options{
				NoUpdates: true,
				Logger:    log.Nop,
				OnAuthLost: func(ctx context.Context, err error) error {
					<-ctx.Done() // Run callback has returned.
					return reason
				},
			})
			conn := newTestClient(authLostHandler(&fail)).conn
			c.conn = authRunConn{conn: conn, client: c}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			err := c.Run(ctx, func(ctx context.Context) error {
				if _, err := c.API().AuthImportBotAuthorization(ctx, &tg.AuthImportBotAuthorizationRequest{}); err != nil {
					return err
				}
				fail.Store(tgerr.New(401, "AUTH_KEY_UNREGISTERED"))
				_, err := c.API().HelpGetConfig(ctx)
				return tt.f(err)
			})
			var lost *AuthLostError
			require.ErrorAs(t, err, &lost, "got %v", err)
			require.ErrorIs(t, lost.Reason, reason)
		})
	}
}

// TestClientAuthLostAfterRun checks that no OnAuthLost call starts once Run
// is stopping.
func TestClientAuthLostAfterRun(t *testing.T) {
	var called atomic.Bool
	c := newTestClient(func(id int64, body bin.Encoder) (bin.Encoder, error) {
		return nil, tgerr.New(401, "AUTH_KEY_UNREGISTERED")
	})
	c.onAuthLost = func(context.Context, error) error {
		called.Store(true)
		return nil
	}
	c.auth.authorized.Store(true)
	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	cancel() // Run is stopping.

	_, _ = c.API().HelpGetConfig(context.Background())
	require.NoError(t, c.waitAuthLost(context.Background()))
	require.False(t, called.Load())
	require.Nil(t, c.auth.done.Load(), "no report started")
}
