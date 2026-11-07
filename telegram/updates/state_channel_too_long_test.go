package updates

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// tooLongAPI returns channelDifferenceTooLong with the latest messages on the
// first call, and an empty difference afterwards.
type tooLongAPI struct {
	gateChannelDiffAPI

	mu        sync.Mutex
	final     bool
	remotePts int // dialog pts, 5000 if zero
	requests  []*tg.UpdatesGetChannelDifferenceRequest
}

func (a *tooLongAPI) UpdatesGetChannelDifference(
	ctx context.Context, req *tg.UpdatesGetChannelDifferenceRequest,
) (tg.UpdatesChannelDifferenceClass, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.requests = append(a.requests, req)
	if len(a.requests) > 1 {
		return &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: req.Pts}, nil
	}

	d := &tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 42}, TopMessage: 1003}
	pts := a.remotePts
	if pts == 0 {
		pts = 5000
	}
	d.SetPts(pts)
	peer := &tg.PeerChannel{ChannelID: 42}
	return &tg.UpdatesChannelDifferenceTooLong{
		Final:  a.final,
		Dialog: d,
		Messages: []tg.MessageClass{
			&tg.Message{ID: 1001, PeerID: peer, Message: "a"},
			&tg.Message{ID: 1002, PeerID: peer, Message: "b"},
			&tg.Message{ID: 1003, PeerID: peer, Message: "c"},
		},
		Chats: []tg.ChatClass{&tg.Channel{ID: 42, AccessHash: 1}},
		Users: []tg.UserClass{&tg.User{ID: 7, AccessHash: 7}},
	}, nil
}

func (a *tooLongAPI) requestCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.requests)
}

type tooLongResult struct {
	handled int // updates passed to Handler
	tooLong []int64
	gaps    []ChannelGap
}

// gapMessages returns IDs of messages reported with gaps and clears them, so
// gaps can be compared without messages and entities.
func (r *tooLongResult) gapMessages(t *testing.T) []int {
	t.Helper()
	var ids []int
	for i := range r.gaps {
		require.NotEmpty(t, r.gaps[i].Users, "entities must be passed")
		require.NotEmpty(t, r.gaps[i].Chats, "entities must be passed")
		for _, m := range r.gaps[i].Messages {
			ids = append(ids, m.GetID())
		}
		r.gaps[i].Messages, r.gaps[i].Users, r.gaps[i].Chats = nil, nil, nil
	}
	return ids
}

func newTooLongChannelState(t *testing.T, api API, r *tooLongResult) *channelState {
	t.Helper()
	storage := newMemStorage()
	require.NoError(t, storage.SetState(context.Background(), 0, State{}))
	return newChannelState(channelStateConfig{
		Out:        make(chan tracedUpdate, 10),
		ChannelID:  42,
		AccessHash: 1,
		InitialPts: 100,
		DiffLimit:  diffLimitUser,
		RawClient:  api,
		Storage:    storage,
		Handler: telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error {
			r.handled++
			return nil
		}),
		OnChannelTooLong:      func(id int64) { r.tooLong = append(r.tooLong, id) },
		OnChannelGap:          func(gap ChannelGap) { r.gaps = append(r.gaps, gap) },
		OnChannelInaccessible: func(int64) {},
		RemoveChannel:         make(chan int64, 1),
		Logger:                logzap.New(zaptest.NewLogger(t)),
		Tracer:                noop.NewTracerProvider().Tracer(""),
	})
}

func TestChannelDifferenceTooLong(t *testing.T) {
	ctx := context.Background()
	wantGap := ChannelGap{
		ChannelID:  42,
		AccessHash: 1,
		LocalPts:   100,
		RemotePts:  5000,
		TopMessage: 1003,
	}

	t.Run("Final", func(t *testing.T) {
		api := &tooLongAPI{final: true}
		var r tooLongResult
		s := newTooLongChannelState(t, api, &r)

		require.NoError(t, s.getDifference(ctx, "test"))
		require.Equal(t, []int{1001, 1002, 1003}, r.gapMessages(t), "latest messages must be reported")
		require.Zero(t, r.handled, "latest messages may be old: not passed to Handler")
		require.Equal(t, 5000, s.pts.State())
		require.Equal(t, []int64{42}, r.tooLong)
		require.Equal(t, []ChannelGap{wantGap}, r.gaps)
		require.Equal(t, 1, api.requestCount())

		pts, found, err := s.storage.GetChannelPts(ctx, 0, 42)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, 5000, pts)
	})
	t.Run("NotFinal", func(t *testing.T) {
		api := &tooLongAPI{final: false}
		var r tooLongResult
		s := newTooLongChannelState(t, api, &r)

		require.NoError(t, s.getDifference(ctx, "test"))
		require.Equal(t, []int{1001, 1002, 1003}, r.gapMessages(t))
		require.Equal(t, 2, api.requestCount(), "not final difference must be continued")
		require.Equal(t, 5000, api.requests[1].Pts)
	})
	t.Run("UpdateChannelTooLong", func(t *testing.T) {
		// Gap larger than diffLim must still fetch the latest messages.
		api := &tooLongAPI{final: true}
		var r tooLongResult
		s := newTooLongChannelState(t, api, &r)

		long := &tg.UpdateChannelTooLong{ChannelID: 42}
		long.SetPts(100 + diffLimitUser*10)
		require.NoError(t, s.handleTooLong(ctx, long))
		require.Equal(t, 1, api.requestCount())
		require.Equal(t, []int{1001, 1002, 1003}, r.gapMessages(t))
		require.Equal(t, []ChannelGap{wantGap}, r.gaps)
	})
	t.Run("PtsNotAdvanced", func(t *testing.T) {
		// A non-final difference with stale dialog pts must not be requested
		// again for the same pts.
		api := &tooLongAPI{final: false, remotePts: 50}
		var r tooLongResult
		s := newTooLongChannelState(t, api, &r)

		require.NoError(t, s.getDifference(ctx, "test"))
		require.Equal(t, 2, api.requestCount())
		require.Equal(t, 101, api.requests[1].Pts, "pts must advance")
		r.gapMessages(t)
		require.Equal(t, 101, r.gaps[0].RemotePts)
	})
}
