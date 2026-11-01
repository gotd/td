package updates

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// Inject failures at the public storage boundary, leaving the synchronized
// in-memory implementation responsible for all actual state transitions.
type applyFaultStore struct {
	StateStorage
	operation string
	failure   error
	armed     atomic.Bool
	hits      atomic.Int32
	writes    atomic.Int32
}

func (s *applyFaultStore) check(operation string) error {
	if !s.armed.Load() {
		return nil
	}
	if operation != "GetChannelPts" {
		s.writes.Add(1)
	}
	if operation == s.operation {
		s.hits.Add(1)
		return s.failure
	}
	return nil
}

func (s *applyFaultStore) SetState(ctx context.Context, id int64, v State) error {
	if err := s.check("SetState"); err != nil {
		return err
	}
	return s.StateStorage.SetState(ctx, id, v)
}
func (s *applyFaultStore) SetPts(ctx context.Context, id int64, v int) error {
	if err := s.check("SetPts"); err != nil {
		return err
	}
	return s.StateStorage.SetPts(ctx, id, v)
}
func (s *applyFaultStore) SetQts(ctx context.Context, id int64, v int) error {
	if err := s.check("SetQts"); err != nil {
		return err
	}
	return s.StateStorage.SetQts(ctx, id, v)
}
func (s *applyFaultStore) SetSeq(ctx context.Context, id int64, v int) error {
	if err := s.check("SetSeq"); err != nil {
		return err
	}
	return s.StateStorage.SetSeq(ctx, id, v)
}
func (s *applyFaultStore) SetDate(ctx context.Context, id int64, v int) error {
	if err := s.check("SetDate"); err != nil {
		return err
	}
	return s.StateStorage.SetDate(ctx, id, v)
}
func (s *applyFaultStore) SetDateSeq(ctx context.Context, id int64, date, seq int) error {
	if err := s.check("SetDateSeq"); err != nil {
		return err
	}
	return s.StateStorage.SetDateSeq(ctx, id, date, seq)
}
func (s *applyFaultStore) SetChannelPts(ctx context.Context, id, channel int64, v int) error {
	if err := s.check("SetChannelPts"); err != nil {
		return err
	}
	return s.StateStorage.SetChannelPts(ctx, id, channel, v)
}
func (s *applyFaultStore) GetChannelPts(ctx context.Context, id, channel int64) (pts int, found bool, err error) {
	if err := s.check("GetChannelPts"); err != nil {
		return 0, false, err
	}
	return s.StateStorage.GetChannelPts(ctx, id, channel)
}

type applyBoundaryAPI struct {
	common  func(context.Context, *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error)
	channel func(context.Context, *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error)
}

func (*applyBoundaryAPI) UpdatesGetState(context.Context) (*tg.UpdatesState, error) {
	return &tg.UpdatesState{}, nil
}
func (a *applyBoundaryAPI) UpdatesGetDifference(ctx context.Context, r *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error) {
	return a.common(ctx, r)
}
func (a *applyBoundaryAPI) UpdatesGetChannelDifference(ctx context.Context, r *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
	return a.channel(ctx, r)
}

