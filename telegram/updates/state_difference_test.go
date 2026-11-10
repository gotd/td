package updates

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// seqDiffAPI returns diffs in order (the last one repeatedly) and records the
// pts of every getDifference request.
type seqDiffAPI struct {
	mu    sync.Mutex
	diffs []tg.UpdatesDifferenceClass
	reqs  []int

	chDiffs []tg.UpdatesChannelDifferenceClass
	chReqs  []int
}

func (a *seqDiffAPI) UpdatesGetState(ctx context.Context) (*tg.UpdatesState, error) {
	return &tg.UpdatesState{}, nil
}

func (a *seqDiffAPI) UpdatesGetDifference(
	ctx context.Context, r *tg.UpdatesGetDifferenceRequest,
) (tg.UpdatesDifferenceClass, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.reqs = append(a.reqs, r.Pts)
	if len(a.reqs) > 20 {
		// Guard against unbounded recursion in a broken implementation.
		return &tg.UpdatesDifferenceEmpty{Date: 1}, nil
	}
	if len(a.diffs) == 0 {
		return &tg.UpdatesDifferenceEmpty{Date: 1}, nil
	}
	d := a.diffs[0]
	if len(a.diffs) > 1 {
		a.diffs = a.diffs[1:]
	}
	return d, nil
}

func (a *seqDiffAPI) UpdatesGetChannelDifference(
	ctx context.Context, r *tg.UpdatesGetChannelDifferenceRequest,
) (tg.UpdatesChannelDifferenceClass, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.chReqs = append(a.chReqs, r.Pts)
	if len(a.chDiffs) == 0 {
		return &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: r.Pts}, nil
	}
	d := a.chDiffs[0]
	a.chDiffs = a.chDiffs[1:]
	return d, nil
}

// recordTypes returns a handler that records type names of dispatched updates.
func recordTypes(names *[]string) telegram.UpdateHandler {
	return telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
		for _, x := range u.(*tg.Updates).Updates {
			*names = append(*names, x.TypeName())
		}
		return nil
	})
}

func requireNoPending(t *testing.T, s *internalState) {
	t.Helper()
	for _, box := range []*sequenceBox{s.pts, s.qts, s.seq} {
		require.Empty(t, box.pending, "pending updates")
		require.False(t, box.gaps.Has(), "gaps: %v", box.gaps)
	}
}

// UPD-1: other_updates of updates.difference follow new messages that occupy
// pts in between. They must be dispatched, not postponed as a gap and dropped.
func TestDifferenceOtherUpdatesApplied(t *testing.T) {
	ctx := context.Background()
	var got []string
	api := &seqDiffAPI{diffs: []tg.UpdatesDifferenceClass{
		&tg.UpdatesDifference{
			NewMessages: []tg.MessageClass{
				&tg.Message{ID: 1, PeerID: &tg.PeerChat{ChatID: 1}},
				&tg.Message{ID: 2, PeerID: &tg.PeerChat{ChatID: 1}},
			},
			OtherUpdates: []tg.UpdateClass{
				&tg.UpdateReadHistoryInbox{Peer: &tg.PeerChat{ChatID: 1}, MaxID: 2, Pts: 103, PtsCount: 1},
				&tg.UpdateEditMessage{Message: &tg.Message{ID: 1, PeerID: &tg.PeerChat{ChatID: 1}}, Pts: 104, PtsCount: 1},
			},
			State: tg.UpdatesState{Pts: 104, Date: 1},
		},
	}}
	s := newShortTestState(t, api, recordTypes(&got))
	s.pts.SetState(100, "init")

	require.NoError(t, s.getDifference(ctx, "test"))
	require.Equal(t, []string{
		"updateNewMessage",
		"updateNewMessage",
		"updateReadHistoryInbox",
		"updateEditMessage",
	}, got, "new messages first, then other updates")
	require.Equal(t, 104, s.pts.State())
	requireNoPending(t, s)

	// The next live update applies directly, no second difference.
	got = nil
	require.NoError(t, s.handleUpdates(ctx, &tg.UpdateShort{
		Update: &tg.UpdateDeleteMessages{Messages: []int{2}, Pts: 105, PtsCount: 1},
	}))
	require.Equal(t, []string{"updateDeleteMessages"}, got)
	require.Equal(t, 105, s.pts.State())
	require.Equal(t, []int{100}, api.reqs, "exactly one getDifference")
}

