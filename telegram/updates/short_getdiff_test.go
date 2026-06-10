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

// countDiffAPI counts UpdatesGetDifference calls and always returns an empty
// difference, so no messages are dispatched through the difference path.
type countDiffAPI struct {
	diffCalls int
}

func (a *countDiffAPI) UpdatesGetState(ctx context.Context) (*tg.UpdatesState, error) {
	return &tg.UpdatesState{}, nil
}

func (a *countDiffAPI) UpdatesGetDifference(
	ctx context.Context, request *tg.UpdatesGetDifferenceRequest,
) (tg.UpdatesDifferenceClass, error) {
	a.diffCalls++
	return &tg.UpdatesDifferenceEmpty{}, nil
}

func (a *countDiffAPI) UpdatesGetChannelDifference(
	ctx context.Context, request *tg.UpdatesGetChannelDifferenceRequest,
) (tg.UpdatesChannelDifferenceClass, error) {
	return &tg.UpdatesChannelDifferenceEmpty{}, nil
}

func newShortTestState(t *testing.T, api API, handler telegram.UpdateHandler) *internalState {
	t.Helper()
	ctx := context.Background()
	const selfID = 123
	storage := newMemStorage()
	require.NoError(t, storage.SetState(ctx, selfID, State{}))
	return newState(ctx, stateConfig{
		RawClient:        api,
		Logger:           zaptest.NewLogger(t),
		Tracer:           noop.NewTracerProvider().Tracer(""),
		Handler:          handler,
		OnChannelTooLong: func(int64) {},
		OnTooLong:        func() {},
		Storage:          storage,
		Hasher:           newMemAccessHasher(),
		UserHasher:       newMemUserAccessHasher(),
		SelfID:           selfID,
		DiffLimit:        diffLimitUser,
		WorkGroup:        &errgroup.Group{},
	})
}

func TestShortMessageUnknownPeerForcesDifference(t *testing.T) {
	ctx := context.Background()
	api := &countDiffAPI{}
	var dispatched int
	handler := telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
		dispatched++
		return nil
	})
	s := newShortTestState(t, api, handler)

	err := s.handleUpdates(ctx, &tg.UpdateShortMessage{UserID: 555, ID: 1, Pts: 1, PtsCount: 1})
	require.NoError(t, err)
	require.Equal(t, 1, api.diffCalls, "getDifference must be forced for unknown peer")
	require.Equal(t, 0, dispatched, "short message must not be dispatched directly")
}

func TestShortMessageKnownPeerAppliesDirectly(t *testing.T) {
	ctx := context.Background()
	api := &countDiffAPI{}
	var dispatched int
	handler := telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
		dispatched++
		return nil
	})
	s := newShortTestState(t, api, handler)

	// Seed the sender as known through the normal apply path: a prior *tg.Updates
	// carrying the full (non-min, non-zero hash) user marks it known.
	require.NoError(t, s.handleUpdates(ctx, &tg.Updates{
		Users: []tg.UserClass{&tg.User{ID: 555, AccessHash: 7777}},
	}))

	err := s.handleUpdates(ctx, &tg.UpdateShortMessage{UserID: 555, ID: 1, Pts: 1, PtsCount: 1})
	require.NoError(t, err)
	require.Equal(t, 0, api.diffCalls, "no difference for known peer")
	require.Equal(t, 1, dispatched, "short message must be dispatched directly")
}

// TestShortMessageMinUserNotKnown encodes the min rule: a user observed ONLY as
// min (or with a zero access hash) does NOT count as known, so a short message
// from it still forces getDifference.
func TestShortMessageMinUserNotKnown(t *testing.T) {
	ctx := context.Background()

	for _, tt := range []struct {
		name string
		user *tg.User
	}{
		{"min user", &tg.User{ID: 555, Min: true, AccessHash: 7777}},
		{"zero access hash", &tg.User{ID: 555, AccessHash: 0}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			api := &countDiffAPI{}
			var dispatched int
			handler := telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
				dispatched++
				return nil
			})
			s := newShortTestState(t, api, handler)

			// Feed the min/zero-hash user through the normal apply path. It must
			// NOT be recorded as known.
			require.NoError(t, s.handleUpdates(ctx, &tg.Updates{
				Users: []tg.UserClass{tt.user},
			}))
			_, known, _ := s.userHasher.GetUserAccessHash(ctx, s.selfID, 555)
			require.False(t, known, "min/zero-hash user must not be recorded as known")

			err := s.handleUpdates(ctx, &tg.UpdateShortMessage{UserID: 555, ID: 1, Pts: 1, PtsCount: 1})
			require.NoError(t, err)
			require.Equal(t, 1, api.diffCalls, "min/zero-hash sender must still force getDifference")
			require.Equal(t, 0, dispatched, "short message must not be dispatched directly")
		})
	}
}

