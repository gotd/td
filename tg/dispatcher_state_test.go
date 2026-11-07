package tg_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func TestUpdateDispatcherCopySharesHandlers(t *testing.T) {
	ctx := context.Background()
	d := tg.NewUpdateDispatcher()
	// Copy, as telegram.Options{UpdateHandler: d} or storage.UpdateHook(d, ...) do.
	var h telegram.UpdateHandler = d

	var gotNew, gotFallback bool
	d.OnNewMessage(func(context.Context, tg.Entities, *tg.UpdateNewMessage) error {
		gotNew = true
		return nil
	})
	d.OnFallback(func(context.Context, tg.Entities, tg.UpdateClass) error {
		gotFallback = true
		return nil
	})

	require.NoError(t, h.Handle(ctx, &tg.Updates{Updates: []tg.UpdateClass{
		&tg.UpdateNewMessage{Message: &tg.MessageEmpty{}},
		&tg.UpdateUserTyping{Action: &tg.SendMessageTypingAction{}},
	}}))
	require.True(t, gotNew, "OnNewMessage via copy")
	require.True(t, gotFallback, "OnFallback via copy")
}

func TestUpdateDispatcherRegisterConcurrently(t *testing.T) {
	ctx := context.Background()
	d := tg.NewUpdateDispatcher()
	updates := &tg.Updates{Updates: []tg.UpdateClass{
		&tg.UpdateNewMessage{Message: &tg.MessageEmpty{}},
	}}

	var (
		wg    sync.WaitGroup
		calls atomic.Int64
	)
	for range 4 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for range 100 {
				_ = d.Handle(ctx, updates)
			}
		}()
		go func() {
			defer wg.Done()
			for range 100 {
				d.OnNewMessage(func(context.Context, tg.Entities, *tg.UpdateNewMessage) error {
					calls.Add(1)
					return nil
				})
				d.OnFallback(func(context.Context, tg.Entities, tg.UpdateClass) error { return nil })
			}
		}()
	}
	wg.Wait()

	before := calls.Load()
	require.NoError(t, d.Handle(ctx, updates))
	require.Equal(t, before+1, calls.Load())
}