// UPD-1, differenceSlice path.
func TestDifferenceSliceOtherUpdatesApplied(t *testing.T) {
	ctx := context.Background()
	var got []string
	api := &seqDiffAPI{diffs: []tg.UpdatesDifferenceClass{
		&tg.UpdatesDifferenceSlice{
			NewMessages: []tg.MessageClass{&tg.Message{ID: 1, PeerID: &tg.PeerChat{ChatID: 1}}},
			OtherUpdates: []tg.UpdateClass{
				&tg.UpdateDeleteMessages{Messages: []int{9}, Pts: 102, PtsCount: 1},
			},
			IntermediateState: tg.UpdatesState{Pts: 102, Date: 1},
		},
		&tg.UpdatesDifferenceEmpty{Date: 2},
	}}
	s := newShortTestState(t, api, recordTypes(&got))
	s.pts.SetState(100, "init")

	require.NoError(t, s.getDifference(ctx, "test"))
	require.Equal(t, []string{"updateNewMessage", "updateDeleteMessages"}, got)
	require.Equal(t, 102, s.pts.State())
	require.Equal(t, []int{100, 102}, api.reqs)
	requireNoPending(t, s)
}

// UPD-1, qts: new encrypted messages consume qts before qts other updates.
func TestDifferenceQtsOtherUpdatesApplied(t *testing.T) {
	ctx := context.Background()
	var got []string
	api := &seqDiffAPI{diffs: []tg.UpdatesDifferenceClass{
		&tg.UpdatesDifference{
			NewEncryptedMessages: []tg.EncryptedMessageClass{
				&tg.EncryptedMessage{RandomID: 1, ChatID: 1},
			},
			OtherUpdates: []tg.UpdateClass{
				&tg.UpdateBotStopped{UserID: 7, Stopped: true, Qts: 12},
			},
			State: tg.UpdatesState{Pts: 100, Qts: 12, Date: 1},
		},
	}}
	s := newShortTestState(t, api, recordTypes(&got))
	s.pts.SetState(100, "init")
	s.qts.SetState(10, "init")

	require.NoError(t, s.getDifference(ctx, "test"))
	require.Equal(t, []string{"updateNewEncryptedMessage", "updateBotStopped"}, got)
	require.Equal(t, 12, s.qts.State())
	require.Equal(t, []int{100}, api.reqs)
	requireNoPending(t, s)
}

// UPD-1: updateMessageID is applied before the new messages it refers to.
func TestDifferencePreludeOrder(t *testing.T) {
	ctx := context.Background()
	var got []string
	api := &seqDiffAPI{diffs: []tg.UpdatesDifferenceClass{
		&tg.UpdatesDifference{
			NewMessages: []tg.MessageClass{&tg.Message{ID: 5, PeerID: &tg.PeerChat{ChatID: 1}}},
			OtherUpdates: []tg.UpdateClass{
				&tg.UpdateDeleteMessages{Messages: []int{5}, Pts: 102, PtsCount: 1},
				&tg.UpdateMessageID{ID: 5, RandomID: 42},
			},
			State: tg.UpdatesState{Pts: 102, Date: 1},
		},
	}}
	s := newShortTestState(t, api, recordTypes(&got))
	s.pts.SetState(100, "init")

	require.NoError(t, s.getDifference(ctx, "test"))
	require.Equal(t, []string{
		"updateMessageID",
		"updateNewMessage",
		"updateDeleteMessages",
	}, got)
}

