package updates

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type channelDifferenceFunc struct {
	API
	call func(context.Context, *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error)
}

func (a channelDifferenceFunc) UpdatesGetChannelDifference(ctx context.Context, req *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
	return a.call(ctx, req)
}

func TestChannelRecoveryBeforePollTimeout(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input tg.UpdateClass
		delay time.Duration
	}{
		{
			name: "pts gap",
			input: &tg.UpdateNewChannelMessage{
				Message: &tg.Message{ID: 12, PeerID: &tg.PeerChannel{ChannelID: 42}}, Pts: 12, PtsCount: 1,
			},
			delay: fastgapTimeout,
		},
		{name: "too long", input: &tg.UpdateChannelTooLong{ChannelID: 42}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				started := time.Now()
				var calls, delivered int
				api := channelDifferenceFunc{call: func(_ context.Context, req *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
					calls++
					require.Equal(t, 10, req.Pts)
					if calls == 1 {
						diff := &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: 10}
						diff.SetTimeout(300)
						return diff, nil
					}
					require.Equal(t, 2, calls)
					require.Equal(t, tt.delay, time.Since(started))
					return &tg.UpdatesChannelDifference{
						Final: true, Pts: 12,
						NewMessages: []tg.MessageClass{
							&tg.Message{ID: 11, PeerID: &tg.PeerChannel{ChannelID: 42}},
							&tg.Message{ID: 12, PeerID: &tg.PeerChannel{ChannelID: 42}},
						},
					}, nil
				}}
				s := newShortTestState(t, api, telegram.UpdateHandlerFunc(func(_ context.Context, u tg.UpdatesClass) error {
					delivered += len(u.(*tg.Updates).Updates)
					return nil
				})).newChannelState(42, 1, 10)
				done := make(chan error, 1)
				go func() { done <- s.Run(ctx) }()
				synctest.Wait()
				require.NoError(t, s.Push(ctx, channelUpdate{update: tt.input}))
				synctest.Wait()
				time.Sleep(tt.delay)
				synctest.Wait()
				require.Equal(t, 2, calls, "recovery must not wait for the 300-second polling deadline")
				require.Equal(t, 2, delivered)

				// Normal push processing must continue after recovery.
				require.NoError(t, s.Push(ctx, channelUpdate{update: &tg.UpdateDeleteChannelMessages{
					ChannelID: 42, Messages: []int{13}, Pts: 13, PtsCount: 1,
				}}))
				synctest.Wait()
				require.Equal(t, 3, delivered)
				pts, found, err := s.storage.GetChannelPts(ctx, s.selfID, 42)
				require.NoError(t, err)
				require.True(t, found)
				require.Equal(t, 13, pts)
				cancel()
				require.ErrorIs(t, <-done, context.Canceled)
			})
		})
	}
}

func TestChannelNonFinalDifferenceContinuesImmediately(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		started := time.Now()
		var requested []int
		api := channelDifferenceFunc{call: func(_ context.Context, req *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
			requested = append(requested, req.Pts)
			if len(requested) == 1 {
				// Synthetic timeout on an intermediate response: continuation
				// must ignore it, as only final responses schedule polling.
				diff := &tg.UpdatesChannelDifference{Pts: 11}
				diff.SetTimeout(3600)
				return diff, nil
			}
			if len(requested) == 2 {
				require.Equal(t, started, time.Now())
			}
			return &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: 12}, nil
		}}
		s := newShortTestState(t, api, telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error { return nil })).newChannelState(42, 1, 10)
		s.chDiffSem = newChDiffSem(1)
		done := make(chan error, 1)
		go func() { done <- s.Run(ctx) }()
		synctest.Wait()
		require.Equal(t, []int{10, 11}, requested)
		time.Sleep(idleTimeout)
		synctest.Wait()
		require.Equal(t, []int{10, 11, 12}, requested, "an intermediate timeout must not delay later idle polling")
		cancel()
		require.ErrorIs(t, <-done, context.Canceled)
	})
}

func TestChannelIdlePollTimeout(t *testing.T) {
	for _, activity := range []bool{false, true} {
		name := "cancel while deferred"
		if activity {
			name = "live update and RPC error"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				var calls, delivered int
				api := channelDifferenceFunc{call: func(_ context.Context, req *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
					calls++
					if calls == 1 {
						diff := &tg.UpdatesChannelDifference{Final: true, Pts: 10}
						diff.SetTimeout(int(2 * idleTimeout / time.Second))
						return diff, nil
					}
					require.Equal(t, 11, req.Pts)
					return nil, errors.New("offline RPC failure")
				}}
				s := newShortTestState(t, api, telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error {
					delivered++
					return nil
				})).newChannelState(42, 1, 10)
				done := make(chan error, 1)
				go func() { done <- s.Run(ctx) }()
				synctest.Wait()
				time.Sleep(idleTimeout)
				synctest.Wait()
				require.Equal(t, 1, calls, "idle polling must respect the server deadline")
				if activity {
					require.NoError(t, s.Push(ctx, channelUpdate{update: &tg.UpdateDeleteChannelMessages{
						ChannelID: 42, Messages: []int{11}, Pts: 11, PtsCount: 1,
					}}))
					synctest.Wait()
					require.Equal(t, 1, delivered, "deferred idle polling must not discard live updates")
					time.Sleep(idleTimeout)
					synctest.Wait()
					require.Equal(t, 2, calls)
					time.Sleep(time.Second)
					synctest.Wait()
					require.Equal(t, 2, calls, "an RPC error must not cause immediate retries")
				}
				cancel()
				require.ErrorIs(t, <-done, context.Canceled)
			})
		})
	}
}
