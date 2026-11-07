package telegram

import (
	"context"
	"sync"
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

// authFlowHandler answers sign in requests with given results, users.getUsers
// with a user, and fails all requests with fail when set.
func authFlowHandler(fail *atomic.Pointer[tgerr.Error], signIn tg.AuthAuthorizationClass, token tg.AuthLoginTokenClass) testHandler {
	return func(id int64, body bin.Encoder) (bin.Encoder, error) {
		if err := fail.Load(); err != nil {
			return nil, err
		}
		switch unwrapQuery(body).(type) {
		case *tg.AuthSignInRequest:
			return signIn, nil
		case *tg.AuthExportLoginTokenRequest:
			return token, nil
		case *tg.UsersGetUsersRequest:
			return &tg.UserClassVector{Elems: []tg.UserClass{&tg.User{ID: 1, Self: true}}}, nil
		case *tg.MessagesGetDialogsRequest:
			return &tg.MessagesDialogs{}, nil
		}
		return &tg.Config{}, nil
	}
}

func TestClientAuthorizedBy(t *testing.T) {
	ctx := context.Background()
	authorization := &tg.AuthAuthorization{User: &tg.User{ID: 1}}
	for _, tt := range []struct {
		name       string
		signIn     tg.AuthAuthorizationClass
		token      tg.AuthLoginTokenClass
		do         func(api *tg.Client) error
		authorized bool
	}{
		{
			name: "Self",
			do: func(api *tg.Client) error {
				_, err := api.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}})
				return err
			},
			authorized: true,
		},
		{
			name: "WrappedSelf",
			do: func(api *tg.Client) error {
				var users tg.UserClassVector
				return api.Invoker().Invoke(ctx, &tg.InvokeWithoutUpdatesRequest{
					Query: &tg.UsersGetUsersRequest{ID: []tg.InputUserClass{&tg.InputUserSelf{}}},
				}, &users)
			},
			authorized: true,
		},
		{
			name: "WrappedHelp",
			do: func(api *tg.Client) error {
				var cfg tg.Config
				return api.Invoker().Invoke(ctx, &tg.InvokeWithoutUpdatesRequest{
					Query: &tg.HelpGetConfigRequest{},
				}, &cfg)
			},
		},
		{
			name:   "SignIn",
			signIn: authorization,
			do: func(api *tg.Client) error {
				_, err := api.AuthSignIn(ctx, &tg.AuthSignInRequest{})
				return err
			},
			authorized: true,
		},
		{
			name:   "SignUpRequired",
			signIn: &tg.AuthAuthorizationSignUpRequired{},
			do: func(api *tg.Client) error {
				_, err := api.AuthSignIn(ctx, &tg.AuthSignInRequest{})
				return err
			},
		},
		{
			name:  "LoginTokenSuccess",
			token: &tg.AuthLoginTokenSuccess{Authorization: authorization},
			do: func(api *tg.Client) error {
				_, err := api.AuthExportLoginToken(ctx, &tg.AuthExportLoginTokenRequest{})
				return err
			},
			authorized: true,
		},
		{
			name:  "LoginToken",
			token: &tg.AuthLoginToken{Token: []byte("token")},
			do: func(api *tg.Client) error {
				_, err := api.AuthExportLoginToken(ctx, &tg.AuthExportLoginTokenRequest{})
				return err
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var (
				fail     atomic.Pointer[tgerr.Error]
				reported int
			)
			client := newTestClient(authFlowHandler(&fail, tt.signIn, tt.token))
			client.onAuthLost = func(ctx context.Context, err error) error {
				reported++
				return nil
			}
			api := client.API()
			require.NoError(t, tt.do(api))
			require.Equal(t, tt.authorized, client.auth.authorized.Load())

			fail.Store(tgerr.New(401, "AUTH_KEY_UNREGISTERED"))
			_, _ = api.HelpGetConfig(ctx)
			if tt.authorized {
				require.Equal(t, 1, reported)
			} else {
				require.Zero(t, reported)
			}
		})
	}
}

