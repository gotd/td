package updates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// diffAPI is a stub API that returns the configured differences in order,
// repeating the last one once exhausted.
type diffAPI struct {
	diffs []tg.UpdatesDifferenceClass
	calls int
}

func (a *diffAPI) UpdatesGetState(ctx context.Context) (*tg.UpdatesState, error) {
	return &tg.UpdatesState{}, nil
}

func (a *diffAPI) UpdatesGetDifference(
	ctx context.Context, request *tg.UpdatesGetDifferenceRequest,
) (tg.UpdatesDifferenceClass, error) {
	diff := a.diffs[a.calls]
	if a.calls < len(a.diffs)-1 {
		a.calls++
	}
	return diff, nil
}

func (a *diffAPI) UpdatesGetChannelDifference(
	ctx context.Context, request *tg.UpdatesGetChannelDifferenceRequest,
) (tg.UpdatesChannelDifferenceClass, error) {
	return &tg.UpdatesChannelDifferenceEmpty{}, nil
}

// TestStateOnTooLong ensures the OnTooLong hook is invoked when the manager
// receives updates.differenceTooLong and cannot recover the common state gap.
func TestStateOnTooLong(t *testing.T) {
	ctx := context.Background()

	const selfID = 123
	storage := newMemStorage()
	require.NoError(t, storage.SetState(ctx, selfID, State{}))

	api := &diffAPI{diffs: []tg.UpdatesDifferenceClass{
		&tg.UpdatesDifferenceTooLong{Pts: 100},
		// Stops the recursive getDifference loop.
		&tg.UpdatesDifferenceEmpty{},
	}}

	var tooLong int
	s := newState(ctx, stateConfig{
		RawClient: api,
		Logger:    zaptest.NewLogger(t),
		Tracer:    noop.NewTracerProvider().Tracer(""),
		Handler: telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
			return nil
		}),
		OnChannelTooLong: func(int64) {},
		OnTooLong:        func() { tooLong++ },
		Storage:          storage,
		Hasher:           newMemAccessHasher(),
		SelfID:           selfID,
		DiffLimit:        diffLimitUser,
		WorkGroup:        &errgroup.Group{},
	})

	require.NoError(t, s.getDifference(ctx))
	require.Equal(t, 1, tooLong, "OnTooLong must be called once")
	require.Equal(t, 100, s.pts.State(), "pts must be advanced to the value from differenceTooLong")
}
