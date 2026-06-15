package messages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func TestGetHistoryResolve(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	raw := tg.NewClient(mock)

	resolved := &tg.InputPeerUser{UserID: 10, AccessHash: 42}

	// Default resolver issues contacts.resolveUsername for the domain.
	mock.ExpectCall(&tg.ContactsResolveUsernameRequest{Username: "telegram"}).
		ThenResult(&tg.ContactsResolvedPeer{
			Peer:  &tg.PeerUser{UserID: 10},
			Users: []tg.UserClass{&tg.User{ID: 10, AccessHash: 42}},
		})

	// The resolved peer must reach messages.getHistory.
	mock.ExpectCall(&tg.MessagesGetHistoryRequest{
		Peer:  resolved,
		Limit: 1,
	}).ThenResult(messagesClass(generateMessages(1), 1))

	res, err := NewQueryBuilder(raw).
		GetHistoryResolve("telegram").
		Query(ctx, Request{Limit: 1})
	require.NoError(t, err)
	require.Equal(t, 1, res.(*tg.MessagesChannelMessages).Count)
}

type stubResolver struct {
	peer tg.InputPeerClass
}

func (s stubResolver) ResolveDomain(context.Context, string) (tg.InputPeerClass, error) {
	return s.peer, nil
}

func (s stubResolver) ResolvePhone(context.Context, string) (tg.InputPeerClass, error) {
	return s.peer, nil
}

var _ peer.Resolver = stubResolver{}

func TestGetHistoryResolveWithResolver(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	raw := tg.NewClient(mock)

	resolved := &tg.InputPeerUser{UserID: 1, AccessHash: 1}

	// Custom resolver short-circuits, so no contacts.resolveUsername is expected.
	mock.ExpectCall(&tg.MessagesGetHistoryRequest{
		Peer:  resolved,
		Limit: 1,
	}).ThenResult(messagesClass(generateMessages(1), 1))

	res, err := NewQueryBuilder(raw).
		WithResolver(stubResolver{peer: resolved}).
		GetHistoryResolve("telegram").
		Query(ctx, Request{Limit: 1})
	require.NoError(t, err)
	require.Equal(t, 1, res.(*tg.MessagesChannelMessages).Count)
}