// UPD-1: pending updates covered by the difference are dropped, not kept as
// stale entries.
func TestDifferenceDropsCoveredPending(t *testing.T) {
	ctx := context.Background()
	var got []string
	api := &seqDiffAPI{diffs: []tg.UpdatesDifferenceClass{
		&tg.UpdatesDifference{
			NewMessages: []tg.MessageClass{
				&tg.Message{ID: 1, PeerID: &tg.PeerChat{ChatID: 1}},
				&tg.Message{ID: 2, PeerID: &tg.PeerChat{ChatID: 1}},
			},
			OtherUpdates: []tg.UpdateClass{
				&tg.UpdateDeleteMessages{Messages: []int{1}, Pts: 103, PtsCount: 1},
			},
			State: tg.UpdatesState{Pts: 103, Date: 1},
		},
	}}
	s := newShortTestState(t, api, recordTypes(&got))
	s.pts.SetState(100, "init")

	// Live update after a gap: postponed.
	require.NoError(t, s.handleUpdates(ctx, &tg.UpdateShort{
		Update: &tg.UpdateDeleteMessages{Messages: []int{1}, Pts: 103, PtsCount: 1},
	}))
	require.Len(t, s.pts.pending, 1)
	require.Empty(t, got)

	// Gap timer fires.
	require.NoError(t, s.getDifference(ctx, "pts-gap-timeout"))
	require.Equal(t, []string{
		"updateNewMessage",
		"updateNewMessage",
		"updateDeleteMessages",
	}, got, "delete dispatched once, from the difference")
	require.Equal(t, 103, s.pts.State())
	requireNoPending(t, s)
}

// UPD-3: an update in a difference referencing a user without a known access
// hash (min user) must be applied, not re-trigger getDifference from the same
// pts forever.
func TestDifferenceUnknownPeerNoRecursion(t *testing.T) {
	ctx := context.Background()
	var got []string
	api := &seqDiffAPI{diffs: []tg.UpdatesDifferenceClass{
		&tg.UpdatesDifference{
			OtherUpdates: []tg.UpdateClass{
				&tg.UpdateEditMessage{
					Message:  &tg.Message{ID: 1, PeerID: &tg.PeerUser{UserID: 555}},
					Pts:      101,
					PtsCount: 1,
				},
			},
			Users: []tg.UserClass{&tg.User{ID: 555, Min: true, AccessHash: 1}},
			State: tg.UpdatesState{Pts: 101, Date: 1},
		},
		&tg.UpdatesDifferenceEmpty{Date: 1},
	}}
	s := newShortTestState(t, api, recordTypes(&got))
	s.pts.SetState(100, "init")

	require.NoError(t, s.getDifference(ctx, "test"))
	require.Equal(t, []int{100}, api.reqs, "exactly one getDifference")
	require.Equal(t, []string{"updateEditMessage"}, got)
	require.Equal(t, 101, s.pts.State())
}

// UPD-3, slice path.
func TestDifferenceSliceUnknownPeerNoRecursion(t *testing.T) {
	ctx := context.Background()
	var got []string
	api := &seqDiffAPI{diffs: []tg.UpdatesDifferenceClass{
		&tg.UpdatesDifferenceSlice{
			OtherUpdates: []tg.UpdateClass{
				&tg.UpdateNewMessage{
					Message:  &tg.Message{ID: 1, PeerID: &tg.PeerUser{UserID: 555}},
					Pts:      101,
					PtsCount: 1,
				},
			},
			IntermediateState: tg.UpdatesState{Pts: 101, Date: 1},
		},
		&tg.UpdatesDifferenceEmpty{Date: 1},
	}}
	s := newShortTestState(t, api, recordTypes(&got))
	s.pts.SetState(100, "init")

	require.NoError(t, s.getDifference(ctx, "test"))
	require.Equal(t, []int{100, 101}, api.reqs)
	require.Equal(t, []string{"updateNewMessage"}, got)
}

// newTestChannel creates a channel worker of s without starting it.
func newTestChannel(s *internalState, channelID int64, pts int) *channelState {
	return s.newChannelState(channelID, 99, pts)
}

