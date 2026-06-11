package members

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestChannelMembers_Promote(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	u := m.User(getTestUser())
	ch := m.Channel(getTestChannel())
	members := Channel(ch)

	rights := tg.ChatAdminRights{BanUsers: true}
	rights.SetFlags()

	mock.ExpectCall(&tg.ChannelsEditAdminRequest{
		Channel:     ch.InputChannel(),
		UserID:      u.InputUser(),
		AdminRights: rights,
	}).ThenResult(&tg.Updates{})
	a.NoError(members.Promote(ctx, u.InputUser(), AdminRights{BanUsers: true}))

	var empty tg.ChatAdminRights
	empty.SetFlags()
	mock.ExpectCall(&tg.ChannelsEditAdminRequest{
		Channel:     ch.InputChannel(),
		UserID:      u.InputUser(),
		AdminRights: empty,
	}).ThenResult(&tg.Updates{})
	a.NoError(members.Demote(ctx, u.InputUser()))
}

func TestChannelMembers_BanUnban(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	u := m.User(getTestUser())
	ch := m.Channel(getTestChannel())
	members := Channel(ch)

	banned := MemberRights{DenyViewMessages: true}.IntoChatBannedRights()
	mock.ExpectCall(&tg.ChannelsEditBannedRequest{
		Channel:      ch.InputChannel(),
		Participant:  u.InputPeer(),
		BannedRights: banned,
	}).ThenResult(&tg.Updates{})
	a.NoError(members.Ban(ctx, u.InputPeer()))

	unban := MemberRights{}.IntoChatBannedRights()
	mock.ExpectCall(&tg.ChannelsEditBannedRequest{
		Channel:      ch.InputChannel(),
		Participant:  u.InputPeer(),
		BannedRights: unban,
	}).ThenResult(&tg.Updates{})
	a.NoError(members.Unban(ctx, u.InputPeer()))
}
