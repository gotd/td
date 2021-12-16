package peers

import (
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
}
