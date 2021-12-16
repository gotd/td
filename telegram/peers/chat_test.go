package peers

import (
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
}
