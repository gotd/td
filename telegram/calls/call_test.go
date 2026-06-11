package calls

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func validDHConfig(t *testing.T) *tg.MessagesDhConfig {
	t.Helper()
	p, ok := new(big.Int).SetString(rfc3526Prime2048, 16)
	require.True(t, ok)
	return &tg.MessagesDhConfig{G: 2, P: p.Bytes(), Random: make([]byte, keySize)}
}

func TestClientHandleRouting(t *testing.T) {
	ctx := context.Background()
	c := NewClient(tg.NewClient(tgmock.NewRequire(t)), Options{})

	var incoming *IncomingCall
	c.OnIncoming(func(in *IncomingCall) { incoming = in })

	require.NoError(t, c.Handle(ctx, &tg.UpdatePhoneCall{
		PhoneCall: &tg.PhoneCallRequested{ID: 7, AdminID: 5, Video: true},
	}))
	require.NotNil(t, incoming)
	require.Equal(t, int64(5), incoming.UserID())
	require.True(t, incoming.Video())

	cl := newCallState(true)
	c.setActive(cl)

	require.NoError(t, c.Handle(ctx, &tg.UpdatePhoneCall{PhoneCall: &tg.PhoneCallAccepted{ID: 1}}))
	require.NoError(t, c.Handle(ctx, &tg.UpdatePhoneCall{PhoneCall: &tg.PhoneCall{ID: 2}}))
	require.NoError(t, c.Handle(ctx, &tg.UpdatePhoneCall{PhoneCall: &tg.PhoneCallDiscarded{ID: 3}}))
	// Non-routed variants must not panic.
	require.NoError(t, c.Handle(ctx, &tg.UpdatePhoneCall{PhoneCall: &tg.PhoneCallWaiting{ID: 4}}))
	require.NoError(t, c.Handle(ctx, &tg.UpdatePhoneCall{PhoneCall: &tg.PhoneCallEmpty{ID: 5}}))

	requireRecv := func(name string, ok bool) {
		if !ok {
			t.Fatalf("%s channel did not receive", name)
		}
	}
	select {
	case acc := <-cl.accepted:
		require.Equal(t, int64(1), acc.ID)
	default:
		requireRecv("accepted", false)
	}
	select {
	case <-cl.confirmed:
	default:
		requireRecv("confirmed", false)
	}
	select {
	case <-cl.discarded:
	default:
		requireRecv("discarded", false)
	}
}

func TestClientRegister(t *testing.T) {
	ctx := context.Background()
	d := tg.NewUpdateDispatcher()
	c := NewClient(tg.NewClient(tgmock.NewRequire(t)), Options{})

	var incoming *IncomingCall
	c.OnIncoming(func(in *IncomingCall) { incoming = in })
	c.Register(d)

	require.NoError(t, d.Handle(ctx, &tg.Updates{
		Updates: []tg.UpdateClass{
			&tg.UpdatePhoneCall{PhoneCall: &tg.PhoneCallRequested{ID: 1, AdminID: 9}},
		},
	}))
	require.NotNil(t, incoming)
	require.Equal(t, int64(9), incoming.UserID())
}

func TestClientHandleSignalingDataNoCall(t *testing.T) {
	ctx := context.Background()
	c := NewClient(tg.NewClient(tgmock.NewRequire(t)), Options{})
	// No active call: must be dropped without error.
	require.NoError(t, c.HandleSignalingData(ctx, &tg.UpdatePhoneCallSignalingData{
		PhoneCallID: 1, Data: []byte{1, 2, 3},
	}))
}

func TestClientDiscard(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	mock.Expect().ThenResult(&tg.Updates{})
	c := NewClient(tg.NewClient(mock), Options{})

	cl := newCallState(true)
	cl.input = tg.InputPhoneCall{ID: 1, AccessHash: 2}
	c.setActive(cl)
	require.NoError(t, c.Discard(ctx, DiscardHangup))

	// Discarding with no active call is a no-op.
	require.NoError(t, c.Discard(ctx, DiscardHangup))
}

func TestIncomingCallReject(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	mock.Expect().ThenResult(&tg.Updates{})
	c := NewClient(tg.NewClient(mock), Options{})

	ic := &IncomingCall{client: c, req: &tg.PhoneCallRequested{ID: 5, AccessHash: 6, AdminID: 7}}
	require.Equal(t, int64(7), ic.UserID())
	require.False(t, ic.Video())
	require.NoError(t, ic.Reject(ctx))
}

func TestGetDHConfig(t *testing.T) {
	ctx := context.Background()

	mock := tgmock.NewRequire(t)
	mock.Expect().ThenResult(validDHConfig(t))
	dh, rnd, err := getDHConfig(ctx, tg.NewClient(mock))
	require.NoError(t, err)
	require.Equal(t, 2, dh.g)
	require.Len(t, rnd, keySize)

	bad := tgmock.NewRequire(t)
	bad.Expect().ThenResult(&tg.MessagesDhConfigNotModified{Random: make([]byte, keySize)})
	_, _, err = getDHConfig(ctx, tg.NewClient(bad))
	require.Error(t, err)
}

func TestRequestUnexpectedResponse(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	mock.Expect().ThenResult(validDHConfig(t))                                    // getDhConfig
	mock.Expect().ThenResult(&tg.PhonePhoneCall{PhoneCall: &tg.PhoneCallEmpty{}}) // requestCall (not waiting)
	c := NewClient(tg.NewClient(mock), Options{})

	_, err := c.Request(ctx, &tg.InputUser{UserID: 1, AccessHash: 2})
	require.Error(t, err)
	// The active call must be cleared on failure.
	c.mu.Lock()
	require.Nil(t, c.call)
	c.mu.Unlock()
}
