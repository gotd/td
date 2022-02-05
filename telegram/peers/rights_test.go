package peers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestParticipantRights_ApplyFor(t *testing.T) {
	var r ParticipantRights
	r.ApplyFor(time.Second)
	require.False(t, r.UntilDate.IsZero())
}

func TestParticipantRights_IntoChatBannedRights(t *testing.T) {
	r := ParticipantRights{
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
