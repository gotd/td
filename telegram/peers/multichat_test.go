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

	SetReactions(ctx context.Context, r ...string) error
	DisableReactions(ctx context.Context) error

	KickUser(ctx context.Context, member tg.InputUserClass, revokeHistory bool) error
	EditRights(ctx context.Context, options ParticipantRights) error
}

var _ = []multiChat{
	Chat{},
	Channel{},
}

func TestReactions(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	req := func(p Peer, r ...string) *tgmock.RequestBuilder {
		return mock.ExpectCall(&tg.MessagesSetChatAvailableReactionsRequest{
			Peer:               p.InputPeer(),
			AvailableReactions: r,
		})
	}
	reactions := []string{"a", "b", "c"}
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

func TestEditRights(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	rights := tg.ChatBannedRights{
		SendInline: true,
	}
	rights.SetFlags()
	req := func(p Peer) *tgmock.RequestBuilder {
		return mock.ExpectCall(&tg.MessagesEditChatDefaultBannedRightsRequest{
			Peer:         p.InputPeer(),
			BannedRights: rights,
		})
	}
	for _, p := range []multiChat{
		m.Chat(getTestChat()),
		m.Channel(getTestChannel()),
	} {
		req(p).ThenRPCErr(getTestError())
		a.Error(p.EditRights(ctx, ParticipantRights{DenySendInline: true}))
		req(p).ThenResult(&tg.Updates{})
		a.NoError(p.EditRights(ctx, ParticipantRights{DenySendInline: true}))
	}
}
