package updates

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/tg"
)

// orderHandler records handled updates in order: message IDs of new
// messages, and -UserID of user statuses (updates without pts).
type orderHandler struct {
	mu  sync.Mutex
	got []int
}

func (h *orderHandler) Handle(_ context.Context, u tg.UpdatesClass) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	var upds []tg.UpdateClass
	switch u := u.(type) {
	case *tg.Updates:
		upds = u.Updates
	case *tg.UpdateShort:
		upds = []tg.UpdateClass{u.Update}
	}
	for _, upd := range upds {
		switch upd := upd.(type) {
		case *tg.UpdateNewMessage:
			h.got = append(h.got, upd.Message.GetID())
		case *tg.UpdateUserStatus:
			h.got = append(h.got, -int(upd.UserID))
		}
	}
	return nil
}

func (h *orderHandler) list() []int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]int(nil), h.got...)
}

func userStatusUpdate(userID int64) tg.UpdatesClass {
	return &tg.UpdateShort{
		Update: &tg.UpdateUserStatus{UserID: userID, Status: &tg.UserStatusOnline{Expires: 1}},
		Date:   1,
	}
}

// remoteStateAPI returns a remote state with the given pts.
type remoteStateAPI struct {
	trackAPI
	pts int
}

func (a *remoteStateAPI) UpdatesGetState(context.Context) (*tg.UpdatesState, error) {
	return &tg.UpdatesState{Pts: a.pts, Qts: 1, Seq: 1, Date: 1}, nil
}

func runBeforeRun(t *testing.T, m *Manager, api API, onStart func(ctx context.Context)) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	started := make(chan struct{})
	go func() {
		done <- m.Run(ctx, api, trackUserID, AuthOptions{
			OnStart: func(ctx context.Context) {
				if onStart != nil {
					onStart(ctx)
				}
				close(started)
			},
		})
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})
	<-started
}

// TestManagerBeforeRunOrder checks that updates received while buffered
// updates are replayed can't overtake them, including updates without pts.
func TestManagerBeforeRunOrder(t *testing.T) {
	ctx := context.Background()
	storage := newMemStorage()
	require.NoError(t, storage.SetState(ctx, trackUserID, State{Pts: 1, Qts: 1, Seq: 1, Date: 1}))
	h := &orderHandler{}
	m := New(Config{
		Handler:   h,
		Storage:   storage,
		Logger:    logzap.New(zaptest.NewLogger(t)),
		BeforeRun: BeforeRunBuffer,
	})
	require.NoError(t, m.Handle(ctx, newMessageUpdates(1, 2)))
	require.NoError(t, m.Handle(ctx, newMessageUpdates(2, 3)))

	runBeforeRun(t, m, &trackAPI{}, func(ctx context.Context) {
		// Live updates received after the state is loaded, before replay.
		require.NoError(t, m.Handle(ctx, userStatusUpdate(10)))
		require.NoError(t, m.Handle(ctx, newMessageUpdates(3, 4)))
	})
	require.NoError(t, m.Handle(ctx, userStatusUpdate(11)))

	want := []int{1, 2, -10, 3, -11}
	require.Eventually(t, func() bool {
		return equalInts(h.list(), want)
	}, 5*time.Second, time.Millisecond, "got %v, want %v", h.list(), want)
}

// TestManagerBeforeRunRemoteState checks that buffered updates are not lost
// when the state is fetched from the server: they are covered by it, so gap
// recovery would drop them as outdated.
func TestManagerBeforeRunRemoteState(t *testing.T) {
	ctx := context.Background()
	h := &orderHandler{}
	m := New(Config{
		Handler:   h,
		Storage:   newMemStorage(), // no stored state
		Logger:    logzap.New(zaptest.NewLogger(t)),
		BeforeRun: BeforeRunBuffer,
	})
	require.NoError(t, m.Handle(ctx, newMessageUpdates(1, 2)))
	require.NoError(t, m.Handle(ctx, newMessageUpdates(2, 3)))

	runBeforeRun(t, m, &remoteStateAPI{pts: 10}, nil)
	// Live update after the remote state goes through gap recovery.
	require.NoError(t, m.Handle(ctx, newMessageUpdates(3, 11)))

	want := []int{1, 2, 3}
	require.Eventually(t, func() bool {
		return equalInts(h.list(), want)
	}, 5*time.Second, time.Millisecond, "got %v, want %v", h.list(), want)
}