func TestManagerStopOnApplyError(t *testing.T) {
	live := func(u tg.UpdateClass, seq, date int) tg.UpdatesClass {
		return &tg.Updates{Updates: []tg.UpdateClass{u}, Seq: seq, Date: date}
	}
	pts := live(&tg.UpdateDeleteMessages{Messages: []int{1}, Pts: 11, PtsCount: 1}, 0, 0)
	qts := live(&tg.UpdateChannelParticipant{ChannelID: 42, UserID: 7, Qts: 11}, 0, 0)
	channel := live(&tg.UpdateDeleteChannelMessages{ChannelID: 42, Messages: []int{1}, Pts: 11, PtsCount: 1}, 0, 0)
	messages := []tg.MessageClass{&tg.Message{ID: 1, PeerID: &tg.PeerUser{UserID: 7}}}
	channelMessages := []tg.MessageClass{&tg.Message{ID: 1, PeerID: &tg.PeerChannel{ChannelID: 42}}}
	remote := tg.UpdatesState{Pts: 20, Qts: 10, Date: 2}
	dialog := &tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 42}}
	dialog.SetPts(20)
	for _, tc := range []struct {
		name, operation string // An empty operation fails Handler instead.
		input           tg.UpdatesClass
		difference      tg.UpdatesDifferenceClass
		channelDiff     tg.UpdatesChannelDifferenceClass
		savedChannel    bool
		affected        bool
		gap             bool
	}{
		{name: "PTS/Handler", input: pts},
		{name: "PTS/Checkpoint", operation: "SetPts", input: pts},
		{name: "QTS/Handler", input: qts},
		{name: "QTS/Checkpoint", operation: "SetQts", input: qts},
		{name: "UnsequencedQTS", input: live(&tg.UpdateChannelParticipant{ChannelID: 42, Qts: 0}, 0, 0)},
		{name: "Seq/Handler", input: live(&tg.UpdateConfig{}, 1, 0)},
		{name: "Seq/Checkpoint", operation: "SetSeq", input: live(&tg.UpdateConfig{}, 1, 0)},
		{name: "Date", operation: "SetDate", input: live(&tg.UpdateConfig{}, 0, 2)},
		{name: "DateSeq", operation: "SetDateSeq", input: live(&tg.UpdateConfig{}, 1, 2)},
		{name: "Difference/Handler", difference: &tg.UpdatesDifference{NewMessages: messages, State: remote}},
		{name: "Difference/Checkpoint", operation: "SetState", difference: &tg.UpdatesDifference{State: remote}},
		{name: "Slice/Handler", difference: &tg.UpdatesDifferenceSlice{NewMessages: messages, IntermediateState: remote}},
		{name: "Slice/Checkpoint", operation: "SetState", difference: &tg.UpdatesDifferenceSlice{IntermediateState: remote}},
		{name: "GapRecovery", gap: true, difference: &tg.UpdatesDifference{NewMessages: messages, State: remote}},
		{name: "Empty", operation: "SetDateSeq", difference: &tg.UpdatesDifferenceEmpty{Date: 2, Seq: 1}},
		{name: "TooLong", operation: "SetPts", difference: &tg.UpdatesDifferenceTooLong{Pts: 20}},
		{name: "ChannelLive/Handler", input: channel, savedChannel: true},
		{name: "ChannelLive/Checkpoint", operation: "SetChannelPts", input: channel, savedChannel: true},
		{name: "ChannelDifference/Handler", savedChannel: true, channelDiff: &tg.UpdatesChannelDifference{Final: true, Pts: 20, NewMessages: channelMessages}},
		{name: "ChannelDifference/Checkpoint", operation: "SetChannelPts", savedChannel: true, channelDiff: &tg.UpdatesChannelDifference{Final: true, Pts: 20}},
		{name: "ChannelGapRecovery", gap: true, savedChannel: true, channelDiff: &tg.UpdatesChannelDifference{Final: true, Pts: 20, NewMessages: channelMessages}},
		{name: "ChannelEmpty", operation: "SetChannelPts", savedChannel: true, channelDiff: &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: 20}},
		{name: "ChannelTooLong", operation: "SetChannelPts", savedChannel: true, channelDiff: &tg.UpdatesChannelDifferenceTooLong{Final: true, Dialog: dialog}},
		{name: "InitialRead", operation: "GetChannelPts", input: channel},
		{name: "InitialWrite", operation: "SetChannelPts", input: channel},
		{name: "Affected", operation: "SetPts", affected: true},
		{name: "ChannelAffected", operation: "SetChannelPts", savedChannel: true, affected: true},
	} {
		for _, stop := range []bool{false, true} {
			mode := "Default"
			if stop {
				mode = "Stop"
			}
			t.Run(tc.name+"/"+mode, func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					initial := State{Pts: 10, Qts: 10, Date: 1}
					memory := newMemStorage()
					require.NoError(t, memory.SetState(ctx, 1, initial))
					if tc.savedChannel {
						require.NoError(t, memory.SetChannelPts(ctx, 1, 42, 10))
					}
					storage := &applyFaultStore{StateStorage: memory, operation: tc.operation, failure: errors.New("application failure")}
					hashes := newMemAccessHasher()
					require.NoError(t, hashes.SetChannelAccessHash(ctx, 1, 42, 123))
					var commonReply, channelReply atomic.Bool
					var commonCalls, channelCalls, delivered atomic.Int32
					api := &applyBoundaryAPI{
						common: func(context.Context, *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error) {
							commonCalls.Add(1)
							if tc.difference != nil && storage.armed.Load() && commonReply.CompareAndSwap(false, true) {
								return tc.difference, nil
							}
							return &tg.UpdatesDifferenceEmpty{Date: 1}, nil
						},
						channel: func(_ context.Context, r *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
							channelCalls.Add(1)
							if tc.channelDiff != nil && storage.armed.Load() && channelReply.CompareAndSwap(false, true) {
								return tc.channelDiff, nil
							}
							return &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: r.Pts}, nil
						},
					}
					m := New(Config{Storage: storage, AccessHasher: hashes, StopOnApplyError: stop,
						OnTooLong: func() {}, OnChannelTooLong: func(int64) {},
						Handler: telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error {
							if storage.armed.Load() && tc.operation == "" {
								storage.hits.Add(1)
								return storage.failure
							}
							delivered.Add(1)
							return nil
						}),
					})
					done := make(chan error, 1)
					finished := false
					defer func() {
						cancel()
						if !finished {
							<-done
						}
					}()
					go func() { done <- m.Run(ctx, api, 1, AuthOptions{}) }()
					synctest.Wait() // Startup is healthy; each case targets the later public input.
					storage.armed.Store(true)
					switch {
					case tc.gap:
						var gap tg.UpdateClass = &tg.UpdateDeleteMessages{Messages: []int{2}, Pts: 12, PtsCount: 1}
						if tc.savedChannel {
							gap = &tg.UpdateDeleteChannelMessages{ChannelID: 42, Messages: []int{2}, Pts: 12, PtsCount: 1}
						}
						require.NoError(t, m.Handle(ctx, live(gap, 0, 0)))
						synctest.Wait()
						time.Sleep(fastgapTimeout)
					case tc.affected:
						id := int64(0)
						if tc.savedChannel {
							id = 42
						}
						require.NoError(t, m.HandleAffected(ctx, id, 11, 1))
					case tc.difference != nil:
						require.NoError(t, m.Handle(ctx, &tg.UpdatesTooLong{}))
					case tc.channelDiff != nil:
						require.NoError(t, m.Handle(ctx, live(&tg.UpdateChannelTooLong{ChannelID: 42}, 0, 0)))
					default:
						require.NoError(t, m.Handle(ctx, tc.input))
					}
					synctest.Wait()
					require.Positive(t, storage.hits.Load(), "the chosen application boundary must actually fail")
					if stop {
						select {
						case err := <-done:
							finished = true
							require.ErrorIs(t, err, storage.failure)
						default:
							t.Fatal("Run did not stop after the application failure")
						}
						state, found, err := memory.GetState(ctx, 1)
						require.NoError(t, err)
						require.True(t, found)
						require.Equal(t, initial, state, "this fixture has no earlier completed common step")
						p, found, err := memory.GetChannelPts(ctx, 1, 42)
						require.NoError(t, err)
						require.Equal(t, tc.savedChannel, found)
						if found {
							require.Equal(t, 10, p)
						}
						writes := int32(1)
						if tc.operation == "" || tc.operation == "GetChannelPts" {
							writes = 0
						}
						require.Equal(t, writes, storage.writes.Load(), "no checkpoint operation follows the failed step")
						if tc.difference != nil {
							require.EqualValues(t, 2, commonCalls.Load(), "no continuation after failure")
						}
						if tc.name == "InitialRead" || tc.name == "InitialWrite" {
							require.Zero(t, channelCalls.Load(), "no worker after failed initialization")
						}
					} else {
						select {
						case err := <-done:
							finished = true
							t.Fatalf("default mode stopped: %v", err)
						default:
						}
						storage.armed.Store(false)
						before := delivered.Load()
						require.NoError(t, m.Handle(ctx, live(&tg.UpdateConfig{}, 0, 0)))
						synctest.Wait()
						require.Equal(t, before+1, delivered.Load(), "default mode continues delivery")
					}
				})
			})
		}
	}
}

