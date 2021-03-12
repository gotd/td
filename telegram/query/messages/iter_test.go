package messages

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/internal/rpcmock"
	"github.com/gotd/td/tg"
)

func generateMessages(count int) []tg.MessageClass {
	r := make([]tg.MessageClass, 0, count)

	for i := 0; i < count; i++ {
		r = append(r, &tg.Message{
			Flags:   0,
			ID:      i,
			PeerID:  &tg.PeerUser{UserID: 10},
			Message: strconv.Itoa(i),
		})
	}

	return r
}

func messagesClass(r []tg.MessageClass, count int) tg.MessagesMessagesClass {
	return &tg.MessagesChannelMessages{
		Messages: r,
		Count:    count,
	}
}

func TestIterator(t *testing.T) {
	ctx := context.Background()
	mock := rpcmock.NewMock(t, require.New(t))
	limit := 10
	totalMessages := 3 * limit
	expected := generateMessages(totalMessages)
	raw := tg.NewClient(mock)

	mock.ExpectCall(&tg.MessagesSearchRequest{
		Q:        "query",
		Peer:     &tg.InputPeerSelf{},
		OffsetID: 0,
		FromID:   &tg.InputPeerEmpty{},
		Filter:   &tg.InputMessagesFilterEmpty{},
		Limit:    limit,
	}).ThenResult(messagesClass(expected[2*limit:3*limit], totalMessages))
	mock.ExpectCall(&tg.MessagesSearchRequest{
		Q:        "query",
		Peer:     &tg.InputPeerSelf{},
		OffsetID: 20,
		FromID:   &tg.InputPeerEmpty{},
		Filter:   &tg.InputMessagesFilterEmpty{},
		Limit:    limit,
	}).ThenResult(messagesClass(expected[limit:2*limit], totalMessages))
	mock.ExpectCall(&tg.MessagesSearchRequest{
		Q:        "query",
		Peer:     &tg.InputPeerSelf{},
		OffsetID: 10,
		FromID:   &tg.InputPeerEmpty{},
		Filter:   &tg.InputMessagesFilterEmpty{},
		Limit:    limit,
	}).ThenResult(messagesClass(expected[:limit], totalMessages))
	mock.ExpectCall(&tg.MessagesSearchRequest{
		Q:        "query",
		Peer:     &tg.InputPeerSelf{},
		OffsetID: 0,
		FromID:   &tg.InputPeerEmpty{},
		Filter:   &tg.InputMessagesFilterEmpty{},
		Limit:    limit,
	}).ThenResult(messagesClass(expected[:0], totalMessages))

	iter := NewQueryBuilder(raw).Search(&tg.InputPeerSelf{}).
		Filter(&tg.InputMessagesFilterEmpty{}).
		Q("query").BatchSize(10).Iter()
	i := 0
	for iter.Next(ctx) {
		mock.Equal(expected[len(expected)-i-1], iter.Value().Msg)
		i++
	}
	mock.NoError(iter.Err())
	mock.Equal(totalMessages, i)

	total, err := iter.Total(ctx)
	mock.NoError(err)
	mock.Equal(totalMessages, total)

	mock.ExpectCall(&tg.MessagesSearchRequest{
		Q:        "query",
		Peer:     &tg.InputPeerSelf{},
		OffsetID: 0,
		FromID:   &tg.InputPeerEmpty{},
		Filter:   &tg.InputMessagesFilterEmpty{},
		Limit:    1,
	}).ThenResult(messagesClass(expected[:0], totalMessages))
	total, err = iter.FetchTotal(ctx)
	mock.NoError(err)
	mock.Equal(totalMessages, total)
}
