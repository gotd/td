package updates

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// trackAPI records channel difference requests.
type trackAPI struct {
	mu       sync.Mutex
	requests map[int64][]int // channel ID -> requested pts
}

func (a *trackAPI) UpdatesGetState(ctx context.Context) (*tg.UpdatesState, error) {
	return &tg.UpdatesState{Pts: 1, Qts: 1, Seq: 1, Date: 1}, nil
}

func (a *trackAPI) UpdatesGetDifference(
	ctx context.Context, req *tg.UpdatesGetDifferenceRequest,
) (tg.UpdatesDifferenceClass, error) {
	return &tg.UpdatesDifferenceEmpty{Date: req.Date, Seq: 1}, nil
}

func (a *trackAPI) UpdatesGetChannelDifference(
	ctx context.Context, req *tg.UpdatesGetChannelDifferenceRequest,
) (tg.UpdatesChannelDifferenceClass, error) {
	ch := req.Channel.(*tg.InputChannel)
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.requests == nil {
		a.requests = map[int64][]int{}
	}
	a.requests[ch.ChannelID] = append(a.requests[ch.ChannelID], req.Pts)
	return &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: req.Pts, Timeout: 1000}, nil
}

func (a *trackAPI) ptsFor(channelID int64) []int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]int(nil), a.requests[channelID]...)
}

type trackTest struct {
	m       *Manager
	api     *trackAPI
	storage *memStorage
	hasher  *memAccessHasher
}

const trackUserID = 1

func newTrackTest(t *testing.T) *trackTest {
	t.Helper()
	ctx := context.Background()
	storage := newMemStorage()
	require.NoError(t, storage.SetState(ctx, trackUserID, State{Pts: 1, Qts: 1, Seq: 1, Date: 1}))
	hasher := newMemAccessHasher()
	return &trackTest{
		m: newTestManager(t, Config{
			Storage:      storage,
			AccessHasher: hasher,
		}),
		api:     &trackAPI{},
		storage: storage,
		hasher:  hasher,
	}
}

func (tt *trackTest) run(t *testing.T, opt AuthOptions) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- tt.m.Run(ctx, tt.api, trackUserID, opt) }()
	t.Cleanup(func() {
		cancel()
		<-done
	})
}

func (tt *trackTest) waitPts(t *testing.T, channelID int64, pts ...int) {
	t.Helper()
	require.Eventually(t, func() bool {
		got := tt.api.ptsFor(channelID)
		return len(got) >= len(pts) && equalInts(got[:len(pts)], pts)
	}, 5*time.Second, time.Millisecond, "channel %d difference requests: %v", channelID, tt.api.ptsFor(channelID))
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestManagerTrackChannel(t *testing.T) {
	ctx := context.Background()

	t.Run("BeforeRun", func(t *testing.T) {
		tt := newTrackTest(t)
		require.NoError(t, tt.m.TrackChannel(ctx, 42, 4242, 100))
		tt.run(t, AuthOptions{})
		tt.waitPts(t, 42, 100)

		hash, found, err := tt.hasher.GetChannelAccessHash(ctx, trackUserID, 42)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, int64(4242), hash)
	})
	t.Run("OnStart", func(t *testing.T) {
		tt := newTrackTest(t)
		tt.run(t, AuthOptions{OnStart: func(ctx context.Context) {
			require.NoError(t, tt.m.TrackChannel(ctx, 42, 4242, 100))
		}})
		tt.waitPts(t, 42, 100)
	})
	t.Run("Running", func(t *testing.T) {
		tt := newTrackTest(t)
		started := make(chan struct{})
		tt.run(t, AuthOptions{OnStart: func(context.Context) { close(started) }})
		<-started

		require.Eventually(t, func() bool {
			// Wait until manager is started.
			return tt.m.TrackChannel(ctx, 42, 4242, 100) == nil && len(tt.api.ptsFor(42)) > 0
		}, 5*time.Second, time.Millisecond)
		tt.waitPts(t, 42, 100)

		// Already tracked: no-op.
		require.NoError(t, tt.m.TrackChannel(ctx, 42, 4242, 500))
		require.Equal(t, []int{100}, tt.api.ptsFor(42))

		// No stored pts and no pts.
		require.Error(t, tt.m.TrackChannel(ctx, 43, 4343, 0))
	})
	t.Run("StoredPts", func(t *testing.T) {
		// Stored pts without access hash (OnLoadChannelStateFailed case):
		// tracking starts from stored pts to fetch missed updates.
		tt := newTrackTest(t)
		require.NoError(t, tt.storage.SetChannelPts(ctx, trackUserID, 42, 50))
		require.NoError(t, tt.m.TrackChannel(ctx, 42, 4242, 100))
		tt.run(t, AuthOptions{})
		tt.waitPts(t, 42, 50)
	})
}

func TestManagerChannelTooLongUntracked(t *testing.T) {
	ctx := context.Background()
	tt := newTrackTest(t)
	// Stored pts without access hash: channel is skipped on start.
	require.NoError(t, tt.storage.SetChannelPts(ctx, trackUserID, 42, 50))
	started := make(chan struct{})
	tt.run(t, AuthOptions{OnStart: func(context.Context) { close(started) }})
	<-started

	long := &tg.UpdateChannelTooLong{ChannelID: 42}
	long.SetPts(1000)
	require.NoError(t, tt.m.Handle(ctx, &tg.Updates{
		Updates: []tg.UpdateClass{long},
		Chats:   []tg.ChatClass{&tg.Channel{ID: 42, AccessHash: 4242}},
		Date:    1,
		Seq:     0,
	}))
	tt.waitPts(t, 42, 50)

	hash, found, err := tt.hasher.GetChannelAccessHash(ctx, trackUserID, 42)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(4242), hash)
}

func TestManagerResetDropsPendingTracks(t *testing.T) {
	m := New(Config{Handler: telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error { return nil })})
	// Not started: the request is kept until Run.
	require.NoError(t, m.TrackChannel(context.Background(), 42, 1, 100))
	require.Len(t, m.pendingTracks, 1)

	// Pending tracks of the previous user must not be applied to the next one.
	m.Reset()
	require.Empty(t, m.pendingTracks)
}
