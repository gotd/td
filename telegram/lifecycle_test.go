package telegram

import (
	"context"
	"net"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log"
	"github.com/gotd/log/logzap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgtest"
	"github.com/gotd/td/tgtest/cluster"
	"github.com/gotd/td/transport"
)

const lifecycleDC = 2

// lifecycleCluster runs a single DC cluster until the end of the test and
// returns options of a client connecting to it.
func lifecycleCluster(t *testing.T, setup func(d *tgtest.Dispatcher)) Options {
	t.Helper()
	zapLog := zaptest.NewLogger(t)

	c := cluster.NewCluster(cluster.Options{
		Logger:   logzap.New(zapLog.Named("cluster")),
		Protocol: transport.Intermediate,
	})
	d := c.Dispatch(lifecycleDC, "dc")
	if setup != nil {
		setup(d)
	}

	ctx, cancel := context.WithCancel(context.Background())
	upDone := make(chan error, 1)
	go func() { upDone <- c.Up(ctx) }()
	t.Cleanup(func() {
		cancel()
		if err := <-upDone; err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("cluster: %v", err)
		}
	})
	select {
	case <-c.Ready():
	case err := <-upDone:
		t.Fatalf("cluster: %v", err)
	}

	return Options{
		PublicKeys:     c.Keys(),
		Resolver:       c.Resolver(),
		DCList:         c.List(),
		DC:             lifecycleDC,
		Logger:         logzap.New(zapLog.Named("client")),
		SessionStorage: &session.StorageMemory{},
		NoUpdates:      true,
		ReconnectionBackoff: func() backoff.BackOff {
			return backoff.NewConstantBackOff(10 * time.Millisecond)
		},
	}
}

// fakeRunClient returns a client whose primary connection is ready at once
// and answers requests with h, without network.
func fakeRunClient(h testHandler) *Client {
	c := NewClient(1, "hash", Options{NoUpdates: true, Logger: log.Nop})
	c.conn = authRunConn{conn: newTestClient(h).conn, client: c}
	return c
}

func configHandler(int64, bin.Encoder) (bin.Encoder, error) {
	return &tg.Config{ThisDC: lifecycleDC}, nil
}

// clientGoroutines returns stacks of goroutines of the client by ID: those
// running gotd code except the test server and this test.
func clientGoroutines() map[string]string {
	buf := make([]byte, 1<<22)
	buf = buf[:runtime.Stack(buf, true)]
	r := map[string]string{}
	for _, g := range strings.Split(string(buf), "\n\n") {
		if !strings.Contains(g, "github.com/gotd/td/") ||
			strings.Contains(g, "github.com/gotd/td/tgtest") ||
			strings.Contains(g, "lifecycleCluster") ||
			strings.Contains(g, "clientGoroutines") {
			continue
		}
		id, _, _ := strings.Cut(g, " [")
		r[id] = g
	}
	return r
}

