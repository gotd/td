package telegram

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gotd/neo"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func TestAuthLostKindOf(t *testing.T) {
	for _, tt := range []struct {
		err  error
		kind AuthLostKind
	}{
		{tgerr.New(401, "USER_DEACTIVATED_BAN"), AuthLostBanned},
		{tgerr.New(401, "USER_DEACTIVATED"), AuthLostDeleted},
		{tgerr.New(401, "SESSION_REVOKED"), AuthLostRevoked},
		{tgerr.New(401, "AUTH_KEY_UNREGISTERED"), AuthLostUnregistered},
		{tgerr.New(401, "SESSION_EXPIRED"), AuthLostExpired},
		{tgerr.New(401, "AUTH_KEY_INVALID"), AuthLostKeyInvalid},
		{tgerr.New(406, "AUTH_KEY_DUPLICATED"), AuthLostKeyInvalid},
		{AuthKeyDropError(&tg.UpdateServiceNotification{Type: "AUTH_KEY_DROP_DUPLICATE"}), AuthLostRevoked},
		{tgerr.New(400, "PEER_ID_INVALID"), AuthLostUnknown},
		{errors.New("not rpc"), AuthLostUnknown},
	} {
		require.Equal(t, tt.kind, AuthLostKindOf(tt.err), tt.err)
		require.Equal(t, tt.kind != AuthLostUnknown, IsAuthLost(tt.err), tt.err)
	}
	require.Equal(t, "banned", AuthLostBanned.String())
	require.Equal(t, AuthLostBanned, (&AuthLostError{Err: tgerr.New(401, "USER_DEACTIVATED_BAN")}).Kind())

	require.True(t, IsAccountBanned(tgerr.New(401, "USER_DEACTIVATED_BAN")))
	require.True(t, IsAccountBanned(tgerr.New(400, "PHONE_NUMBER_BANNED")))
	require.False(t, IsAccountBanned(tgerr.New(401, "USER_DEACTIVATED")))

	require.True(t, IsFrozen(tgerr.New(420, "FROZEN_METHOD_INVALID")))
	require.True(t, IsFrozen(tgerr.New(400, "FROZEN_PARTICIPANT_MISSING")))
	require.False(t, IsFrozen(tgerr.New(420, "FLOOD_WAIT_3")))
	_, flood := tgerr.AsFloodWait(tgerr.New(420, "FROZEN_METHOD_INVALID"))
	require.False(t, flood, "frozen error is not a flood wait")
}

func freezeConfig(since, until int, url string) *tg.HelpAppConfig {
	return &tg.HelpAppConfig{Config: &tg.JSONObject{Value: []tg.JSONObjectValue{
		{Key: "freeze_since_date", Value: &tg.JSONNumber{Value: float64(since)}},
		{Key: "freeze_until_date", Value: &tg.JSONNumber{Value: float64(until)}},
		{Key: "freeze_appeal_url", Value: &tg.JSONString{Value: url}},
		{Key: "other", Value: &tg.JSONBool{Value: true}},
	}}}
}

func TestFreezeStateFrom(t *testing.T) {
	s := FreezeStateFrom(freezeConfig(100, 200, "https://t.me/spambot").Config)
	require.True(t, s.Frozen())
	require.Equal(t, time.Unix(100, 0), s.Since)
	require.Equal(t, time.Unix(200, 0), s.Until)
	require.Equal(t, "https://t.me/spambot", s.AppealURL)

	require.False(t, FreezeStateFrom(&tg.JSONObject{}).Frozen())
	require.False(t, FreezeStateFrom(&tg.JSONNull{}).Frozen())
}

func TestClientAuthKeyDrop(t *testing.T) {
	var fail atomic.Pointer[tgerr.Error]
	client := newTestClient(authLostHandler(&fail))
	client.authLost = make(chan error, 1)
	reported := make(chan error, 1)
	client.onAuthLost = func(ctx context.Context, err error) error {
		reported <- err
		return err
	}
	update := &tg.UpdateShort{Update: &tg.UpdateServiceNotification{
		Type:    "AUTH_KEY_DROP_DUPLICATE",
		Message: "Please log in again",
		Media:   &tg.MessageMediaEmpty{},
	}}

	// Not authorized: ignored.
	require.NoError(t, client.processUpdates(update))
	require.Nil(t, client.auth.done.Load())

	client.markAuthorized()
	require.NoError(t, client.processUpdates(update))
	err := <-reported
	require.Equal(t, AuthLostRevoked, AuthLostKindOf(err))
	rpcErr, ok := tgerr.As(err)
	require.True(t, ok)
	require.Equal(t, "AUTH_KEY_DROP_DUPLICATE", rpcErr.Type)
	require.Equal(t, "Please log in again", rpcErr.Message)

	var lost *AuthLostError
	require.ErrorAs(t, <-client.authLost, &lost)
	require.Equal(t, AuthLostRevoked, lost.Kind())

	// Other notifications don't report.
	client.markAuthorized()
	require.NoError(t, client.processUpdates(&tg.UpdateShort{Update: &tg.UpdateServiceNotification{
		Type: "123", Popup: true, Message: "hi", Media: &tg.MessageMediaEmpty{},
	}}))
	require.False(t, client.auth.reported.Load())
}

func TestClientOnFrozen(t *testing.T) {
	ctx := context.Background()
	var (
		mux      sync.Mutex
		appCalls int
	)
	client := newTestClient(func(id int64, body bin.Encoder) (bin.Encoder, error) {
		switch body.(type) {
		case *tg.HelpGetAppConfigRequest:
			mux.Lock()
			appCalls++
			mux.Unlock()
			return freezeConfig(100, 200, "https://t.me/spambot"), nil
		case *tg.MessagesSendMessageRequest:
			return nil, tgerr.New(420, "FROZEN_METHOD_INVALID")
		}
		return &tg.Config{}, nil
	})
	clock := neo.NewTime(time.Unix(1000, 0))
	client.clock = clock
	frozen := make(chan FreezeState, 2)
	client.onFrozen = func(ctx context.Context, state FreezeState) {
		frozen <- state
	}

	send := func() {
		_, err := client.API().MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
			Peer: &tg.InputPeerSelf{}, Message: "hi",
		})
		require.True(t, IsFrozen(err))
	}
	send()
	state := <-frozen
	require.Equal(t, time.Unix(100, 0), state.Since)
	require.Equal(t, "https://t.me/spambot", state.AppealURL)

	// Throttled: no new check within freezeCheckInterval.
	send()
	clock.Travel(freezeCheckInterval - time.Second)
	send()
	select {
	case <-frozen:
		t.Fatal("unexpected freeze check")
	case <-time.After(50 * time.Millisecond):
	}

	clock.Travel(2 * time.Second)
	send()
	<-frozen
	mux.Lock()
	require.Equal(t, 2, appCalls)
	mux.Unlock()
}

func TestClientFreezeState(t *testing.T) {
	client := newTestClient(func(id int64, body bin.Encoder) (bin.Encoder, error) {
		return &tg.HelpAppConfig{Config: &tg.JSONObject{}}, nil
	})
	state, err := client.FreezeState(context.Background())
	require.NoError(t, err)
	require.False(t, state.Frozen())
}