func TestManagerStopOnApplyErrorRestart(t *testing.T) {
	for _, checkpointFailure := range []bool{false, true} {
		name := "PartialHandler"
		if checkpointFailure {
			name = "CheckpointAfterHandler"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				memory := newMemStorage()
				require.NoError(t, memory.SetState(ctx, 1, State{Pts: 10, Date: 1}))
				storage := &applyFaultStore{StateStorage: memory, operation: "SetState", failure: errors.New("apply failed")}
				storage.armed.Store(checkpointFailure)
				var partial atomic.Bool
				partial.Store(!checkpointFailure)
				var attempts [2]atomic.Int32
				var present [2]atomic.Bool // Idempotent effects keyed by message ID.
				requests := make(chan int, 2)
				api := &applyBoundaryAPI{common: func(_ context.Context, r *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error) {
					requests <- r.Pts
					return &tg.UpdatesDifference{State: tg.UpdatesState{Pts: 12, Date: 2}, NewMessages: []tg.MessageClass{
						&tg.Message{ID: 1, PeerID: &tg.PeerUser{UserID: 7}},
						&tg.Message{ID: 2, PeerID: &tg.PeerUser{UserID: 7}},
					}}, nil
				}}
				cfg := Config{Storage: storage, StopOnApplyError: true,
					Handler: telegram.UpdateHandlerFunc(func(_ context.Context, u tg.UpdatesClass) error {
						for _, update := range u.(*tg.Updates).Updates {
							id := update.(*tg.UpdateNewMessage).Message.GetID() - 1
							attempts[id].Add(1)
							present[id].Store(true)
							if partial.Load() {
								return storage.failure
							}
						}
						return nil
					}),
				}
				require.ErrorIs(t, New(cfg).Run(ctx, api, 1, AuthOptions{Forget: false}), storage.failure)
				state, _, err := memory.GetState(ctx, 1)
				require.NoError(t, err)
				require.Equal(t, 10, state.Pts)
				require.True(t, present[0].Load(), "effects before failure cannot be rolled back")
				storage.armed.Store(false)
				partial.Store(false)
				done := make(chan error, 1)
				go func() { done <- New(cfg).Run(ctx, api, 1, AuthOptions{Forget: false}) }()
				synctest.Wait()
				cancel()
				require.ErrorIs(t, <-done, context.Canceled)
				require.Equal(t, 10, <-requests)
				require.Equal(t, 10, <-requests, "the new manager requests replay from accepted storage")
				state, _, err = memory.GetState(ctx, 1)
				require.NoError(t, err)
				require.Equal(t, 12, state.Pts)
				require.True(t, present[1].Load())
				require.EqualValues(t, 2, attempts[0].Load(), "delivery is not exactly once")
				wantSecond := int32(1)
				if checkpointFailure {
					wantSecond = 2
				}
				require.Equal(t, wantSecond, attempts[1].Load())
			})
		})
	}
}

