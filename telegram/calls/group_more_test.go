package calls

import (
	"context"
	"testing"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func TestGroupCallSettersAndWriteAudio(t *testing.T) {
	gc := NewGroupCall(tg.NewClient(tgmock.NewRequire(t)), Options{})

	// Setters create the transport lazily without opening it.
	gc.OnConnected(func() {})
	gc.OnDisconnected(func() {})
	gc.OnTrack(func(*webrtc.TrackRemote, *webrtc.RTPReceiver) {})
	gc.OnParticipants(func([]tg.GroupCallParticipant) {})

	require.Nil(t, gc.AudioTrack(), "no track before join")
	require.Zero(t, gc.AudioSSRC())

	err := gc.WriteAudio(&rtp.Packet{})
	require.Error(t, err, "WriteAudio must fail before join")
}

func TestGroupCallRegisterParticipants(t *testing.T) {
	ctx := context.Background()
	d := tg.NewUpdateDispatcher()
	gc := NewGroupCall(tg.NewClient(tgmock.NewRequire(t)), Options{})

	var got []tg.GroupCallParticipant
	gc.OnParticipants(func(p []tg.GroupCallParticipant) { got = p })
	gc.Register(d)

	// No active call yet: update is ignored.
	require.NoError(t, d.Handle(ctx, &tg.Updates{Updates: []tg.UpdateClass{
		&tg.UpdateGroupCallParticipants{Call: &tg.InputGroupCall{ID: 1}},
	}}))
	require.Nil(t, got)

	// With a matching active call, participants are delivered.
	gc.mu.Lock()
	gc.call = &tg.InputGroupCall{ID: 1, AccessHash: 2}
	gc.mu.Unlock()
	require.NoError(t, d.Handle(ctx, &tg.Updates{Updates: []tg.UpdateClass{
		&tg.UpdateGroupCallParticipants{
			Call:         &tg.InputGroupCall{ID: 1},
			Participants: []tg.GroupCallParticipant{{Source: 42}},
		},
	}}))
	require.Len(t, got, 1)
	require.Equal(t, 42, got[0].Source)
}

func TestGroupCallLeaveNotJoined(t *testing.T) {
	ctx := context.Background()
	gc := NewGroupCall(tg.NewClient(tgmock.NewRequire(t)), Options{})
	gc.connOnce() // create transport without opening
	// Not joined: no PhoneLeaveGroupCall RPC is issued.
	require.NoError(t, gc.Leave(ctx))
}
