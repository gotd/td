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
	"github.com/gotd/td/tgerr"
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

// channelPrivateAPI returns CHANNEL_PRIVATE from getChannelDifference,
// simulating the account losing access to a channel.
type channelPrivateAPI struct {
	diffAPI
}

func (a *channelPrivateAPI) UpdatesGetChannelDifference(
	ctx context.Context, request *tg.UpdatesGetChannelDifferenceRequest,
) (tg.UpdatesChannelDifferenceClass, error) {
	return nil, tgerr.New(400, "CHANNEL_PRIVATE")
}

// TestChannelStateInaccessible ensures that when getChannelDifference reports
// CHANNEL_PRIVATE the worker invokes the OnChannelInaccessible hook, signals
// removal, and stops instead of looping with error logs.
func TestChannelStateInaccessible(t *testing.T) {
	ctx := context.Background()

	const channelID = 1750799539

	removeChannel := make(chan int64, 1)
	var inaccessible int64
	state := newChannelState(channelStateConfig{
		Out:        make(chan tracedUpdate, 10),
		ChannelID:  channelID,
		AccessHash: 42,
		RawClient:  &channelPrivateAPI{},
		Storage:    newMemStorage(),
		Handler: telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
			return nil
		}),
		OnChannelTooLong:      func(int64) {},
		OnChannelInaccessible: func(id int64) { inaccessible = id },
		RemoveChannel:         removeChannel,
		Logger:                zaptest.NewLogger(t),
		Tracer:                noop.NewTracerProvider().Tracer(""),
	})

	// Run subscribes via getChannelDifference, hits CHANNEL_PRIVATE, and must
	// return nil (clean stop) rather than spinning on the error.
	require.NoError(t, state.Run(ctx))
	require.Equal(t, int64(channelID), inaccessible, "OnChannelInaccessible must be called with channel ID")
	select {
	case got := <-removeChannel:
		require.Equal(t, int64(channelID), got, "channel ID must be signaled for removal")
	default:
		t.Fatal("expected channel removal signal")
	}

	// done must be closed so Push no longer blocks on the stopped worker.
	require.NoError(t, state.Push(ctx, channelUpdate{}))
}