func TestManagerStopOnApplyErrorCancellationAdmission(t *testing.T) {
	for _, boundary := range []string{"CommonResponse", "ChannelResponse", "BeforeStartup"} {
		t.Run(boundary, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				storage, hashes := newMemStorage(), newMemAccessHasher()
				initial := State{Pts: 10, Date: 1}
				require.NoError(t, storage.SetState(ctx, 1, initial))
				channel := boundary == "ChannelResponse"
				if channel {
					require.NoError(t, storage.SetChannelPts(ctx, 1, 42, 10))
					require.NoError(t, hashes.SetChannelAccessHash(ctx, 1, 42, 123))
				}
				var calls, handled atomic.Int32
				api := &applyBoundaryAPI{
					common: func(context.Context, *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error) {
						if channel {
							return &tg.UpdatesDifferenceEmpty{Date: 1}, nil
						}
						calls.Add(1)
						cancel() // A successful response may race account cancellation.
						return &tg.UpdatesDifference{State: tg.UpdatesState{Pts: 20, Date: 2},
							NewMessages: []tg.MessageClass{&tg.Message{ID: 1, PeerID: &tg.PeerUser{UserID: 7}}}}, nil
					},
					channel: func(context.Context, *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
						calls.Add(1)
						cancel()
						return &tg.UpdatesChannelDifference{Final: true, Pts: 20,
							NewMessages: []tg.MessageClass{&tg.Message{ID: 1, PeerID: &tg.PeerChannel{ChannelID: 42}}}}, nil
					},
				}
				m := New(Config{Storage: storage, AccessHasher: hashes, StopOnApplyError: true,
					Handler: telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error { handled.Add(1); return nil }),
				})
				wantCalls := int32(1)
				if boundary == "BeforeStartup" {
					cancel()
					wantCalls = 0
				}
				require.ErrorIs(t, m.Run(ctx, api, 1, AuthOptions{}), context.Canceled)
				require.Equal(t, wantCalls, calls.Load())
				require.Zero(t, handled.Load(), "do not admit response callbacks after cancellation")
				state, _, err := storage.GetState(ctx, 1)
				require.NoError(t, err)
				require.Equal(t, initial, state)
				if channel {
					pts, found, err := storage.GetChannelPts(ctx, 1, 42)
					require.NoError(t, err)
					require.True(t, found)
					require.Equal(t, 10, pts)
				}
			})
		})
	}
}
