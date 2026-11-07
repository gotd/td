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

// authLostHandler fails all non-auth requests with given error.
func authLostHandler(fail *atomic.Pointer[tgerr.Error]) testHandler {
	return func(id int64, body bin.Encoder) (bin.Encoder, error) {
		switch body.(type) {
		case *tg.AuthImportBotAuthorizationRequest, *tg.AuthSignInRequest:
			return &tg.AuthAuthorization{User: &tg.User{ID: 1}}, nil
		case *tg.AuthLogOutRequest:
			return &tg.AuthLoggedOut{}, nil
		}
		if err := fail.Load(); err != nil {
			return nil, err
		}
		return &tg.Config{}, nil
	}
}

func TestClientOnAuthLost(t *testing.T) {
	ctx := context.Background()
	var (
		fail     atomic.Pointer[tgerr.Error]
		reported []error
		stop     error
	)
	client := newTestClient(authLostHandler(&fail))
	client.authLost = make(chan error, 1)
	client.onAuthLost = func(ctx context.Context, err error) error {
		reported = append(reported, err)
		return stop
	}
	api := client.API()
	signIn := func() {
		_, err := api.AuthImportBotAuthorization(ctx, &tg.AuthImportBotAuthorizationRequest{})
		require.NoError(t, err)
	}

	// Not authorized yet: 401 is expected and not reported.
	fail.Store(tgerr.New(401, "AUTH_KEY_UNREGISTERED"))
	_, err := api.HelpGetConfig(ctx)
	require.Error(t, err)
	require.Empty(t, reported)

	// Not lost authorization.
	signIn()
	fail.Store(tgerr.New(401, "SESSION_PASSWORD_NEEDED"))
	_, _ = api.HelpGetConfig(ctx)
	require.Empty(t, reported)

	// Lost: reported once, handler returns nil, so client keeps running.
	fail.Store(tgerr.New(401, "SESSION_REVOKED"))
	_, _ = api.HelpGetConfig(ctx)
	_, _ = api.HelpGetConfig(ctx)
	require.Len(t, reported, 1)
	require.True(t, tgerr.Is(reported[0], "SESSION_REVOKED"))
	require.Empty(t, client.authLost)

	// After log out: not reported.
	signIn()
	_, err = api.AuthLogOut(ctx)
	require.NoError(t, err)
	_, _ = api.HelpGetConfig(ctx)
	require.Len(t, reported, 1)

	// Reported again after next sign in, and stops Run.
	signIn()
	stop = errors.New("stop account")
	fail.Store(tgerr.New(406, "AUTH_KEY_DUPLICATED"))
	_, _ = api.HelpGetConfig(ctx)
	require.Len(t, reported, 2)
	require.Len(t, client.authLost, 1)

	var lost *AuthLostError
	require.ErrorAs(t, <-client.authLost, &lost)
	require.True(t, tgerr.Is(lost.Err, "AUTH_KEY_DUPLICATED"))
	require.ErrorIs(t, lost, stop)
}

func TestClientRunStopsOnAuthLost(t *testing.T) {
	var fail atomic.Pointer[tgerr.Error]
	stop := errors.New("stop account")
	c := NewClient(1, "hash", Options{
		NoUpdates: true,
		Logger:    log.Nop,
		OnAuthLost: func(ctx context.Context, err error) error {
			return stop
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
		_, _ = c.API().HelpGetConfig(ctx)
		<-ctx.Done()
		return ctx.Err()
	})

	var lost *AuthLostError
	require.ErrorAs(t, err, &lost)
	require.ErrorIs(t, err, stop)
	require.True(t, IsAuthLost(err))
}

type authRunConn struct {
	conn   clientConn
	client *Client
}

func (a authRunConn) Run(ctx context.Context) error {
	a.client.onReady()
	<-ctx.Done()
	return ctx.Err()
}

func (a authRunConn) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	return a.conn.Invoke(ctx, input, output)
}

func (a authRunConn) Ping(ctx context.Context) error { return nil }

