package telegram

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/log"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// TestClientOnAuthLostAsync checks that the failing request does not wait
// for OnAuthLost.
func TestClientOnAuthLostAsync(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var fail atomic.Pointer[tgerr.Error]
	client := newTestClient(authLostHandler(&fail))
	release := make(chan struct{})
	called := make(chan struct{})
	client.onAuthLost = func(ctx context.Context, err error) error {
		close(called)
		<-release
		return nil
	}
	api := client.API()
	_, err := api.AuthImportBotAuthorization(ctx, &tg.AuthImportBotAuthorizationRequest{})
	require.NoError(t, err)

	fail.Store(tgerr.New(401, "AUTH_KEY_UNREGISTERED"))
	returned := make(chan error, 1)
	go func() {
		_, err := api.HelpGetConfig(ctx)
		returned <- err
	}()
	select {
	case err := <-returned:
		require.True(t, IsAuthLost(err))
	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("request waits for OnAuthLost")
	}

	<-called
	waitCtx, waitCancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer waitCancel()
	require.ErrorIs(t, client.waitAuthLost(waitCtx), context.DeadlineExceeded,
		"report is in progress")
	close(release)
	require.NoError(t, client.waitAuthLost(ctx))
}

// TestClientRunWaitsOnAuthLost checks that Run does not return while
// OnAuthLost runs.
func TestClientRunWaitsOnAuthLost(t *testing.T) {
	var (
		fail     atomic.Pointer[tgerr.Error]
		finished atomic.Bool
	)
	c := NewClient(1, "hash", Options{
		NoUpdates: true,
		Logger:    log.Nop,
		OnAuthLost: func(ctx context.Context, err error) error {
			<-ctx.Done() // Run is stopping.
			time.Sleep(50 * time.Millisecond)
			finished.Store(true)
			return nil
		},
	})
	conn := newTestClient(authLostHandler(&fail)).conn
	c.conn = authRunConn{conn: conn, client: c}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, c.Run(ctx, func(ctx context.Context) error {
		if _, err := c.API().AuthImportBotAuthorization(ctx, &tg.AuthImportBotAuthorizationRequest{}); err != nil {
			return err
		}
		fail.Store(tgerr.New(401, "AUTH_KEY_UNREGISTERED"))
		_, _ = c.API().HelpGetConfig(ctx)
		return nil
	}))
	require.True(t, finished.Load(), "Run must wait for OnAuthLost")
}
