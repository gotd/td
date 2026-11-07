package telegram

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/log"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/telegram/internal/manager"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func TestConnKindString(t *testing.T) {
	for k, s := range map[ConnKind]string{
		ConnKindPrimary: "primary",
		ConnKindData:    "data",
		ConnKindCDN:     "cdn",
		ConnKind(-1):    "unknown",
	} {
		require.Equal(t, s, k.String())
	}
}

func TestClientConnDeadEvent(t *testing.T) {
	testErr := errors.New("test")

	var (
		dead   []error
		events []ConnDeadEvent
		// onDeadOf captures OnDead of created connections.
		onDeadOf []func(error)
	)
	c := &Client{
		log:        log.For(log.Nop),
		session:    pool.NewSyncSession(pool.Session{DC: 2}),
		onDead:     func(err error) { dead = append(dead, err) },
		onConnDead: func(e ConnDeadEvent) { events = append(events, e) },
	}
	c.init()
	c.create = func(
		_ mtproto.Dialer,
		_ manager.ConnMode,
		_ int,
		_ mtproto.Options,
		opts manager.ConnOptions,
	) pool.Conn {
		onDeadOf = append(onDeadOf, opts.OnDead)
		return nil
	}

	// Primary and primary-session pool connections report the DC they were
	// created for, even after migration.
	c.createPrimaryConn(nil)
	c.primaryConn(nil)
	require.Len(t, onDeadOf, 2)
	c.session.Store(pool.Session{DC: 4})
	onDeadOf[0](testErr)
	onDeadOf[1](testErr)

	c.handleDCConnDead(5, testErr)
	c.handleCDNConnDead(203, testErr)

	require.Equal(t, []ConnDeadEvent{
		{DC: 2, Kind: ConnKindPrimary, Err: testErr},
		{DC: 2, Kind: ConnKindData, Err: testErr},
		{DC: 5, Kind: ConnKindData, Err: testErr},
		{DC: 203, Kind: ConnKindCDN, Err: testErr},
	}, events)
	require.Equal(t, []error{testErr, testErr, testErr, testErr}, dead,
		"OnDead must still be called for every connection")
}

// TestClientConnectionStateEventOnly checks that a client with only
// OnConnectionStateEvent set gets all states, Ready included, and that states
// of the primary connection are reported for its DC across migration.
func TestClientConnectionStateEventOnly(t *testing.T) {
	var (
		mux    sync.Mutex
		events []ConnectionStateEvent
	)
	client := newMigrationClient(t, func(id int64, dc int, body bin.Encoder) (bin.Encoder, error) {
		switch body.(type) {
		case *tg.UsersGetUsersRequest:
			return nil, tgerr.New(401, "AUTH_KEY_UNREGISTERED")
		case *tg.AuthLogOutRequest:
			if dc == 2 {
				return nil, tgerr.New(303, "USER_MIGRATE_10")
			}
			return &tg.BoolTrue{}, nil
		default:
			return nil, errors.Errorf("unexpected body %T", body)
		}
	})
	client.onConnectionStateEvent = func(e ConnectionStateEvent) {
		mux.Lock()
		events = append(events, e)
		mux.Unlock()
	}

	require.NoError(t, client.Run(context.Background(), func(ctx context.Context) error {
		var result tg.BoolTrue
		if err := client.Invoke(ctx, &tg.AuthLogOutRequest{}, &result); err != nil {
			return err
		}
		// Wait for Ready of the new primary connection.
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			mux.Lock()
			n := len(events)
			ready := n > 0 && events[n-1] == ConnectionStateEvent{DC: 10, State: ConnectionStateReady}
			mux.Unlock()
			if ready {
				return nil
			}
			time.Sleep(time.Millisecond)
		}
		return errors.New("no ready state of new primary connection")
	}))

	mux.Lock()
	defer mux.Unlock()
	require.Equal(t, []ConnectionStateEvent{
		{DC: 2, State: ConnectionStateConnecting},
		{DC: 2, State: ConnectionStateReady},
		{DC: 2, State: ConnectionStateDisconnected},
		{DC: 10, State: ConnectionStateConnecting},
		{DC: 10, State: ConnectionStateReady},
	}, events[:5])
}
