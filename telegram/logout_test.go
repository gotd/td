package telegram

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/session"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// guestPool is a cached non-primary DC pool that records auth.logOut and Close.
type guestPool struct {
	loggedOut atomic.Bool
	closed    atomic.Bool
}

func (g *guestPool) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	if _, ok := input.(*tg.AuthLogOutRequest); ok {
		g.loggedOut.Store(true)
		var b bin.Buffer
		if err := (&tg.AuthLoggedOut{}).Encode(&b); err != nil {
			return err
		}
		return output.Decode(&b)
	}
	return nil
}

func (g *guestPool) Close() error {
	g.closed.Store(true)
	return nil
}

// logOutHandler answers sign-in, auth.logOut with logOutResult/logOutErr, and
// everything else with fail or an empty config.
func logOutHandler(logOutErr *atomic.Pointer[tgerr.Error], fail *atomic.Pointer[tgerr.Error]) testHandler {
	return func(id int64, body bin.Encoder) (bin.Encoder, error) {
		switch body.(type) {
		case *tg.AuthImportBotAuthorizationRequest:
			return &tg.AuthAuthorization{User: &tg.User{ID: 1}}, nil
		case *tg.AuthLogOutRequest:
			if err := logOutErr.Load(); err != nil {
				return nil, err
			}
			return &tg.AuthLoggedOut{FutureAuthToken: []byte{1, 2, 3}}, nil
		}
		if err := fail.Load(); err != nil {
			return nil, err
		}
		return &tg.Config{}, nil
	}
}

func newLogOutTestClient(t *testing.T, logOutErr, fail *atomic.Pointer[tgerr.Error]) (*Client, *session.StorageMemory) {
	t.Helper()
	client := newTestClient(logOutHandler(logOutErr, fail))
	storage := &session.StorageMemory{}
	client.storage = &session.Loader{Storage: storage}
	client.session = pool.NewSyncSession(pool.Session{DC: 2})
	require.NoError(t, client.saveSession(tg.Config{ThisDC: 2}, mtproto.Session{}))
	_, err := client.storage.Load(context.Background())
	require.NoError(t, err)
	return client, storage
}

func TestClientLogOut(t *testing.T) {
	ctx := context.Background()
	var logOutErr, fail atomic.Pointer[tgerr.Error]
	client, _ := newLogOutTestClient(t, &logOutErr, &fail)

	var reported int
	client.authLost = make(chan error, 1)
	client.onAuthLost = func(ctx context.Context, err error) error {
		reported++
		return nil
	}

	_, err := client.API().AuthImportBotAuthorization(ctx, &tg.AuthImportBotAuthorizationRequest{})
	require.NoError(t, err)
	require.True(t, client.auth.authorized.Load())

	guest, homeFile := &guestPool{}, &guestPool{}
	client.subConns.conns[4] = guest
	client.filePools.conns[2] = homeFile
	client.sessions[4] = pool.NewSyncSession(pool.Session{DC: 4})

	res, err := client.LogOut(ctx)
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3}, res.FutureAuthToken)
	require.False(t, client.auth.authorized.Load())

	// Guest DC: logged out and closed; home DC file pool: closed only.
	require.True(t, guest.loggedOut.Load())
	require.True(t, guest.closed.Load())
	require.False(t, homeFile.loggedOut.Load())
	require.True(t, homeFile.closed.Load())
	require.Zero(t, client.subConns.len())
	require.Zero(t, client.filePools.len())
	require.Empty(t, client.sessions)

	// Session wiped and not saved again by new sessions.
	_, err = client.storage.Load(ctx)
	require.ErrorIs(t, err, session.ErrNotFound)
	require.NoError(t, client.saveSession(tg.Config{ThisDC: 2}, mtproto.Session{}))
	_, err = client.storage.Load(ctx)
	require.ErrorIs(t, err, session.ErrNotFound)

	// Lost authorization after log out is not reported.
	fail.Store(tgerr.New(401, "AUTH_KEY_UNREGISTERED"))
	_, _ = client.API().HelpGetConfig(ctx)
	require.NoError(t, client.waitAuthLost(ctx))
	require.Zero(t, reported)
	fail.Store(nil)

	// Signing in again saves the session again.
	_, err = client.API().AuthImportBotAuthorization(ctx, &tg.AuthImportBotAuthorizationRequest{})
	require.NoError(t, err)
	require.NoError(t, client.saveSession(tg.Config{ThisDC: 2}, mtproto.Session{}))
	_, err = client.storage.Load(ctx)
	require.NoError(t, err)
}

func TestClientLogOutUnauthorizedIsSuccess(t *testing.T) {
	ctx := context.Background()
	var logOutErr, fail atomic.Pointer[tgerr.Error]
	client, _ := newLogOutTestClient(t, &logOutErr, &fail)

	logOutErr.Store(tgerr.New(401, "AUTH_KEY_UNREGISTERED"))
	res, err := client.LogOut(ctx)
	require.NoError(t, err)
	require.Empty(t, res.FutureAuthToken)
	_, err = client.storage.Load(ctx)
	require.ErrorIs(t, err, session.ErrNotFound)
}

func TestClientLogOutErrorStillWipes(t *testing.T) {
	ctx := context.Background()
	var logOutErr, fail atomic.Pointer[tgerr.Error]
	client, _ := newLogOutTestClient(t, &logOutErr, &fail)
	guest := &guestPool{}
	client.subConns.conns[4] = guest

	logOutErr.Store(tgerr.New(500, "INTERNAL"))
	_, err := client.LogOut(ctx)
	require.True(t, tgerr.Is(err, "INTERNAL"), "unexpected error: %v", err)
	require.False(t, client.auth.authorized.Load())
	require.True(t, guest.closed.Load())
	_, err = client.storage.Load(ctx)
	require.ErrorIs(t, err, session.ErrNotFound)
}
