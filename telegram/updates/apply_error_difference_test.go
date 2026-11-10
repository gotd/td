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

// Pending updates that continue the state set by a difference are applied
// after it (sequenceBox.resyncPending). The difference itself is a completed
// step and is not rolled back when the pending updates fail.
func TestManagerStopOnApplyErrorDifferencePending(t *testing.T) {
	for _, channel := range []bool{false, true} {
		for _, stop := range []bool{false, true} {
			name := "Common"
			if channel {
				name = "Channel"
			}
			mode := "Default"
			if stop {
				mode = "Stop"
			}
			t.Run(name+"/"+mode, func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					memory, hashes := newMemStorage(), newMemAccessHasher()
					require.NoError(t, memory.SetState(ctx, 1, State{Pts: 10, Date: 1}))
					if channel {
						require.NoError(t, memory.SetChannelPts(ctx, 1, 42, 10))
						require.NoError(t, hashes.SetChannelAccessHash(ctx, 1, 42, 123))
					}
					failure := errors.New("pending handler failure")
					var armed atomic.Bool
					var hits, delivered atomic.Int32
					api := &applyBoundaryAPI{
						common: func(context.Context, *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error) {
							if armed.Load() && !channel {
								// Covers pts up to 20, the pending update 21 continues it.
								return &tg.UpdatesDifference{State: tg.UpdatesState{Pts: 20, Date: 2}}, nil
							}
							return &tg.UpdatesDifferenceEmpty{Date: 1}, nil
						},
						channel: func(_ context.Context, r *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
							if armed.Load() {
								return &tg.UpdatesChannelDifference{Final: true, Pts: 20}, nil
							}
							return &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: r.Pts}, nil
						},
					}
					m := New(Config{Storage: memory, AccessHasher: hashes, StopOnApplyError: stop,
						Handler: telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error {
							if armed.Load() {
								hits.Add(1)
								return failure
							}
							delivered.Add(1)
							return nil
						}),
					})
					done := make(chan error, 1)
					go func() { done <- m.Run(ctx, api, 1, AuthOptions{}) }()
					synctest.Wait()
					armed.Store(true)

					// A gap: 11..20 are missing.
					var u tg.UpdateClass = &tg.UpdateDeleteMessages{Messages: []int{1}, Pts: 21, PtsCount: 1}
					if channel {
						u = &tg.UpdateDeleteChannelMessages{ChannelID: 42, Messages: []int{1}, Pts: 21, PtsCount: 1}
					}
					require.NoError(t, m.Handle(ctx, &tg.Updates{Updates: []tg.UpdateClass{u}}))
					synctest.Wait()
					require.Zero(t, hits.Load(), "postponed by the gap")
					time.Sleep(fastgapTimeout)
					synctest.Wait()
					require.EqualValues(t, 1, hits.Load(), "pending update applied after the difference")

					state, _, err := memory.GetState(ctx, 1)
					require.NoError(t, err)
					channelPts, _, err := memory.GetChannelPts(ctx, 1, 42)
					require.NoError(t, err)
					if stop {
						select {
						case err := <-done:
							require.ErrorIs(t, err, failure)
						default:
							t.Fatal("Run did not stop after the pending failure")
						}
						// The difference is checkpointed, the failed pending
						// update is not.
						if channel {
							require.Equal(t, 20, channelPts)
						} else {
							require.Equal(t, State{Pts: 20, Date: 2}, state)
						}
						return
					}

					select {
					case err := <-done:
						t.Fatalf("default mode stopped: %v", err)
					default:
					}
					if channel {
						// Unchanged default: a failed channel handler skips the
						// checkpoint write.
						require.Equal(t, 20, channelPts)
					} else {
						require.Equal(t, State{Pts: 21, Date: 2}, state)
					}
					armed.Store(false)
					before := delivered.Load()
					require.NoError(t, m.Handle(ctx, &tg.Updates{Updates: []tg.UpdateClass{&tg.UpdateConfig{}}}))
					synctest.Wait()
					require.Equal(t, before+1, delivered.Load(), "default mode continues delivery")
					cancel()
					require.ErrorIs(t, <-done, context.Canceled)
				})
			})
		}
	}
}