type authCtxKey struct{}

func TestClientOnAuthLostContext(t *testing.T) {
	var fail atomic.Pointer[tgerr.Error]
	client := newTestClient(authFlowHandler(&fail, nil, nil))
	client.ctx = context.WithValue(context.Background(), authCtxKey{}, "run")
	var got context.Context
	client.onAuthLost = func(ctx context.Context, err error) error {
		got = ctx
		return nil
	}
	api := client.API()
	_, err := api.UsersGetUsers(context.Background(), []tg.InputUserClass{&tg.InputUserSelf{}})
	require.NoError(t, err)

	reqCtx := context.WithValue(context.Background(), authCtxKey{}, "request")
	fail.Store(tgerr.New(401, "SESSION_REVOKED"))
	_, _ = api.HelpGetConfig(reqCtx)
	require.NotNil(t, got)
	require.Equal(t, "run", got.Value(authCtxKey{}))
}

// TestClientRunRestoredSessionAuthLost checks that a session restored from
// storage with NoUpdates (no sign in, no connect-time Self) is covered.
func TestClientRunRestoredSessionAuthLost(t *testing.T) {
	var fail atomic.Pointer[tgerr.Error]
	stop := errors.New("stop account")
	c := NewClient(1, "hash", Options{
		NoUpdates: true,
		Logger:    log.Nop,
		OnAuthLost: func(ctx context.Context, err error) error {
			return stop
		},
	})
	conn := newTestClient(authFlowHandler(&fail, nil, nil)).conn
	c.conn = authRunConn{conn: conn, client: c}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := c.Run(ctx, func(ctx context.Context) error {
		if _, err := c.Self(ctx); err != nil {
			return err
		}
		fail.Store(tgerr.New(401, "AUTH_KEY_UNREGISTERED"))
		_, _ = c.API().MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{OffsetPeer: &tg.InputPeerEmpty{}})
		<-ctx.Done()
		return ctx.Err()
	})

	var lost *AuthLostError
	require.ErrorAs(t, err, &lost)
	require.ErrorIs(t, err, stop)
}

// TestClientConnectSelfAuthLost checks reporting of lost authorization found
// by the connect-time Self of an authorized client, e.g. on reconnect.
func TestClientConnectSelfAuthLost(t *testing.T) {
	for _, tt := range []struct {
		name   string
		stop   error
		events []string
	}{
		{name: "Continue", events: []string{"auth lost", "self error"}},
		{name: "Stop", stop: errors.New("stop"), events: []string{"auth lost"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var (
				fail   atomic.Pointer[tgerr.Error]
				mux    sync.Mutex
				events []string
			)
			add := func(e string) {
				mux.Lock()
				defer mux.Unlock()
				events = append(events, e)
			}
			client := newTestClient(authFlowHandler(&fail, nil, nil))
			client.onAuthLost = func(ctx context.Context, err error) error {
				add("auth lost")
				return tt.stop
			}
			selfDone := make(chan struct{})
			client.onSelfError = func(ctx context.Context, err error) error {
				add("self error")
				close(selfDone)
				return nil
			}
			client.markAuthorized()
			fail.Store(tgerr.New(401, "AUTH_KEY_UNREGISTERED"))

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() { var dc int; done <- client.runUntilRestart(ctx, &dc) }()
			if tt.stop != nil {
				select {
				case err := <-client.authLost:
					require.ErrorIs(t, err, tt.stop)
				case <-ctx.Done():
					t.Fatal("timeout")
				}
			} else {
				select {
				case <-selfDone:
				case <-ctx.Done():
					t.Fatal("timeout")
				}
			}
			// Wait for the Self goroutine to finish.
			cancel()
			<-done

			mux.Lock()
			defer mux.Unlock()
			require.Equal(t, tt.events, events)
		})
	}
}
