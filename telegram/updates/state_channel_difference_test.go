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

type channelDifferenceAPI struct {
	countDiffAPI
	diff tg.UpdatesChannelDifferenceClass
}

func (a *channelDifferenceAPI) UpdatesGetChannelDifference(context.Context, *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
	return a.diff, nil
}

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
					Final: true, Pts: 20, NewMessages: []tg.MessageClass{message}, OtherUpdates: facts,
					Users: []tg.UserClass{&tg.User{ID: 7, AccessHash: 777}}, Chats: []tg.ChatClass{channel},
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
					if tt.fail && u.(*tg.Updates).Updates[0] == facts[0] {
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
				require.Equal(t, []int{10, 10}, checkpoints, "deliver both batches before checkpointing")
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

func TestChannelDifferencePreservesOtherRoutes(t *testing.T) {
	ctx := context.Background()
	forwarded := []tg.UpdateClass{
		&tg.UpdateMessageID{ID: 1, RandomID: 2},
		&tg.UpdateDeleteMessages{Messages: []int{3}, Pts: 11, PtsCount: 1},
		&tg.UpdateDeleteChannelMessages{ChannelID: 99, Messages: []int{4}, Pts: 12, PtsCount: 1},
		&tg.UpdateChannelTooLong{ChannelID: 42}, &tg.UpdatePtsChanged{},
	}
	local := &tg.UpdatePinnedChannelMessages{ChannelID: 42, Messages: []int{1}}
	api := &channelDifferenceAPI{diff: &tg.UpdatesChannelDifference{
		Final: true, Pts: 20, OtherUpdates: append([]tg.UpdateClass{local}, forwarded...),
	}}
	h := &recordingHandler{}
	s := newShortTestState(t, api, h)
	ch := s.newChannelState(42, 99, 10)
	require.NoError(t, ch.getDifference(ctx, "test"))
	require.Equal(t, []tg.UpdateClass{local}, h.updates())
	require.Equal(t, forwarded, (<-s.internalQueue).update.(*tg.Updates).Updates)
}
