package updates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
)

func channelMessageUpdate(channelID int64) *tg.Updates {
	return &tg.Updates{Updates: []tg.UpdateClass{&tg.UpdateNewChannelMessage{
		Message:  &tg.Message{ID: 5, PeerID: &tg.PeerChannel{ChannelID: channelID}},
		Pts:      5,
		PtsCount: 1,
	}}}
}

// A channel whose access hash the hasher already knows (e.g. *peers.Manager
// that saw it in dialogs) is tracked on its first update without the
// updates.getDifference that looks for the hash (tdesktop-vs-gotd item 12).
func TestFirstChannelUpdateKnownHashNoDifference(t *testing.T) {
	ctx := context.Background()
	api := &countDiffAPI{}
	s := newShortTestState(t, api, nopHandlerForHashTest())
	require.NoError(t, s.hasher.SetChannelAccessHash(ctx, s.selfID, 42, 99))

	require.NoError(t, s.handleUpdates(ctx, channelMessageUpdate(42)))
	require.Equal(t, 0, api.diffCalls, "no getDifference for a known hash")
	state, ok := s.channels[42]
	require.True(t, ok, "channel must be tracked")
	require.EqualValues(t, 99, state.accessHash)
}

// Without a known hash and without the channel in the update's entities, the
// manager still asks getDifference for the hash.
func TestFirstChannelUpdateUnknownHashDifference(t *testing.T) {
	ctx := context.Background()
	api := &countDiffAPI{}
	s := newShortTestState(t, api, nopHandlerForHashTest())

	require.NoError(t, s.handleUpdates(ctx, channelMessageUpdate(42)))
	require.Equal(t, 1, api.diffCalls)
	_, ok := s.channels[42]
	require.False(t, ok, "no hash: update ignored")
}

func nopHandlerForHashTest() telegram.UpdateHandler {
	return telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error { return nil })
}

// *peers.Manager can be passed as Config.AccessHasher and serves users too.
var (
	_ ChannelAccessHasher = (*peers.Manager)(nil)
	_ UserAccessHasher    = (*peers.Manager)(nil)
)
