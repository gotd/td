package takeout

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/neo"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

type recordInvoker struct {
	mux        sync.Mutex
	initErrs   []error
	inits      int
	takeoutIDs []int64
	finished   []bool
}

func (r *recordInvoker) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	switch req := input.(type) {
	case *tg.AccountInitTakeoutSessionRequest:
		r.inits++
		if len(r.initErrs) > 0 {
			err := r.initErrs[0]
			r.initErrs = r.initErrs[1:]
			return err
		}
		output.(*tg.AccountTakeout).ID = 42
		return nil
	case *tg.AccountFinishTakeoutSessionRequest:
		r.finished = append(r.finished, req.GetSuccess())
		return ctx.Err()
	case *tg.InvokeWithTakeoutRequest:
		r.takeoutIDs = append(r.takeoutIDs, req.TakeoutID)
		return nil
	}
	return nil
}

func TestNewClientResume(t *testing.T) {
	ctx := context.Background()
	inv := &recordInvoker{}
	c := NewClient(7, inv)
	require.Equal(t, int64(7), c.ID())

	_, err := tg.NewClient(c).HelpGetConfig(ctx)
	require.NoError(t, err)
	require.NoError(t, c.Finish(ctx, true))
	require.Equal(t, []int64{7}, inv.takeoutIDs)
	require.Equal(t, []bool{true}, inv.finished)
}

func TestInitDelay(t *testing.T) {
	delay := tgerr.New(420, "TAKEOUT_INIT_DELAY_3600")
	d, ok := AsInitDelay(errors.Wrap(delay, "wrap"))
	require.True(t, ok)
	require.Equal(t, time.Hour, d)

	t.Run("NoCallback", func(t *testing.T) {
		inv := &recordInvoker{initErrs: []error{delay}}
		_, err := Init(context.Background(), inv, Config{})
		_, ok := AsInitDelay(err)
		require.True(t, ok)
	})
	t.Run("Poll", func(t *testing.T) {
		clk := neo.NewTime(time.Unix(0, 0))
		inv := &recordInvoker{initErrs: []error{delay, delay}}
		var delays []time.Duration
		cfg := Config{
			Clock:            clk,
			InitPollInterval: time.Minute,
			OnInitDelay: func(ctx context.Context, d time.Duration) error {
				delays = append(delays, d)
				return nil
			},
		}

		done := make(chan error, 1)
		var c *Client
		// Observe before timer is planned.
		planned := clk.Observe()
		go func() {
			var err error
			c, err = Init(context.Background(), inv, cfg)
			done <- err
		}()
		for range 2 {
			<-planned
			planned = clk.Observe()
			clk.Travel(time.Minute)
		}
		require.NoError(t, <-done)
		require.Equal(t, int64(42), c.ID())
		require.Equal(t, []time.Duration{time.Hour, time.Hour}, delays)
		require.Equal(t, 3, inv.inits)
	})
	t.Run("Abort", func(t *testing.T) {
		inv := &recordInvoker{initErrs: []error{delay}}
		abort := errors.New("user declined")
		_, err := Init(context.Background(), inv, Config{
			OnInitDelay: func(context.Context, time.Duration) error { return abort },
		})
		require.ErrorIs(t, err, abort)
	})
	t.Run("Canceled", func(t *testing.T) {
		clk := neo.NewTime(time.Unix(0, 0))
		inv := &recordInvoker{initErrs: []error{delay}}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		planned := clk.Observe()
		go func() {
			_, err := Init(ctx, inv, Config{
				Clock:       clk,
				OnInitDelay: func(context.Context, time.Duration) error { return nil },
			})
			done <- err
		}()
		<-planned
		cancel()
		require.ErrorIs(t, <-done, context.Canceled)
	})
}

func TestRunFinishesAfterCancel(t *testing.T) {
	inv := &recordInvoker{}
	ctx, cancel := context.WithCancel(context.Background())
	err := Run(ctx, inv, Config{}, func(ctx context.Context, client *Client) error {
		cancel()
		return ctx.Err()
	})
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, []bool{false}, inv.finished, "session is finished with canceled ctx")
}
