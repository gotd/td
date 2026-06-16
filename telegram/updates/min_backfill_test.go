package updates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// TestDispatchBackfillsMinAccessHash is the regression test for
// https://github.com/gotd/td/issues/1553: entities delivered to user handlers
// must carry a usable (full) access hash for min channels/users whose full hash
// the manager already knows, so e.Channels[id].AsInputPeer() works in any chat.
func TestDispatchBackfillsMinAccessHash(t *testing.T) {
	ctx := context.Background()

	const (
		minChannelID  = int64(100) // seen as min, full hash known
		fullChannelID = int64(200) // known full hash
		minHash       = int64(11)  // min access hash carried by the update
		realHash      = int64(99)  // full access hash from the store
		unknownChanID = int64(300) // min, but no stored hash

		minUserID  = int64(400)
		userMin    = int64(44)
		userReal   = int64(88)
		unknownUID = int64(500)
	)

	var got *tg.Updates
	handler := telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
		got = u.(*tg.Updates)
		return nil
	})
	s := newShortTestState(t, &countDiffAPI{}, handler)

	// Pre-seed full hashes, as a prior non-min observation would.
	require.NoError(t, s.hasher.SetChannelAccessHash(ctx, s.selfID, minChannelID, realHash))
	require.NoError(t, s.userHasher.SetUserAccessHash(ctx, s.selfID, minUserID, userReal))

	minChannel := &tg.Channel{ID: minChannelID}
	minChannel.SetAccessHash(minHash)
	minChannel.SetMin(true)

	fullChannel := &tg.Channel{ID: fullChannelID}
	fullChannel.SetAccessHash(realHash) // already full, not min

	unknownMin := &tg.Channel{ID: unknownChanID}
	unknownMin.SetAccessHash(minHash)
	unknownMin.SetMin(true)

	minUser := &tg.User{ID: minUserID, Min: true, AccessHash: userMin}
	unknownUser := &tg.User{ID: unknownUID, Min: true, AccessHash: userMin}

	require.NoError(t, s.dispatch(ctx, &tg.Updates{
		Chats: []tg.ChatClass{minChannel, fullChannel, unknownMin},
		Users: []tg.UserClass{minUser, unknownUser},
	}))
	require.NotNil(t, got)

	channelByID := func(id int64) *tg.Channel {
		for _, c := range got.Chats {
			if ch, ok := c.(*tg.Channel); ok && ch.ID == id {
				return ch
			}
		}
		t.Fatalf("channel %d not delivered", id)
		return nil
	}
	userByID := func(id int64) *tg.User {
		for _, u := range got.Users {
			if usr, ok := u.(*tg.User); ok && usr.ID == id {
				return usr
			}
		}
		t.Fatalf("user %d not delivered", id)
		return nil
	}

	// Min channel with a known full hash: hash backfilled, Min flag retained.
	delivered := channelByID(minChannelID)
	require.Equal(t, realHash, delivered.AccessHash, "min channel hash must be backfilled")
	require.True(t, delivered.Min, "Min flag must be retained")

	// Full channel: untouched.
	require.Equal(t, realHash, channelByID(fullChannelID).AccessHash)

	// Min channel with no stored hash: left unchanged, no regression.
	require.Equal(t, minHash, channelByID(unknownChanID).AccessHash)

	// Min user with a known full hash: backfilled. Unknown: unchanged.
	require.Equal(t, userReal, userByID(minUserID).AccessHash)
	require.Equal(t, userMin, userByID(unknownUID).AccessHash)

	// The original entity must not be mutated (copy-on-write).
	require.Equal(t, minHash, minChannel.AccessHash, "original min channel must be untouched")
	require.Equal(t, userMin, minUser.AccessHash, "original min user must be untouched")
}
