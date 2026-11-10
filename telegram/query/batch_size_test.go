package query

import (
	"context"
	"sync"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/query/channels/participants"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
)

// pagingServer serves total dialogs, messages or participants, honouring the request limit.
type pagingServer struct {
	total int

	mux    sync.Mutex
	limits []int
}

func (s *pagingServer) Limits() []int {
	s.mux.Lock()
	defer s.mux.Unlock()
	return append([]int(nil), s.limits...)
}

func reencode(res bin.Encoder, out bin.Decoder) error {
	var b bin.Buffer
	if err := res.Encode(&b); err != nil {
		return err
	}
	return out.Decode(&b)
}

func (s *pagingServer) Invoke(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	switch r := in.(type) {
	case *tg.MessagesGetDialogsRequest:
		s.limits = append(s.limits, r.Limit)
		// Dialogs have top message IDs total..1, descending.
		start := s.total
		if r.OffsetID != 0 {
			start = r.OffsetID - 1
		}
		res := &tg.MessagesDialogsSlice{Count: s.total}
		for id := start; id >= 1 && len(res.Dialogs) < r.Limit; id-- {
			p := &tg.PeerUser{UserID: int64(id)}
			res.Dialogs = append(res.Dialogs, &tg.Dialog{Peer: p, TopMessage: id})
			res.Messages = append(res.Messages, &tg.Message{ID: id, PeerID: p, Date: id})
			res.Users = append(res.Users, &tg.User{ID: int64(id), AccessHash: 1})
		}
		return reencode(res, out)
	case *tg.MessagesGetHistoryRequest:
		s.limits = append(s.limits, r.Limit)
		start := s.total
		if r.OffsetID != 0 {
			start = r.OffsetID - 1
		}
		res := &tg.MessagesMessagesSlice{Count: s.total}
		for id := start; id >= 1 && len(res.Messages) < r.Limit; id-- {
			res.Messages = append(res.Messages, &tg.Message{ID: id, PeerID: &tg.PeerUser{UserID: 1}, Date: id})
		}
		return reencode(res, out)
	case *tg.ChannelsGetParticipantsRequest:
		s.limits = append(s.limits, r.Limit)
		res := &tg.ChannelsChannelParticipants{Count: s.total}
		for id := r.Offset + 1; id <= s.total && len(res.Participants) < r.Limit; id++ {
			res.Participants = append(res.Participants, &tg.ChannelParticipant{UserID: int64(id)})
			res.Users = append(res.Users, &tg.User{ID: int64(id), AccessHash: 1})
		}
		return reencode(res, out)
	}
	return errors.Errorf("unexpected %T", in)
}

func TestIteratorDefaultBatchSize(t *testing.T) {
	ctx := context.Background()
	const (
		total    = 30
		explicit = 7
	)

	// iterate runs the iterator with the given batch size (0 means default) and
	// returns the number of elements.
	type iterate func(q *Query, batchSize int) (int, error)
	for _, tt := range []struct {
		name  string
		limit int
		run   iterate
	}{
		{
			name:  "Dialogs",
			limit: 100,
			run: func(q *Query, batchSize int) (n int, err error) {
				b := q.GetDialogs()
				if batchSize > 0 {
					b = b.BatchSize(batchSize)
				}
				err = b.ForEach(ctx, func(context.Context, dialogs.Elem) error { n++; return nil })
				return n, err
			},
		},
		{
			name:  "History",
			limit: 100,
			run: func(q *Query, batchSize int) (n int, err error) {
				b := q.Messages().GetHistory(&tg.InputPeerSelf{})
				if batchSize > 0 {
					b = b.BatchSize(batchSize)
				}
				err = b.ForEach(ctx, func(context.Context, messages.Elem) error { n++; return nil })
				return n, err
			},
		},
		{
			name:  "Participants",
			limit: 200,
			run: func(q *Query, batchSize int) (n int, err error) {
				b := q.GetParticipants(&tg.InputChannel{ChannelID: 1})
				if batchSize > 0 {
					b = b.BatchSize(batchSize)
				}
				err = b.ForEach(ctx, func(context.Context, participants.Elem) error { n++; return nil })
				return n, err
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("Default", func(t *testing.T) {
				s := &pagingServer{total: total}
				n, err := tt.run(NewQuery(tg.NewClient(s)), 0)
				require.NoError(t, err)
				require.Equal(t, total, n)

				limits := s.Limits()
				require.NotEmpty(t, limits)
				require.LessOrEqual(t, len(limits), 2, "default iteration must page, not fetch one element per RPC")
				for _, l := range limits {
					require.Equal(t, tt.limit, l)
				}
			})
			t.Run("Explicit", func(t *testing.T) {
				s := &pagingServer{total: total}
				n, err := tt.run(NewQuery(tg.NewClient(s)), explicit)
				require.NoError(t, err)
				require.Equal(t, total, n)

				limits := s.Limits()
				require.GreaterOrEqual(t, len(limits), total/explicit)
				for _, l := range limits {
					require.Equal(t, explicit, l, "BatchSize must override the default")
				}
			})
		})
	}
}
