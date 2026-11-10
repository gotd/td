package updates

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// TestCommonStateGapKeepsBufferedUpdate reproduces a common-state (private chat
// / basic group) pts gap that strands a buffered update (ported from #1854).
//
// A live edit at pts 12 arrives while the state is 10, so it is buffered behind
// the gap at 11. A gap-triggered getDifference then replays the same edit in
// OtherUpdates and advances the state to 12. The replayed edit must be
// dispatched exactly once, and the buffered copy dropped as covered.
func TestCommonStateGapKeepsBufferedUpdate(t *testing.T) {
	ctx := context.Background()

	msg := &tg.Message{
		ID:       1,
		PeerID:   &tg.PeerUser{UserID: 555},
		FromID:   &tg.PeerUser{UserID: 555},
		Message:  "v1",
		Date:     1,
		EditDate: 1,
	}

	// getDifference replays the edit and advances the state to 12.
	api := &oneDiffAPI{diff: &tg.UpdatesDifference{
		OtherUpdates: []tg.UpdateClass{
			&tg.UpdateEditMessage{Message: msg, Pts: 12, PtsCount: 1},
		},
		State: tg.UpdatesState{Pts: 12, Date: 1},
	}}

	var dispatched []int
	handler := telegram.UpdateHandlerFunc(func(_ context.Context, u tg.UpdatesClass) error {
		if up, ok := u.(*tg.Updates); ok {
			for _, x := range up.Updates {
				if e, ok := x.(*tg.UpdateEditMessage); ok {
					dispatched = append(dispatched, e.Pts)
				}
			}
		}
		return nil
	})

	s := newShortTestState(t, api, handler)
	s.pts.SetState(10, "init")
	s.date = 1

	// Mark the sender known so the edit is not diverted to getDifference by the
	// peer-access-hash check.
	require.NoError(t, s.handleUpdates(ctx, &tg.Updates{
		Users: []tg.UserClass{&tg.User{ID: 555, AccessHash: 7777}},
	}))

	// Live edit at pts 12 while the state is 10: gap at 11, so it is buffered.
	require.NoError(t, s.handleUpdates(ctx, &tg.Updates{Updates: []tg.UpdateClass{
		&tg.UpdateEditMessage{Message: msg, Pts: 12, PtsCount: 1},
	}}))
	require.Empty(t, dispatched, "the edit is buffered behind the gap, not dispatched yet")

	// A gap-triggered difference replays the edit and advances the state to 12.
	require.NoError(t, s.getDifference(ctx, "test"))

	require.Equal(t, []int{12}, dispatched, "the replayed edit at pts 12 is dispatched once")
	require.Equal(t, 12, s.pts.State())
	requireNoPending(t, s)
}

type channelDifferenceAPI struct {
	countDiffAPI
	diff tg.UpdatesChannelDifferenceClass
}

func (a *channelDifferenceAPI) UpdatesGetChannelDifference(
	context.Context, *tg.UpdatesGetChannelDifferenceRequest,
) (tg.UpdatesChannelDifferenceClass, error) {
	return a.diff, nil
}

