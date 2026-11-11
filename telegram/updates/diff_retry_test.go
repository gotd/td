package updates

import (
	"context"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// scriptedDiffAPI is a fake API whose getDifference and getState are scripted.
type scriptedDiffAPI struct {
	API
	diff  func(n int) (tg.UpdatesDifferenceClass, error)
	state func() (*tg.UpdatesState, error)
	calls atomic.Int64
}

func (a *scriptedDiffAPI) UpdatesGetDifference(context.Context, *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error) {
	return a.diff(int(a.calls.Add(1)))
}

func (a *scriptedDiffAPI) UpdatesGetState(context.Context) (*tg.UpdatesState, error) {
	return a.state()
}

func nopUpdateHandler() telegram.UpdateHandler {
	return telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error { return nil })
}

func runState(t *testing.T, s *internalState) (cancel func()) {
	t.Helper()
	ctx, stop := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx) }()
	return func() {
		stop()
		require.ErrorIs(t, <-done, context.Canceled)
	}
}

// A failed getDifference is retried with doubling delays until it succeeds,
// then retries stop (UPD-4). Before, it waited for the 15-minute idle timer.
func TestDifferenceRetryBackoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		api := &scriptedDiffAPI{diff: func(n int) (tg.UpdatesDifferenceClass, error) {
			if n <= 3 {
				return nil, tgerr.New(500, "INTERNAL_SERVER_ERROR")
			}
			return &tg.UpdatesDifferenceEmpty{Date: 1, Seq: 0}, nil
		}}
		stop := runState(t, newShortTestState(t, api, nopUpdateHandler()))
		defer stop()

		synctest.Wait()
		require.EqualValues(t, 1, api.calls.Load(), "startup difference")
		for _, step := range []struct {
			wait  time.Duration
			calls int64
		}{
			{time.Second - time.Millisecond, 1},
			{time.Millisecond, 2}, // +1 s
			{2*time.Second - time.Millisecond, 2},
			{time.Millisecond, 3}, // +2 s
			{4 * time.Second, 4},  // +4 s, succeeds
			{5 * time.Minute, 4},  // no more retries
		} {
			time.Sleep(step.wait)
			synctest.Wait()
			require.EqualValues(t, step.calls, api.calls.Load())
		}
	})
}

// FLOOD_WAIT is waited out before the retry.
func TestDifferenceRetryFloodWait(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		api := &scriptedDiffAPI{diff: func(n int) (tg.UpdatesDifferenceClass, error) {
			if n == 1 {
				return nil, tgerr.New(420, "FLOOD_WAIT_5")
			}
			return &tg.UpdatesDifferenceEmpty{Date: 1}, nil
		}}
		stop := runState(t, newShortTestState(t, api, nopUpdateHandler()))
		defer stop()

		synctest.Wait()
		time.Sleep(5*time.Second - time.Millisecond)
		synctest.Wait()
		require.EqualValues(t, 1, api.calls.Load(), "must wait out FLOOD_WAIT_5")
		time.Sleep(time.Millisecond)
		synctest.Wait()
		require.EqualValues(t, 2, api.calls.Load())
	})
}

// A lost authorization is not retried: no retry fixes it.
func TestDifferenceNoRetryOnAuthError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		api := &scriptedDiffAPI{diff: func(int) (tg.UpdatesDifferenceClass, error) {
			return nil, tgerr.New(401, "AUTH_KEY_UNREGISTERED")
		}}
		stop := runState(t, newShortTestState(t, api, nopUpdateHandler()))
		defer stop()

		synctest.Wait()
		time.Sleep(5 * time.Minute)
		synctest.Wait()
		require.EqualValues(t, 1, api.calls.Load())
	})
}

// PERSISTENT_TIMESTAMP_INVALID resets the common state from updates.getState
// and reports the lost updates with OnTooLong, keeping stored channels.
func TestDifferencePersistentTimestampReset(t *testing.T) {
	for _, typ := range []string{"PERSISTENT_TIMESTAMP_INVALID", "PERSISTENT_TIMESTAMP_EMPTY"} {
		t.Run(typ, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				api := &scriptedDiffAPI{
					diff: func(n int) (tg.UpdatesDifferenceClass, error) {
						if n == 1 {
							return nil, tgerr.New(400, typ)
						}
						return &tg.UpdatesDifferenceEmpty{Date: 50, Seq: 7}, nil
					},
					state: func() (*tg.UpdatesState, error) {
						return &tg.UpdatesState{Pts: 100, Qts: 5, Date: 50, Seq: 7}, nil
					},
				}
				s := newShortTestState(t, api, nopUpdateHandler())
				ctx := context.Background()
				require.NoError(t, s.storage.SetChannelPts(ctx, s.selfID, 42, 10))
				var tooLong atomic.Int64
				s.onCommonTooLong = func() { tooLong.Add(1) }
				stop := runState(t, s)

				synctest.Wait()
				stop()
				require.EqualValues(t, 1, tooLong.Load())
				require.Equal(t, 100, s.pts.State())
				require.Equal(t, 5, s.qts.State())
				require.Equal(t, 7, s.seq.State())
				st, ok, err := s.storage.GetState(ctx, s.selfID)
				require.NoError(t, err)
				require.True(t, ok)
				require.Equal(t, State{Pts: 100, Qts: 5, Date: 50, Seq: 7}, st)
				pts, ok, err := s.storage.GetChannelPts(ctx, s.selfID, 42)
				require.NoError(t, err)
				require.True(t, ok, "stored channels must be kept")
				require.Equal(t, 10, pts)
				require.EqualValues(t, 1, api.calls.Load(), "no retry after a reset")
			})
		})
	}
}

// Channel differences: PERSISTENT_TIMESTAMP_OUTDATED (and other transient
// errors) are retried with backoff; CHANNEL_INVALID stops tracking like
// CHANNEL_PRIVATE.
func TestChannelDifferenceRetry(t *testing.T) {
	t.Run("outdated", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var calls int
			api := channelDifferenceFunc{call: func(context.Context, *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
				calls++
				if calls <= 2 {
					return nil, tgerr.New(500, "PERSISTENT_TIMESTAMP_OUTDATED")
				}
				return &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: 10}, nil
			}}
			s := newShortTestState(t, api, nopUpdateHandler()).newChannelState(42, 1, 10)
			done := make(chan error, 1)
			go func() { done <- s.Run(ctx) }()
			synctest.Wait()
			require.Equal(t, 1, calls)
			time.Sleep(time.Second)
			synctest.Wait()
			require.Equal(t, 2, calls)
			time.Sleep(2 * time.Second)
			synctest.Wait()
			require.Equal(t, 3, calls)
			time.Sleep(time.Minute)
			synctest.Wait()
			require.Equal(t, 3, calls, "retries stop after success")
			cancel()
			require.ErrorIs(t, <-done, context.Canceled)
		})
	})
	t.Run("invalid", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var calls int
			api := channelDifferenceFunc{call: func(context.Context, *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
				calls++
				return nil, tgerr.New(400, "CHANNEL_INVALID")
			}}
			st := newShortTestState(t, api, nopUpdateHandler())
			var lost []int64
			st.onChannelInaccessible = func(id int64) { lost = append(lost, id) }
			s := st.newChannelState(42, 1, 10)
			require.NoError(t, s.Run(context.Background()), "worker stops")
			require.Equal(t, []int64{42}, lost)
			require.Equal(t, 1, calls)
		})
	})
}