func TestShortChatMessageUnknownPeerForcesDifference(t *testing.T) {
	ctx := context.Background()
	api := &countDiffAPI{}
	var dispatched int
	handler := telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
		dispatched++
		return nil
	})
	s := newShortTestState(t, api, handler)

	err := s.handleUpdates(ctx, &tg.UpdateShortChatMessage{FromID: 555, ChatID: 42, ID: 1, Pts: 1, PtsCount: 1})
	require.NoError(t, err)
	require.Equal(t, 1, api.diffCalls, "getDifference must be forced for unknown sender of a chat message")
	require.Equal(t, 0, dispatched, "short chat message must not be dispatched directly")
}

// oneDiffAPI returns a single configured difference for every getDifference call.
type oneDiffAPI struct {
	diff  tg.UpdatesDifferenceClass
	calls int
}

func (a *oneDiffAPI) UpdatesGetState(ctx context.Context) (*tg.UpdatesState, error) {
	return &tg.UpdatesState{}, nil
}

func (a *oneDiffAPI) UpdatesGetDifference(
	ctx context.Context, request *tg.UpdatesGetDifferenceRequest,
) (tg.UpdatesDifferenceClass, error) {
	a.calls++
	return a.diff, nil
}

func (a *oneDiffAPI) UpdatesGetChannelDifference(
	ctx context.Context, request *tg.UpdatesGetChannelDifferenceRequest,
) (tg.UpdatesChannelDifferenceClass, error) {
	return &tg.UpdatesChannelDifferenceEmpty{}, nil
}

// TestShortMessageRecoveredDifferenceCarriesSender locks the end-to-end invariant:
// forcing getDifference for an unknown sender re-delivers the message together with
// the sender entity (access hash) in the SAME envelope, so a downstream
// entity-ingesting handler learns the access hash before the message is processed.
// Afterwards, that peer is recorded as known.
func TestShortMessageRecoveredDifferenceCarriesSender(t *testing.T) {
	ctx := context.Background()
	user := &tg.User{ID: 555, AccessHash: 7777}
	msg := &tg.Message{
		ID:      10,
		PeerID:  &tg.PeerUser{UserID: 555},
		FromID:  &tg.PeerUser{UserID: 555},
		Message: "hi",
	}
	api := &oneDiffAPI{diff: &tg.UpdatesDifference{
		NewMessages: []tg.MessageClass{msg},
		Users:       []tg.UserClass{user},
		State:       tg.UpdatesState{Pts: 1, Date: 1},
	}}

	var got *tg.Updates
	handler := telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
		if up, ok := u.(*tg.Updates); ok && len(up.Updates) > 0 {
			got = up
		}
		return nil
	})
	s := newShortTestState(t, api, handler)

	err := s.handleUpdates(ctx, &tg.UpdateShortMessage{UserID: 555, ID: 10, Pts: 1, PtsCount: 1})
	require.NoError(t, err)
	require.Equal(t, 1, api.calls, "getDifference must be forced once for the unknown sender")
	require.NotNil(t, got, "recovered message must be dispatched")

	require.Len(t, got.Users, 1, "recovered envelope must carry the sender entity")
	gotUser, ok := got.Users[0].(*tg.User)
	require.True(t, ok)
	require.Equal(t, int64(555), gotUser.ID)
	require.Equal(t, int64(7777), gotUser.AccessHash, "sender access hash present before message dispatch")

	// The recovered sender is now known: a subsequent short message from it
	// applies directly without another getDifference.
	_, known, _ := s.userHasher.GetUserAccessHash(ctx, s.selfID, 555)
	require.True(t, known, "recovered sender must be recorded as known")
}