// TestChannelDifferenceRecoveredFacts checks that edit/delete/pin facts of a
// channel difference are delivered after new messages, before the pts
// checkpoint, whether they carry a zero or a covered pts (ported from #1867).
func TestChannelDifferenceRecoveredFacts(t *testing.T) {
	for _, tt := range []struct {
		name string
		pts  int
		fail bool
	}{
		{name: "zero PTS"},
		{name: "covered PTS", pts: 15},
		{name: "handler error", pts: 15, fail: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				message := &tg.Message{ID: 1, PeerID: &tg.PeerChannel{ChannelID: 42}}
				facts := []tg.UpdateClass{
					&tg.UpdateEditChannelMessage{Message: message, Pts: tt.pts},
					&tg.UpdateDeleteChannelMessages{ChannelID: 42, Messages: []int{2}, Pts: tt.pts},
					&tg.UpdatePinnedChannelMessages{ChannelID: 42, Messages: []int{1}, Pts: tt.pts},
				}
				channel := &tg.Channel{ID: 99}
				channel.SetAccessHash(999)
				api := &channelDifferenceAPI{diff: &tg.UpdatesChannelDifference{
					Final:        true,
					Pts:          20,
					NewMessages:  []tg.MessageClass{message},
					OtherUpdates: facts,
					Users:        []tg.UserClass{&tg.User{ID: 7, AccessHash: 777}},
					Chats:        []tg.ChatClass{channel},
				}}
				h := &recordingHandler{}
				var ch *channelState
				var checkpoints []int
				s := newShortTestState(t, api, telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
					checkpoint, _, err := ch.storage.GetChannelPts(ctx, ch.selfID, 42)
					if err != nil {
						return err
					}
					checkpoints = append(checkpoints, checkpoint)
					err = h.Handle(ctx, u)
					if tt.fail {
						return errors.New("offline handler failure")
					}
					return err
				}))
				ch = s.newChannelState(42, 99, 10)
				s.channels[42] = ch
				require.NoError(t, s.storage.SetChannelPts(ctx, s.selfID, 42, 10))
				done := make(chan error, 2)
				go func() { done <- s.Run(ctx) }()
				go func() { done <- ch.Run(ctx) }()
				synctest.Wait()

				got := h.updates()
				require.Len(t, got, 4)
				require.IsType(t, &tg.UpdateNewChannelMessage{}, got[0])
				require.Equal(t, facts, got[1:], "recovered facts must follow new messages, without a live PTS round trip")
				// One batch: new messages and facts are delivered together.
				require.Equal(t, []int{10}, checkpoints, "deliver before checkpointing")
				stored, found, err := s.storage.GetChannelPts(ctx, s.selfID, 42)
				require.NoError(t, err)
				require.True(t, found)
				require.Equal(t, 20, stored)

				userHash, found, err := s.userHasher.GetUserAccessHash(ctx, s.selfID, 7)
				require.NoError(t, err)
				require.True(t, found, "the common worker must still learn response entities")
				require.EqualValues(t, 777, userHash)
				channelHash, found, err := s.hasher.GetChannelAccessHash(ctx, s.selfID, 99)
				require.NoError(t, err)
				require.True(t, found)
				require.EqualValues(t, 999, channelHash)

				live := &tg.UpdateDeleteChannelMessages{ChannelID: 42, Messages: []int{3}, Pts: 21, PtsCount: 1}
				require.NoError(t, ch.Push(ctx, channelUpdate{update: live}))
				synctest.Wait()
				require.Equal(t, append(got, live), h.updates(), "normal live sequencing must continue without replay duplicates")
				cancel()
				require.ErrorIs(t, <-done, context.Canceled)
				require.ErrorIs(t, <-done, context.Canceled)
			})
		})
	}
}

// TestChannelDifferencePreservesOtherRoutes checks that updates of other
// sequences in a channel difference are still handed to internalState (ported
// from #1867). Unlike #1867, updateMessageID is applied by the channel worker
// (as TDLib does) and updateChannelTooLong of the same channel is ignored: the
// difference is the recovery itself.
func TestChannelDifferencePreservesOtherRoutes(t *testing.T) {
	ctx := context.Background()
	forwarded := []tg.UpdateClass{
		&tg.UpdateDeleteMessages{Messages: []int{3}, Pts: 11, PtsCount: 1},
		&tg.UpdateDeleteChannelMessages{ChannelID: 99, Messages: []int{4}, Pts: 12, PtsCount: 1},
		&tg.UpdatePtsChanged{},
	}
	messageID := &tg.UpdateMessageID{ID: 1, RandomID: 2}
	local := &tg.UpdatePinnedChannelMessages{ChannelID: 42, Messages: []int{1}}
	others := []tg.UpdateClass{local, messageID, &tg.UpdateChannelTooLong{ChannelID: 42}}
	api := &channelDifferenceAPI{diff: &tg.UpdatesChannelDifference{
		Final: true, Pts: 20, OtherUpdates: append(others, forwarded...),
	}}
	h := &recordingHandler{}
	s := newShortTestState(t, api, h)
	ch := s.newChannelState(42, 99, 10)
	require.NoError(t, ch.getDifference(ctx, "test"))
	require.Equal(t, []tg.UpdateClass{messageID, local}, h.updates())
	require.Equal(t, forwarded, (<-s.internalQueue).update.(*tg.Updates).Updates)
}
