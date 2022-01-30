package peers

import (
	"time"

	"github.com/gotd/td/tg"
)

// ParticipantRights is options for Channel.EditParticipantRights.
type ParticipantRights struct {
	// If set, does not allow a user to view messages in a supergroup/channel/chat
	ViewMessages bool
	// If set, does not allow a user to send messages in a supergroup/chat
	SendMessages bool
	// If set, does not allow a user to send any media in a supergroup/chat
	SendMedia bool
	// If set, does not allow a user to send stickers in a supergroup/chat
	SendStickers bool
	// If set, does not allow a user to send gifs in a supergroup/chat
	SendGifs bool
	// If set, does not allow a user to send games in a supergroup/chat
	SendGames bool
	// If set, does not allow a user to use inline bots in a supergroup/chat
	SendInline bool
	// If set, does not allow a user to embed links in the messages of a supergroup/chat
	EmbedLinks bool
	// If set, does not allow a user to send polls in a supergroup/chat
	SendPolls bool
	// If set, does not allow any user to change the description of a supergroup/chat
	ChangeInfo bool
	// If set, does not allow any user to invite users in a supergroup/chat
	InviteUsers bool
	// If set, does not allow any user to pin messages in a supergroup/chat
	PinMessages bool
	// Validity of said permissions (it is considered forever any value less than 30 seconds or more than 366 days).
	UntilDate time.Time
}

// IntoChatBannedRights converts ParticipantRights into tg.ChatBannedRights.
func (b ParticipantRights) IntoChatBannedRights() (r tg.ChatBannedRights) {
	r = tg.ChatBannedRights{
		ViewMessages: b.ViewMessages,
		SendMessages: b.SendMessages,
		SendMedia:    b.SendMedia,
		SendStickers: b.SendStickers,
		SendGifs:     b.SendGifs,
		SendGames:    b.SendGames,
		SendInline:   b.SendInline,
		EmbedLinks:   b.EmbedLinks,
		SendPolls:    b.SendPolls,
		ChangeInfo:   b.ChangeInfo,
		InviteUsers:  b.InviteUsers,
		PinMessages:  b.PinMessages,
	}
	if !b.UntilDate.IsZero() {
		r.UntilDate = int(b.UntilDate.Unix())
	}
	r.SetFlags()
	return r
}
