package telegram

import (
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/log"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/telegram/internal/manager"
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

func TestClientConnectionStateEvent(t *testing.T) {
	var (
		states []ConnectionState
		events []ConnectionStateEvent
	)
	c := &Client{
		session:                pool.NewSyncSession(pool.Session{DC: 2}),
		onConnectionState:      func(s ConnectionState) { states = append(states, s) },
		onConnectionStateEvent: func(e ConnectionStateEvent) { events = append(events, e) },
	}
	c.notifyConnectionState(ConnectionStateConnecting)
	c.session.Store(pool.Session{DC: 4})
	c.notifyConnectionState(ConnectionStateReady)

	require.Equal(t, []ConnectionState{ConnectionStateConnecting, ConnectionStateReady}, states)
	require.Equal(t, []ConnectionStateEvent{
		{DC: 2, State: ConnectionStateConnecting},
		{DC: 4, State: ConnectionStateReady},
	}, events)
}
