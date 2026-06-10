package updates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// A realtime incoming message can arrive as a full *tg.UpdateNewMessage inside
// *tg.Updates / *tg.UpdatesCombined / *tg.UpdateShort whose sender is absent (or
// only min) in the envelope users[]. Like the short forms, such an update must
// force getDifference so the sender access hash is recovered before dispatch.
// Mirrors TDLib is_acceptable_update for the non-channel get_difference branch.

func newMessageUpdate(senderID int64) *tg.UpdateNewMessage {
	return &tg.UpdateNewMessage{
		Message: &tg.Message{
			ID:      1,
			PeerID:  &tg.PeerUser{UserID: senderID},
			FromID:  &tg.PeerUser{UserID: senderID},
			Message: "hi",
		},
		Pts:      1,
		PtsCount: 1,
	}
}

func TestNewMessageUnknownSenderForcesDifference(t *testing.T) {
	ctx := context.Background()
	api := &countDiffAPI{}
	var dispatched int
	handler := telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
		dispatched++
		return nil
	})
	s := newShortTestState(t, api, handler)

	err := s.handleUpdates(ctx, &tg.Updates{
		Updates: []tg.UpdateClass{newMessageUpdate(555)},
		// sender NOT supplied inline
	})
	require.NoError(t, err)
	require.Equal(t, 1, api.diffCalls, "full new message from unknown sender must force getDifference")
	require.Equal(t, 0, dispatched, "message must not be dispatched directly")
}

func TestUpdateShortWrappedNewMessageUnknownSenderForcesDifference(t *testing.T) {
	ctx := context.Background()
	api := &countDiffAPI{}
	var dispatched int
	handler := telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
		dispatched++
		return nil
	})
	s := newShortTestState(t, api, handler)

	// The exact reproduced shape: updateShort wrapping a pts updateNewMessage,
	// no inline users[].
	err := s.handleUpdates(ctx, &tg.UpdateShort{Update: newMessageUpdate(555)})
	require.NoError(t, err)
	require.Equal(t, 1, api.diffCalls, "updateShort-wrapped new message from unknown sender must force getDifference")
	require.Equal(t, 0, dispatched, "message must not be dispatched directly")
}

func TestNewMessageKnownSenderAppliesDirectly(t *testing.T) {
	ctx := context.Background()
	api := &countDiffAPI{}
	var dispatched int
	handler := telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
		dispatched++
		return nil
	})
	s := newShortTestState(t, api, handler)

	// Sender supplied full (non-min, non-zero hash) inline: it becomes known
	// via saveUserHashes before the acceptability check, so the message applies
	// directly without a getDifference.
	err := s.handleUpdates(ctx, &tg.Updates{
		Updates: []tg.UpdateClass{newMessageUpdate(555)},
		Users:   []tg.UserClass{&tg.User{ID: 555, AccessHash: 7777}},
	})
	require.NoError(t, err)
	require.Equal(t, 0, api.diffCalls, "no difference when the sender is supplied full inline")
	require.Equal(t, 1, dispatched, "message with known sender must be dispatched directly")
}

// TestMessageUpdatesPeersKnownSkipsChannelMessages locks the performance
// boundary: a channel (megagroup) message with a min/unknown sender is excluded
// from the non-channel acceptability check, so it never forces a global
// getDifference. Min senders are routine in megagroups; the channel-difference
// machinery owns that recovery path. (Tested at the function level to isolate it
// from gotd's pre-existing unknown-channel access-hash recovery.)
func TestMessageUpdatesPeersKnownSkipsChannelMessages(t *testing.T) {
	handler := telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
		return nil
	})
	s := newShortTestState(t, &countDiffAPI{}, handler)

	known := s.messageUpdatesPeersKnown(context.Background(), []tg.UpdateClass{&tg.UpdateNewChannelMessage{
		Message: &tg.Message{
			ID:      1,
			PeerID:  &tg.PeerChannel{ChannelID: 42},
			FromID:  &tg.PeerUser{UserID: 555},
			Message: "hi",
		},
		Pts:      1,
		PtsCount: 1,
	}})
	require.True(t, known, "channel messages are excluded from the global-getDifference acceptability check")
}