// goroutineLeakCheck returns a function that waits until client goroutines
// started after the call to goroutineLeakCheck are gone. Goroutines left by
// other tests are ignored.
func goroutineLeakCheck(t *testing.T) func() {
	t.Helper()
	before := clientGoroutines()
	return func() {
		t.Helper()
		deadline := time.Now().Add(10 * time.Second)
		for {
			var left []string
			for id, g := range clientGoroutines() {
				if _, ok := before[id]; !ok {
					left = append(left, g)
				}
			}
			if len(left) == 0 {
				return
			}
			if time.Now().After(deadline) {
				t.Fatalf("%d client goroutines leaked:\n%s", len(left), strings.Join(left, "\n\n"))
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestClientStartClose(t *testing.T) {
	opts := lifecycleCluster(t, nil)
	checkLeaks := goroutineLeakCheck(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := Connect(ctx, 1, "hash", opts)
	require.NoError(t, err)
	select {
	case <-client.Done():
		t.Fatal("done before Close")
	default:
	}

	// Calls after Start are race-free, also concurrent ones.
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg, err := client.API().HelpGetConfig(ctx)
			if err != nil {
				t.Errorf("getConfig: %v", err)
				return
			}
			if cfg.ThisDC != lifecycleDC {
				t.Errorf("ThisDC: %d", cfg.ThisDC)
			}
		}()
	}
	wg.Wait()

	require.NoError(t, client.Close())
	require.NoError(t, client.Close(), "Close is idempotent")
	<-client.Done()
	require.NoError(t, client.Wait(ctx))
	checkLeaks()
}

// TestClientStartStartupContext checks that the startup ctx does not bound
// the lifetime of the started client.
func TestClientStartStartupContext(t *testing.T) {
	client := fakeRunClient(configHandler)
	ctx, cancel := context.WithCancel(context.Background())
	require.NoError(t, client.Start(ctx))
	cancel()

	_, err := client.API().HelpGetConfig(context.Background())
	require.NoError(t, err)
	select {
	case <-client.Done():
		t.Fatal("client stopped with startup ctx")
	case <-time.After(50 * time.Millisecond):
	}
	require.NoError(t, client.Close())
}

func TestClientStartWithLifetime(t *testing.T) {
	client := fakeRunClient(configHandler)
	lifetime, cancel := context.WithCancel(context.Background())
	require.NoError(t, client.Start(context.Background(), WithLifetime(lifetime)))
	cancel()

	select {
	case <-client.Done():
	case <-time.After(10 * time.Second):
		t.Fatal("client did not stop on lifetime cancel")
	}
	require.NoError(t, client.Wait(context.Background()))
	require.NoError(t, client.Close())
}

func TestClientStartTimeoutUnreachable(t *testing.T) {
	// Closed port: every connection attempt is refused.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().(*net.TCPAddr)
	require.NoError(t, l.Close())

	checkLeaks := goroutineLeakCheck(t)
	client := NewClient(1, "hash", Options{
		Logger:    log.Nop,
		NoUpdates: true,
		DC:        lifecycleDC,
		DCList: dcs.List{Options: []tg.DCOption{{
			ID:        lifecycleDC,
			IPAddress: addr.IP.String(),
			Port:      addr.Port,
		}}},
		ReconnectionBackoff: func() backoff.BackOff {
			return backoff.NewConstantBackOff(10 * time.Millisecond)
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	err = client.Start(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	<-client.Done()
	require.ErrorIs(t, client.Wait(context.Background()), context.DeadlineExceeded)
	require.NoError(t, client.Close(), "Start already returned the error")
	checkLeaks()

	require.Error(t, client.Start(context.Background()), "single use")
}

func TestClientStartCanceled(t *testing.T) {
	client := fakeRunClient(configHandler)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, client.Start(ctx), context.Canceled)
	<-client.Done()
	require.NoError(t, client.Close())
}

func TestClientStartConfigError(t *testing.T) {
	opts := lifecycleCluster(t, func(d *tgtest.Dispatcher) {
		d.HandleFunc(tg.HelpGetConfigRequestTypeID, func(server *tgtest.Server, req *tgtest.Request) error {
			return server.SendErr(req, tgerr.New(400, "API_ID_INVALID"))
		})
	})
	checkLeaks := goroutineLeakCheck(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client := NewClient(1, "hash", opts)
	err := client.Start(ctx)
	var cfgErr *ConfigError
	require.ErrorAs(t, err, &cfgErr, "got %v", err)
	require.True(t, cfgErr.Err.IsType("API_ID_INVALID"))
	require.NotErrorIs(t, err, context.DeadlineExceeded)

	<-client.Done()
	require.ErrorAs(t, client.Wait(ctx), &cfgErr)
	require.NoError(t, client.Close())
	checkLeaks()

	_, err = Connect(ctx, 1, "hash", opts)
	require.ErrorAs(t, err, &cfgErr)
}

// TestClientStartAuthLost checks that a failure after Start is reported via
// Done and Wait without Close.
func TestClientStartAuthLost(t *testing.T) {
	var lost atomic.Bool
	opts := lifecycleCluster(t, func(d *tgtest.Dispatcher) {
		d.HandleFunc(tg.MessagesGetDialogsRequestTypeID, func(server *tgtest.Server, req *tgtest.Request) error {
			if lost.Load() {
				// Server-side drop of the authorization.
				return server.SendErr(req, tgerr.New(401, "AUTH_KEY_UNREGISTERED"))
			}
			return server.SendResult(req, &tg.MessagesDialogs{})
		})
	})
	reason := errors.New("stop account")
	opts.OnAuthLost = func(ctx context.Context, err error) error {
		return reason
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := Connect(ctx, 1, "hash", opts)
	require.NoError(t, err)

	req := &tg.MessagesGetDialogsRequest{OffsetPeer: &tg.InputPeerEmpty{}}
	_, err = client.API().MessagesGetDialogs(ctx, req) // Authorized.
	require.NoError(t, err)
	lost.Store(true)
	_, err = client.API().MessagesGetDialogs(ctx, req)
	require.True(t, IsAuthLost(err), "got %v", err)

	select {
	case <-client.Done():
	case <-ctx.Done():
		t.Fatal("client did not stop on lost authorization")
	}
	var lostErr *AuthLostError
	require.ErrorAs(t, client.Wait(ctx), &lostErr)
	require.ErrorIs(t, lostErr.Reason, reason)
	err = client.Close()
	require.ErrorAs(t, err, &lostErr, "Close returns the error that stopped the client")
}

func TestClientWaitContext(t *testing.T) {
	client := fakeRunClient(configHandler)
	require.NoError(t, client.Start(context.Background()))
	defer func() { require.NoError(t, client.Close()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	require.ErrorIs(t, client.Wait(ctx), context.DeadlineExceeded)
}

func TestClientCloseConcurrent(t *testing.T) {
	t.Run("Started", func(t *testing.T) {
		client := fakeRunClient(configHandler)
		require.NoError(t, client.Start(context.Background()))

		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := client.Close(); err != nil {
					t.Errorf("Close: %v", err)
				}
			}()
		}
		wg.Wait()
		<-client.Done()
	})
	t.Run("DuringStart", func(t *testing.T) {
		// Connection is never ready, so Start blocks until Close.
		client := NewClient(1, "hash", Options{NoUpdates: true, Logger: log.Nop})
		client.conn = blockingConn{}

		started := make(chan error, 1)
		go func() { started <- client.Start(context.Background()) }()
		var wg sync.WaitGroup
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := client.Close(); err != nil {
					t.Errorf("Close: %v", err)
				}
			}()
		}
		wg.Wait()
		select {
		case err := <-started:
			require.ErrorIs(t, err, errClosed)
		case <-time.After(10 * time.Second):
			t.Fatal("Close did not stop Start")
		}
		require.NoError(t, client.Close())
	})
}

// blockingConn is a connection that never gets ready.
type blockingConn struct{}

func (blockingConn) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (blockingConn) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	<-ctx.Done()
	return ctx.Err()
}

func (blockingConn) Ping(ctx context.Context) error { return nil }

func TestClientCloseBeforeStart(t *testing.T) {
	client := fakeRunClient(configHandler)
	require.NoError(t, client.Close())
	require.NoError(t, client.Close())
	<-client.Done()
	require.NoError(t, client.Wait(context.Background()))

	require.ErrorIs(t, client.Start(context.Background()), errClosed)
	require.ErrorIs(t, client.Run(context.Background(), func(ctx context.Context) error {
		return nil
	}), errClosed)
}

func TestClientStartRunExclusive(t *testing.T) {
	t.Run("RunAfterStart", func(t *testing.T) {
		client := fakeRunClient(configHandler)
		require.NoError(t, client.Start(context.Background()))
		err := client.Run(context.Background(), func(ctx context.Context) error {
			t.Error("callback called")
			return nil
		})
		require.ErrorIs(t, err, errDrivenByStart)
		require.ErrorIs(t, client.Start(context.Background()), errAlreadyStarted)
		require.NoError(t, client.Close())
		require.ErrorIs(t, client.Start(context.Background()), errAlreadyStarted)
	})
	t.Run("StartAfterRun", func(t *testing.T) {
		client := fakeRunClient(configHandler)
		require.NoError(t, client.Run(context.Background(), func(ctx context.Context) error {
			require.ErrorIs(t, client.Start(ctx), errDrivenByRun)
			require.ErrorIs(t, client.Close(), errDrivenByRun)
			return nil
		}))
		require.ErrorIs(t, client.Start(context.Background()), errDrivenByRun)
	})
}
