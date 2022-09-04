package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

type multiChat interface {
	Peer
	Creator() bool
	Left() bool
	NoForwards() bool
	CallActive() bool
	CallNotEmpty() bool
	ParticipantsCount() int
	AdminRights() (tg.ChatAdminRights, bool)
	DefaultBannedRights() (tg.ChatBannedRights, bool)

	Leave(ctx context.Context) error
	SetTitle(ctx context.Context, title string) error
	SetDescription(ctx context.Context, about string) error

	InviteLinks() InviteLinks
	ToBroadcast() (Broadcast, bool)
	IsBroadcast() bool
	ToSupergroup() (Supergroup, bool)
	IsSupergroup() bool

	SetReactions(ctx context.Context, r ...tg.ReactionClass) error
	DisableReactions(ctx context.Context) error
}

var _ = []multiChat{
	Chat{},
	Channel{},
}

func TestReactions(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	req := func(p Peer, r ...tg.ReactionClass) *tgmock.RequestBuilder {
		var reactions tg.ChatReactionsClass = &tg.ChatReactionsSome{Reactions: r}
		if len(r) == 0 {
			reactions = &tg.ChatReactionsNone{}
		}
		return mock.ExpectCall(&tg.MessagesSetChatAvailableReactionsRequest{
			Peer:               p.InputPeer(),
			AvailableReactions: reactions,
		})
	}
	reactions := []tg.ReactionClass{
		&tg.ReactionEmoji{Emoticon: "👍"},
		&tg.ReactionEmoji{Emoticon: "A"},
	}
	for _, p := range []multiChat{
		m.Chat(getTestChat()),
		m.Channel(getTestChannel()),
	} {
		req(p, reactions...).ThenRPCErr(getTestError())
		a.Error(p.SetReactions(ctx, reactions...))
		req(p, reactions...).ThenResult(&tg.Updates{})
		a.NoError(p.SetReactions(ctx, reactions...))

		req(p).ThenRPCErr(getTestError())
		a.Error(p.DisableReactions(ctx))
		req(p).ThenResult(&tg.Updates{})
		a.NoError(p.DisableReactions(ctx))
	}
}