// UPD-2: other_updates of updates.channelDifference must be dispatched, not
// sent to *internalState and dropped after pts was set to diff.Pts.
func TestChannelDifferenceOtherUpdatesApplied(t *testing.T) {
	ctx := context.Background()
	var got []string
	peer := &tg.PeerChannel{ChannelID: 77}
	api := &seqDiffAPI{chDiffs: []tg.UpdatesChannelDifferenceClass{
		&tg.UpdatesChannelDifference{
			Final:       true,
			Pts:         5,
			NewMessages: []tg.MessageClass{&tg.Message{ID: 4, PeerID: peer}},
			OtherUpdates: []tg.UpdateClass{
				&tg.UpdateDeleteChannelMessages{ChannelID: 77, Messages: []int{1}, Pts: 5, PtsCount: 1},
				&tg.UpdateMessageID{ID: 4, RandomID: 42},
			},
		},
	}}
	s := newShortTestState(t, api, recordTypes(&got))
	ch := newTestChannel(s, 77, 3)

	require.NoError(t, ch.getDifference(ctx, "test"))
	require.Equal(t, []string{
		"updateMessageID",
		"updateNewChannelMessage",
		"updateDeleteChannelMessages",
	}, got)
	require.Equal(t, 5, ch.pts.State())
	require.Empty(t, ch.pts.pending)
	require.Empty(t, s.internalQueue, "nothing is sent to internalState")
	require.Equal(t, []int{3}, api.chReqs)
}

// UPD-2, non-final difference: other updates of every slice are dispatched.
func TestChannelDifferenceNotFinalOtherUpdatesApplied(t *testing.T) {
	ctx := context.Background()
	var got []string
	peer := &tg.PeerChannel{ChannelID: 77}
	api := &seqDiffAPI{chDiffs: []tg.UpdatesChannelDifferenceClass{
		&tg.UpdatesChannelDifference{
			Pts:         5,
			NewMessages: []tg.MessageClass{&tg.Message{ID: 4, PeerID: peer}},
			OtherUpdates: []tg.UpdateClass{
				&tg.UpdateDeleteChannelMessages{ChannelID: 77, Messages: []int{1}, Pts: 5, PtsCount: 1},
			},
		},
		&tg.UpdatesChannelDifference{
			Final: true,
			Pts:   7,
			OtherUpdates: []tg.UpdateClass{
				&tg.UpdateEditChannelMessage{Message: &tg.Message{ID: 4, PeerID: peer}, Pts: 6, PtsCount: 1},
				&tg.UpdatePinnedChannelMessages{ChannelID: 77, Messages: []int{4}, Pinned: true, Pts: 7, PtsCount: 1},
			},
		},
	}}
	s := newShortTestState(t, api, recordTypes(&got))
	ch := newTestChannel(s, 77, 3)

	require.NoError(t, ch.getDifference(ctx, "test"))
	require.Equal(t, []string{
		"updateNewChannelMessage",
		"updateDeleteChannelMessages",
		"updateEditChannelMessage",
		"updatePinnedChannelMessages",
	}, got)
	require.Equal(t, 7, ch.pts.State())
	require.Equal(t, []int{3, 5}, api.chReqs)
}

// UPD-2: pending channel updates covered by the difference are dropped, and
// updates of other sequences are still handed to internalState.
func TestChannelDifferencePendingAndForeign(t *testing.T) {
	ctx := context.Background()
	var got []string
	peer := &tg.PeerChannel{ChannelID: 77}
	api := &seqDiffAPI{chDiffs: []tg.UpdatesChannelDifferenceClass{
		&tg.UpdatesChannelDifference{
			Final:       true,
			Pts:         5,
			NewMessages: []tg.MessageClass{&tg.Message{ID: 4, PeerID: peer}},
			OtherUpdates: []tg.UpdateClass{
				&tg.UpdateDeleteChannelMessages{ChannelID: 77, Messages: []int{1}, Pts: 5, PtsCount: 1},
				&tg.UpdateDeleteChannelMessages{ChannelID: 88, Messages: []int{1}, Pts: 9, PtsCount: 1},
			},
		},
	}}
	s := newShortTestState(t, api, recordTypes(&got))
	ch := newTestChannel(s, 77, 3)

	// Live update after a gap: postponed.
	require.NoError(t, ch.handleUpdate(ctx,
		&tg.UpdateDeleteChannelMessages{ChannelID: 77, Messages: []int{1}, Pts: 5, PtsCount: 1},
		entities{},
	))
	require.Len(t, ch.pts.pending, 1)

	require.NoError(t, ch.getDifference(ctx, "channel-pts-gap-timeout"))
	require.Equal(t, []string{
		"updateNewChannelMessage",
		"updateDeleteChannelMessages",
	}, got, "delete dispatched once, from the difference")
	require.Empty(t, ch.pts.pending)
	require.False(t, ch.pts.gaps.Has())

	require.Len(t, s.internalQueue, 1, "foreign update is handed to internalState")
	out := (<-s.internalQueue).update.(*tg.Updates)
	require.Len(t, out.Updates, 1)
	require.Equal(t, int64(88), out.Updates[0].(*tg.UpdateDeleteChannelMessages).ChannelID)
}

