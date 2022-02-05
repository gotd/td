package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestChannelGetters(t *testing.T) {
	a := require.New(t)
	u := Channel{
		raw: &tg.Channel{
			Creator:             true,
			Left:                true,
			Broadcast:           true,
			Verified:            true,
			Megagroup:           true,
			Restricted:          true,
			Signatures:          true,
			Min:                 true,
			Scam:                true,
			HasLink:             true,
			HasGeo:              true,
			SlowmodeEnabled:     true,
			CallActive:          true,
			CallNotEmpty:        true,
			Fake:                true,
			Gigagroup:           true,
			Noforwards:          true,
			ID:                  10,
			AccessHash:          10,
			Title:               "Title",
			Username:            "username",
			Date:                10,
			AdminRights:         tg.ChatAdminRights{AddAdmins: true},
			BannedRights:        tg.ChatBannedRights{},
			DefaultBannedRights: tg.ChatBannedRights{EmbedLinks: true},
			ParticipantsCount:   10,
		},
	}
	u.raw.SetFlags()
	a.Equal(u.raw, u.Raw())
	a.True(u.TDLibPeerID().IsChannel())

	a.Equal("Title", u.VisibleName())
	a.Equal(&tg.InputPeerChannel{ChannelID: u.raw.ID, AccessHash: u.raw.AccessHash}, u.InputPeer())
	a.Equal(u.raw.GetID(), u.ID())
	a.Equal(u.raw.Creator, u.Creator())
	a.Equal(u.raw.Left, u.Left())
	a.Equal(u.raw.Verified, u.Verified())
	a.Equal(u.raw.Scam, u.Scam())
	a.Equal(u.raw.HasLink, u.HasLink())
	a.Equal(u.raw.HasGeo, u.HasGeo())
	a.Equal(u.raw.CallActive, u.CallActive())
	a.Equal(u.raw.CallNotEmpty, u.CallNotEmpty())
	a.Equal(u.raw.Fake, u.Fake())
	a.Equal(u.raw.Noforwards, u.NoForwards())
	{
		reasons, ok := u.Restricted()
		a.Equal(u.raw.GetRestricted(), ok)
		a.Equal(u.raw.RestrictionReason, reasons)
	}
	{
		s, ok := u.ToSupergroup()
		a.Equal(s.raw.Megagroup, ok)
		a.Equal(s.raw.Signatures, s.SlowmodeEnabled())
	}
	{
		b, ok := u.ToBroadcast()
		a.Equal(b.raw.Broadcast, ok)
		a.Equal(b.raw.Signatures, b.Signatures())
	}
	{
		v, ok := u.AdminRights()
		v2, ok2 := u.raw.GetAdminRights()
		a.Equal(ok, ok2)
		a.Equal(v2, v)
	}
	{
		v, ok := u.BannedRights()
		v2, ok2 := u.raw.GetBannedRights()
		a.Equal(ok, ok2)
		a.Equal(v2, v)
	}
	{
		v, ok := u.DefaultBannedRights()
		v2, ok2 := u.raw.GetDefaultBannedRights()
		a.Equal(ok2, ok)
		a.Equal(v2, v)
	}
}

func TestChannel_Leave(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	ch := m.Channel(getTestChannel())

	mock.ExpectCall(&tg.ChannelsLeaveChannelRequest{
		Channel: ch.InputChannel(),
	}).ThenRPCErr(getTestError())
	a.Error(ch.Leave(ctx))

	mock.ExpectCall(&tg.ChannelsLeaveChannelRequest{
		Channel: ch.InputChannel(),
	}).ThenResult(&tg.Updates{})
	a.NoError(ch.Leave(ctx))
}

func TestChannel_SetTitle(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	title := "title"
	ch := m.Channel(getTestChannel())

	mock.ExpectCall(&tg.ChannelsEditTitleRequest{
		Channel: ch.InputChannel(),
		Title:   title,
	}).ThenRPCErr(getTestError())
	a.Error(ch.SetTitle(ctx, title))

	mock.ExpectCall(&tg.ChannelsEditTitleRequest{
		Channel: ch.InputChannel(),
		Title:   title,
	}).ThenResult(&tg.Updates{})
	a.NoError(ch.SetTitle(ctx, title))
}

func TestChannel_SetDescription(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	about := "about"
	ch := m.Channel(getTestChannel())

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

func TestChannel_Join(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	ch := m.Channel(getTestChannel())

	mock.ExpectCall(&tg.ChannelsJoinChannelRequest{
		Channel: ch.InputChannel(),
	}).ThenRPCErr(getTestError())
	a.Error(ch.Join(ctx))

	mock.ExpectCall(&tg.ChannelsJoinChannelRequest{
		Channel: ch.InputChannel(),
	}).ThenResult(&tg.Updates{})
	a.NoError(ch.Join(ctx))
}

func TestChannel_Delete(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	ch := m.Channel(getTestChannel())

	mock.ExpectCall(&tg.ChannelsDeleteChannelRequest{
		Channel: ch.InputChannel(),
	}).ThenRPCErr(getTestError())
	a.Error(ch.Delete(ctx))

	mock.ExpectCall(&tg.ChannelsDeleteChannelRequest{
		Channel: ch.InputChannel(),
	}).ThenResult(&tg.Updates{})
	a.NoError(ch.Delete(ctx))
}

func TestChannel_KickUser(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	u := m.User(getTestUser())
	ch := m.Channel(getTestChannel())
	rights := tg.ChatBannedRights{
		ViewMessages: true,
	}
	rights.SetFlags()

	mock.ExpectCall(&tg.ChannelsEditBannedRequest{
		Channel:      ch.InputChannel(),
		Participant:  u.InputPeer(),
		BannedRights: rights,
	}).ThenRPCErr(getTestError())
	a.Error(ch.KickUser(ctx, u.InputUser(), false))

	mock.ExpectCall(&tg.ChannelsEditBannedRequest{
		Channel:      ch.InputChannel(),
		Participant:  u.InputPeer(),
		BannedRights: rights,
	}).ThenResult(&tg.Updates{})
	a.NoError(ch.KickUser(ctx, u.InputUser(), false))
}

func TestChannel_EditAdminRights(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	u := m.User(getTestUser())
	ch := m.Channel(getTestChannel())
	rights := tg.ChatAdminRights{
		AddAdmins: true,
	}
	rights.SetFlags()

	mock.ExpectCall(&tg.ChannelsEditAdminRequest{
		Channel:     ch.InputChannel(),
		UserID:      u.InputUser(),
		AdminRights: rights,
		Rank:        "rank",
	}).ThenRPCErr(getTestError())
	a.Error(ch.EditAdminRights(ctx, u.InputUser(), AdminRights{
		Rank:      "rank",
		AddAdmins: true,
	}))

	mock.ExpectCall(&tg.ChannelsEditAdminRequest{
		Channel:     ch.InputChannel(),
		UserID:      u.InputUser(),
		AdminRights: rights,
		Rank:        "rank",
	}).ThenResult(&tg.Updates{})
	a.NoError(ch.EditAdminRights(ctx, u.InputUser(), AdminRights{
		Rank:      "rank",
		AddAdmins: true,
	}))
}
