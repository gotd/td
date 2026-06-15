package updates

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// recordingHandler records every dispatched update for assertions.
type recordingHandler struct {
	mu      sync.Mutex
	batches []tg.UpdatesClass
}

func (h *recordingHandler) Handle(_ context.Context, u tg.UpdatesClass) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.batches = append(h.batches, u)
	return nil
}

func (h *recordingHandler) updates() (out []tg.UpdateClass) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, b := range h.batches {
		switch u := b.(type) {
		case *tg.Updates:
			out = append(out, u.Updates...)
		case *tg.UpdatesCombined:
			out = append(out, u.Updates...)
		}
	}
	return out
}

func newAffectedTestState(t *testing.T, h telegram.UpdateHandler, pts int) *internalState {
	t.Helper()
	ctx := context.Background()
	const selfID = 123
	storage := newMemStorage()
	require.NoError(t, storage.SetState(ctx, selfID, State{Pts: pts}))
	return newState(ctx, stateConfig{
		State:            State{Pts: pts},
		RawClient:        &diffAPI{diffs: []tg.UpdatesDifferenceClass{&tg.UpdatesDifferenceEmpty{}}},
		Logger:           logzap.New(zaptest.NewLogger(t)),
		Tracer:           noop.NewTracerProvider().Tracer(""),
		Handler:          h,
		OnChannelTooLong: func(int64) {},
		OnTooLong:        func() {},
		Storage:          storage,
		Hasher:           newMemAccessHasher(),
		SelfID:           selfID,
		DiffLimit:        diffLimitUser,
		WorkGroup:        &errgroup.Group{},
	})
}

// TestHandleAffectedAdvancesPts ensures a contiguous affected pts advances the
// common sequence without dispatching anything to the handler.
func TestHandleAffectedAdvancesPts(t *testing.T) {
	ctx := context.Background()
	h := &recordingHandler{}
	s := newAffectedTestState(t, h, 100)

	require.NoError(t, s.handleAffected(ctx, 0, 101, 1))
	require.Equal(t, 101, s.pts.State(), "pts must advance to the affected value")
	require.Empty(t, h.updates(), "affected pts must not dispatch an update")
}

// TestHandleAffectedZeroIgnored ensures a zero affected pts is ignored instead
// of resetting the sequence state to 0.
func TestHandleAffectedZeroIgnored(t *testing.T) {
	ctx := context.Background()
	h := &recordingHandler{}
	s := newAffectedTestState(t, h, 100)

	require.NoError(t, s.handleAffected(ctx, 0, 0, 0))
	require.Equal(t, 100, s.pts.State(), "zero affected pts must leave state untouched")
}

// TestHandleAffectedFillsGap reproduces issue #1382: an edit update arrives with
// a pts gap (the intermediate pts came from a self-initiated read whose
// affectedMessages result was dropped). The edit is postponed until the affected
// pts is applied, after which both are delivered in order.
func TestHandleAffectedFillsGap(t *testing.T) {
	ctx := context.Background()
	h := &recordingHandler{}
	s := newAffectedTestState(t, h, 4416)

	edit := &tg.UpdateEditMessage{
		Message:  &tg.Message{ID: 777, Message: "edited"},
		Pts:      4418,
		PtsCount: 1,
	}
	// Edit creates a gap [4416, 4417): pts 4417 (the read) is missing.
	require.NoError(t, s.handlePts(ctx, edit.Pts, edit.PtsCount, edit, entities{}))
	require.Empty(t, h.updates(), "edit must be postponed while the gap is open")
	require.Equal(t, 4416, s.pts.State())

	// The dropped read's affectedMessages pts fills the gap.
	require.NoError(t, s.handleAffected(ctx, 0, 4417, 1))

	require.Equal(t, 4418, s.pts.State(), "pts must advance past the edit")
	got := h.updates()
	require.Len(t, got, 1, "exactly the edit must be dispatched (not the marker)")
	require.IsType(t, &tg.UpdateEditMessage{}, got[0])
}

// TestHandleAffectedUntrackedChannelIgnored ensures an affected pts for a channel
// that is not tracked is dropped rather than spuriously subscribing.
func TestHandleAffectedUntrackedChannelIgnored(t *testing.T) {
	ctx := context.Background()
	h := &recordingHandler{}
	s := newAffectedTestState(t, h, 100)

	require.NoError(t, s.handleAffected(ctx, 555, 10, 1))
	require.Empty(t, h.updates())
	require.Equal(t, 100, s.pts.State(), "common pts must be untouched by a channel affected pts")
}