// A cancelled context during a common difference must not advance the state.
func TestDifferenceCancelledKeepsState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var got []string
	api := &seqDiffAPI{diffs: []tg.UpdatesDifferenceClass{
		&tg.UpdatesDifference{
			NewMessages: []tg.MessageClass{&tg.Message{ID: 1, PeerID: &tg.PeerChat{ChatID: 1}}},
			State:       tg.UpdatesState{Pts: 101, Date: 1},
		},
	}}
	s := newShortTestState(t, api, recordTypes(&got))
	s.pts.SetState(100, "init")

	err := s.getDifference(ctx, "test")
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 100, s.pts.State())
	st, found, err := s.storage.GetState(context.Background(), s.selfID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, 0, st.Pts, "state is not saved")
}

// A cancelled context during a channel difference must not advance the pts.
func TestChannelDifferenceCancelledKeepsPts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var got []string
	api := &seqDiffAPI{chDiffs: []tg.UpdatesChannelDifferenceClass{
		&tg.UpdatesChannelDifference{
			Final:       true,
			Pts:         5,
			NewMessages: []tg.MessageClass{&tg.Message{ID: 4, PeerID: &tg.PeerChannel{ChannelID: 77}}},
		},
	}}
	s := newShortTestState(t, api, recordTypes(&got))
	require.NoError(t, s.storage.SetChannelPts(context.Background(), s.selfID, 77, 3))
	ch := newTestChannel(s, 77, 3)

	err := ch.getDifference(ctx, "test")
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 3, ch.pts.State())
	pts, found, err := s.storage.GetChannelPts(context.Background(), s.selfID, 77)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, 3, pts, "pts is not saved")
}

// Channel updates inside a common difference are routed to the channel worker,
// not dispatched by internalState.
func TestDifferenceRoutesChannelUpdates(t *testing.T) {
	ctx := context.Background()
	var got []string
	edit := &tg.UpdateEditChannelMessage{
		Message: &tg.Message{ID: 4, PeerID: &tg.PeerChannel{ChannelID: 77}}, Pts: 6, PtsCount: 1,
	}
	api := &seqDiffAPI{diffs: []tg.UpdatesDifferenceClass{
		&tg.UpdatesDifference{
			OtherUpdates: []tg.UpdateClass{
				edit,
				&tg.UpdateDeleteMessages{Messages: []int{1}, Pts: 101, PtsCount: 1},
			},
			Chats: []tg.ChatClass{&tg.Channel{ID: 77, AccessHash: 99}},
			State: tg.UpdatesState{Pts: 101, Date: 1},
		},
	}}
	s := newShortTestState(t, api, recordTypes(&got))
	s.pts.SetState(100, "init")
	ch := newTestChannel(s, 77, 5)
	s.channels[77] = ch

	require.NoError(t, s.getDifference(ctx, "test"))
	require.Equal(t, []string{"updateDeleteMessages"}, got)
	require.Len(t, ch.updates, 1)
	cu := <-ch.updates
	require.Same(t, edit, cu.update)
	require.NotEmpty(t, cu.entities.Chats, "entities are passed")
}

