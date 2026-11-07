package updates

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// beforeRunHandler records IDs of new messages it got.
type beforeRunHandler struct {
	mu  sync.Mutex
	ids []int
}

func (h *beforeRunHandler) Handle(_ context.Context, u tg.UpdatesClass) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	upds, ok := u.(*tg.Updates)
	if !ok {
		return nil
	}
	for _, upd := range upds.Updates {
		if m, ok := upd.(*tg.UpdateNewMessage); ok {
			h.ids = append(h.ids, m.Message.GetID())
		}
	}
	return nil
}

func (h *beforeRunHandler) got() []int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]int(nil), h.ids...)
}

var _ telegram.UpdateHandler = (*beforeRunHandler)(nil)

// newMessageUpdates returns a new message with the given ID and pts.
func newMessageUpdates(id, pts int) tg.UpdatesClass {
	return &tg.Updates{
		Updates: []tg.UpdateClass{&tg.UpdateNewMessage{
			Message:  &tg.Message{ID: id, PeerID: &tg.PeerUser{UserID: 10}},
			Pts:      pts,
			PtsCount: 1,
		}},
		Users: []tg.UserClass{&tg.User{ID: 10, AccessHash: 10}},
		Date:  1,
	}
}

func TestManagerBeforeRun(t *testing.T) {
	setup := func(t *testing.T, mode BeforeRunMode, size int) (*Manager, *beforeRunHandler, func()) {
		t.Helper()
		storage := newMemStorage()
		require.NoError(t, storage.SetState(context.Background(), trackUserID, State{Pts: 1, Qts: 1, Seq: 1, Date: 1}))
		h := &beforeRunHandler{}
		m := New(Config{
			Handler:             h,
			Storage:             storage,
			Logger:              logzap.New(zaptest.NewLogger(t)),
			BeforeRun:           mode,
			BeforeRunBufferSize: size,
		})
		run := func() {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			started := make(chan struct{})
			go func() {
				done <- m.Run(ctx, &trackAPI{}, trackUserID, AuthOptions{
					OnStart: func(context.Context) { close(started) },
				})
			}()
			t.Cleanup(func() {
				cancel()
				<-done
			})
			<-started
		}
		return m, h, run
	}
	ctx := context.Background()
	eventually := func(t *testing.T, h *beforeRunHandler, want []int) {
		t.Helper()
		require.Eventually(t, func() bool {
			return equalInts(h.got(), want)
		}, 5*time.Second, time.Millisecond, "got %v", h.got())
	}

	t.Run("Pass", func(t *testing.T) {
		m, h, _ := setup(t, BeforeRunPass, 0)
		require.NoError(t, m.Handle(ctx, newMessageUpdates(1, 2)))
		require.Equal(t, []int{1}, h.got(), "passed as-is before Run")
	})
	t.Run("Buffer", func(t *testing.T) {
		m, h, run := setup(t, BeforeRunBuffer, 0)
		require.NoError(t, m.Handle(ctx, newMessageUpdates(1, 2)))
		require.NoError(t, m.Handle(ctx, newMessageUpdates(2, 3)))
		require.Empty(t, h.got(), "not handled before Run")

		run()
		eventually(t, h, []int{1, 2})

		// Updates after start go through as usual.
		require.NoError(t, m.Handle(ctx, newMessageUpdates(3, 4)))
		eventually(t, h, []int{1, 2, 3})
	})
	t.Run("BufferLimit", func(t *testing.T) {
		m, h, run := setup(t, BeforeRunBuffer, 1)
		require.NoError(t, m.Handle(ctx, newMessageUpdates(1, 2)))
		require.NoError(t, m.Handle(ctx, newMessageUpdates(2, 3)))

		run()
		eventually(t, h, []int{1})
		require.NoError(t, m.Handle(ctx, newMessageUpdates(3, 3)), "same pts as dropped one")
		eventually(t, h, []int{1, 3})
	})
	t.Run("Drop", func(t *testing.T) {
		m, h, run := setup(t, BeforeRunDrop, 0)
		require.NoError(t, m.Handle(ctx, newMessageUpdates(1, 2)))
		run()
		require.NoError(t, m.Handle(ctx, newMessageUpdates(2, 2)))
		eventually(t, h, []int{2})
	})
	t.Run("Reset", func(t *testing.T) {
		m, h, run := setup(t, BeforeRunBuffer, 0)
		require.NoError(t, m.Handle(ctx, newMessageUpdates(1, 2)))
		// Buffered updates of the previous user are dropped.
		m.Reset()
		run()
		require.NoError(t, m.Handle(ctx, newMessageUpdates(2, 2)))
		eventually(t, h, []int{2})
	})
}
