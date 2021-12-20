package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestChatGetters(t *testing.T) {
	a := require.New(t)
	u := Chat{
		raw: &tg.Chat{
			Creator:             true,
			Kicked:              true,
			Left:                true,
			Deactivated:         true,
			CallActive:          true,
			CallNotEmpty:        true,
			Noforwards:          true,
			ID:                  10,
			Title:               "Title",
			ParticipantsCount:   10,
			Date:                10,
			Version:             1,
			AdminRights:         tg.ChatAdminRights{AddAdmins: true},
			DefaultBannedRights: tg.ChatBannedRights{EmbedLinks: true},
		},
	}
	u.raw.SetFlags()
	a.Equal(u.raw, u.Raw())
	a.True(u.TDLibPeerID().IsChat())

	a.Equal("Title", u.VisibleName())
	a.Equal(&tg.InputPeerChat{ChatID: u.raw.ID}, u.InputPeer())
	a.False(u.Verified())
	a.False(u.Scam())
	a.False(u.Fake())
	a.Equal(u.raw.GetID(), u.ID())

	a.Equal(u.raw.Creator, u.Creator())
	a.Equal(u.raw.Kicked, u.Kicked())
	a.Equal(u.raw.Left, u.Left())
	a.Equal(u.raw.Deactivated, u.Deactivated())
	a.Equal(u.raw.CallActive, u.CallActive())
	a.Equal(u.raw.CallNotEmpty, u.CallNotEmpty())
	a.Equal(u.raw.Noforwards, u.NoForwards())
	{
		_, ok := u.ToSupergroup()
		a.False(ok)
	}
	{
		_, ok := u.ToBroadcast()
		a.False(ok)
	}
}

func TestChat_Leave(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	ch := m.Chat(getTestChat())

	mock.ExpectCall(&tg.MessagesDeleteChatUserRequest{
		RevokeHistory: false,
		ChatID:        ch.ID(),
		UserID:        &tg.InputUserSelf{},
	}).ThenRPCErr(getTestError())
	a.Error(ch.Leave(ctx))

	mock.ExpectCall(&tg.MessagesDeleteChatUserRequest{
		RevokeHistory: false,
		ChatID:        ch.ID(),
		UserID:        &tg.InputUserSelf{},
	}).ThenResult(&tg.Updates{})
	a.NoError(ch.Leave(ctx))
}

func TestChat_SetTitle(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	title := "title"
	ch := m.Chat(getTestChat())

	mock.ExpectCall(&tg.MessagesEditChatTitleRequest{
		ChatID: ch.ID(),
		Title:  title,
	}).ThenRPCErr(getTestError())
	a.Error(ch.SetTitle(ctx, title))

	mock.ExpectCall(&tg.MessagesEditChatTitleRequest{
		ChatID: ch.ID(),
		Title:  title,
	}).ThenResult(&tg.Updates{})
	a.NoError(ch.SetTitle(ctx, title))
}

func TestChat_SetDescription(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	about := "about"
	ch := m.Chat(getTestChat())

	mock.ExpectCall(&tg.MessagesEditChatAboutRequest{
		Peer:  ch.InputPeer(),
		About: about,
	}).ThenRPCErr(getTestError())
	a.Error(ch.SetDescription(ctx, about))

	mock.ExpectCall(&tg.MessagesEditChatAboutRequest{
		Peer:  ch.InputPeer(),
		About: about,
	}).ThenTrue()
	a.NoError(ch.SetDescription(ctx, about))
}

func TestChat_LeaveAndDelete(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	ch := m.Chat(getTestChat())

	mock.ExpectCall(&tg.MessagesDeleteChatUserRequest{
		RevokeHistory: true,
		ChatID:        ch.ID(),
		UserID:        &tg.InputUserSelf{},
	}).ThenRPCErr(getTestError())
	a.Error(ch.LeaveAndDelete(ctx))

	mock.ExpectCall(&tg.MessagesDeleteChatUserRequest{
		RevokeHistory: true,
		ChatID:        ch.ID(),
		UserID:        &tg.InputUserSelf{},
	}).ThenResult(&tg.Updates{})
	a.NoError(ch.LeaveAndDelete(ctx))
}

func TestChat_ActualChat(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	_, m := testManager(t)

	ch := m.Chat(getTestChat())
	_, ok, err := ch.ActualChat(ctx)
	a.NoError(err)
	a.False(ok)

	newChat := m.Channel(getTestChannel())
	a.NoError(m.applyChats(ctx, newChat.raw))
	ch.raw.SetMigratedTo(newChat.InputChannel())

	actual, ok, err := ch.ActualChat(ctx)
	a.NoError(err)
	a.True(ok)
	a.Equal(newChat.ID(), actual.ID())
}
