package members

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestMemberRights_ApplyFor(t *testing.T) {
	var r MemberRights
	r.ApplyFor(time.Second)
	require.False(t, r.UntilDate.IsZero())
}

func TestMemberRights_IntoChatBannedRights(t *testing.T) {
	r := MemberRights{
		DenyViewMessages: true,
		DenySendMessages: true,
		DenySendMedia:    true,
		DenySendStickers: true,
		DenySendGifs:     true,
		DenySendGames:    true,
		DenySendInline:   true,
		DenyEmbedLinks:   true,
		DenySendPolls:    true,
		DenyChangeInfo:   true,
		DenyInviteUsers:  true,
		DenyPinMessages:  true,
		UntilDate:        time.Time{},
	}

	rights := r.IntoChatBannedRights()
	expected := tg.ChatBannedRights{
		ViewMessages: true,
		SendMessages: true,
		SendMedia:    true,
		SendStickers: true,
		SendGifs:     true,
		SendGames:    true,
		SendInline:   true,
		EmbedLinks:   true,
		SendPolls:    true,
		ChangeInfo:   true,
		InviteUsers:  true,
		PinMessages:  true,
		UntilDate:    0,
	}
	expected.SetFlags()
	require.Equal(t, expected, rights)

	r.ApplyFor(time.Second)
	rights = r.IntoChatBannedRights()
	require.NotZero(t, rights.UntilDate)
}
