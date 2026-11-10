package telegram

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/session"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/telegram/internal/manager"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgtest"
	"github.com/gotd/td/tgtest/cluster"
	"github.com/gotd/td/transport"
)

func TestAsConfigError(t *testing.T) {
	initErr := func(err error) error {
		return errors.Wrap(&manager.InitError{Err: errors.Wrap(err, "invoke")}, "run")
	}
	for _, tt := range []struct {
		name string
		err  error
		ok   bool
	}{
		{"APIIDInvalid", initErr(tgerr.New(400, "API_ID_INVALID")), true},
		{"APIIDPublishedFlood", initErr(tgerr.New(400, "API_ID_PUBLISHED_FLOOD")), true},
		{"ConnectionAPIIDInvalid", initErr(tgerr.New(400, "CONNECTION_API_ID_INVALID")), true},
		{"ConnectionDeviceModelEmpty", initErr(tgerr.New(400, "CONNECTION_DEVICE_MODEL_EMPTY")), true},
		{"UpdateAppToLogin", initErr(tgerr.New(406, "UPDATE_APP_TO_LOGIN")), true},
		{"ConnectionNotInited", initErr(tgerr.New(400, "CONNECTION_NOT_INITED")), false},
		{"ConnectionLayerInvalid", initErr(tgerr.New(400, "CONNECTION_LAYER_INVALID")), false},
		{"AuthKeyDuplicated", initErr(tgerr.New(406, "AUTH_KEY_DUPLICATED")), false},
		{"Unauthorized", initErr(tgerr.New(401, "AUTH_KEY_UNREGISTERED")), false},
		{"Internal", initErr(tgerr.New(500, "INTERNAL")), false},
		{"FloodWait", initErr(tgerr.New(420, "FLOOD_WAIT_10")), false},
		{"Network", initErr(errors.New("connection reset")), false},
		{"NotInit", errors.Wrap(tgerr.New(400, "API_ID_INVALID"), "setup"), false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfgErr, ok := asConfigError(tt.err)
			require.Equal(t, tt.ok, ok)
			if ok {
				rpcErr, _ := tgerr.As(tt.err)
				require.Same(t, rpcErr, cfgErr.Err)
				require.True(t, tgerr.Is(cfgErr, rpcErr.Type))
			}
		})
	}
}

// runWithInitError runs a client against a cluster whose DC answers
// initConnection with errors returned by initErr (nil means success).
// lastConnState is the last primary connection state seen by runWithInitError.
var lastConnState ConnectionState

func runWithInitError(t *testing.T, initErr func(n int64) *tgerr.Error) (attempts int64, called bool, err error) {
	t.Helper()
	const dcID = 2
	var lastState atomic.Int64

	zapLog := zaptest.NewLogger(t)
	// Bounds the reconnect loop if the error is not treated as fatal.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c := cluster.NewCluster(cluster.Options{
		Logger:   logzap.New(zapLog.Named("cluster")),
		Protocol: transport.Intermediate,
	})
	var inits atomic.Int64
	c.Dispatch(dcID, "dc").HandleFunc(tg.HelpGetConfigRequestTypeID,
		func(server *tgtest.Server, req *tgtest.Request) error {
			if e := initErr(inits.Add(1)); e != nil {
				return server.SendErr(req, e)
			}
			return server.SendResult(req, &tg.Config{ThisDC: dcID})
		},
	)

	var runErr error
	g := tdsync.NewCancellableGroup(ctx)
	g.Go(c.Up)
	g.Go(func(ctx context.Context) error {
		defer cancel()
		select {
		case <-c.Ready():
		case <-ctx.Done():
			return ctx.Err()
		}

		client := NewClient(1, "hash", Options{
			PublicKeys:     c.Keys(),
			Resolver:       c.Resolver(),
			DCList:         c.List(),
			DC:             dcID,
			Logger:         logzap.New(zapLog.Named("client")),
			SessionStorage: &session.StorageMemory{},
			NoUpdates:      true,
			OnConnectionState: func(st ConnectionState) {
				lastState.Store(int64(st))
			},
			ReconnectionBackoff: func() backoff.BackOff {
				return backoff.NewConstantBackOff(time.Millisecond)
			},
		})
		runErr = client.Run(ctx, func(ctx context.Context) error {
			called = true
			return nil
		})
		return nil
	})
	if waitErr := g.Wait(); waitErr != nil && !errors.Is(waitErr, context.Canceled) {
		t.Fatalf("cluster: %v", waitErr)
	}
	lastConnState = ConnectionState(lastState.Load())
	return inits.Load(), called, runErr
}

func TestClientRunInitConfigError(t *testing.T) {
	for _, typ := range []string{
		"API_ID_INVALID",
		"API_ID_PUBLISHED_FLOOD",
		"CONNECTION_API_ID_INVALID",
	} {
		t.Run(typ, func(t *testing.T) {
			attempts, called, err := runWithInitError(t, func(int64) *tgerr.Error {
				return tgerr.New(400, typ)
			})
			require.Error(t, err)
			require.NotErrorIs(t, err, context.DeadlineExceeded, "client reconnected until timeout")
			var cfgErr *ConfigError
			require.ErrorAs(t, err, &cfgErr)
			require.True(t, cfgErr.Err.IsType(typ))
			require.True(t, tgerr.Is(err, typ))
			require.False(t, called)
			require.Equal(t, int64(1), attempts)
			require.Equal(t, ConnectionStateDisconnected, lastConnState,
				"connection state must end disconnected when Run stops")
		})
	}
}

func TestClientRunInitConnectionNotInitedRetried(t *testing.T) {
	attempts, called, err := runWithInitError(t, func(n int64) *tgerr.Error {
		if n == 1 {
			return tgerr.New(400, "CONNECTION_NOT_INITED")
		}
		return nil
	})
	require.NoError(t, err)
	require.True(t, called)
	require.Equal(t, int64(2), attempts)
}