// updatePtsChanged in a difference makes exactly one follow-up getDifference
// instead of trusting the difference state, and cannot loop.
func TestDifferencePtsChangedRefetchesOnce(t *testing.T) {
	ctx := context.Background()
	var got []string
	api := &seqDiffAPI{diffs: []tg.UpdatesDifferenceClass{
		&tg.UpdatesDifference{
			OtherUpdates: []tg.UpdateClass{
				&tg.UpdatePtsChanged{},
				&tg.UpdateDeleteMessages{Messages: []int{1}, Pts: 101, PtsCount: 1},
			},
			State: tg.UpdatesState{Pts: 101, Date: 1},
		},
		&tg.UpdatesDifference{
			OtherUpdates: []tg.UpdateClass{&tg.UpdatePtsChanged{}},
			State:        tg.UpdatesState{Pts: 50, Date: 2},
		},
		&tg.UpdatesDifferenceEmpty{Date: 3},
	}}
	s := newShortTestState(t, api, recordTypes(&got))
	s.pts.SetState(100, "init")

	require.NoError(t, s.getDifference(ctx, "test"))
	require.Equal(t, []int{100, 101}, api.reqs, "one follow-up only")
	require.Equal(t, []string{"updateDeleteMessages"}, got, "updatePtsChanged is not dispatched")
	require.Equal(t, 50, s.pts.State())
}

// A pending update continuing the difference state is applied right away.
func TestDifferenceAppliesContiguousPending(t *testing.T) {
	ctx := context.Background()
	var got []string
	api := &seqDiffAPI{diffs: []tg.UpdatesDifferenceClass{
		&tg.UpdatesDifference{
			NewMessages: []tg.MessageClass{
				&tg.Message{ID: 1, PeerID: &tg.PeerChat{ChatID: 1}},
				&tg.Message{ID: 2, PeerID: &tg.PeerChat{ChatID: 1}},
				&tg.Message{ID: 3, PeerID: &tg.PeerChat{ChatID: 1}},
			},
			State: tg.UpdatesState{Pts: 103, Date: 1},
		},
	}}
	s := newShortTestState(t, api, recordTypes(&got))
	s.pts.SetState(100, "init")

	require.NoError(t, s.handleUpdates(ctx, &tg.UpdateShort{
		Update: &tg.UpdateDeleteMessages{Messages: []int{1}, Pts: 104, PtsCount: 1},
	}))
	require.Len(t, s.pts.pending, 1)

	require.NoError(t, s.getDifference(ctx, "pts-gap-timeout"))
	require.Equal(t, []string{
		"updateNewMessage",
		"updateNewMessage",
		"updateNewMessage",
		"updateDeleteMessages",
	}, got)
	require.Equal(t, 104, s.pts.State())
	requireNoPending(t, s)
}

// A pending update still separated from the difference state by a gap keeps
// waiting, with the gap re-armed.
func TestDifferenceRearmsRemainingGap(t *testing.T) {
	ctx := context.Background()
	var got []string
	api := &seqDiffAPI{diffs: []tg.UpdatesDifferenceClass{
		&tg.UpdatesDifference{
			NewMessages: []tg.MessageClass{&tg.Message{ID: 1, PeerID: &tg.PeerChat{ChatID: 1}}},
			State:       tg.UpdatesState{Pts: 101, Date: 1},
		},
	}}
	s := newShortTestState(t, api, recordTypes(&got))
	s.pts.SetState(100, "init")

	require.NoError(t, s.handleUpdates(ctx, &tg.UpdateShort{
		Update: &tg.UpdateDeleteMessages{Messages: []int{1}, Pts: 105, PtsCount: 1},
	}))
	require.NoError(t, s.getDifference(ctx, "pts-gap-timeout"))
	require.Equal(t, []string{"updateNewMessage"}, got)
	require.Equal(t, 101, s.pts.State())
	require.Len(t, s.pts.pending, 1)
	require.Equal(t, "[[101,104)]", s.pts.gaps.String())

	// The gap is resolved by waiting for the missing updates.
	require.NoError(t, s.handleUpdates(ctx, &tg.UpdateShort{
		Update: &tg.UpdateDeleteMessages{Messages: []int{2}, Pts: 104, PtsCount: 3},
	}))
	require.Equal(t, 105, s.pts.State())
	requireNoPending(t, s)
}
