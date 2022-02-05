package peers

import (
	"time"

	"github.com/gotd/td/tg"
)

// AdminRights represents admin right settings.
type AdminRights struct {
	// Indicates the role (rank) of the admin in the group: just an arbitrary string.
	//
	// If empty, will not be used.
	Rank string
	// If set, allows the admin to modify the description of the channel/supergroup.
	ChangeInfo bool
	// If set, allows the admin to post messages in the channel.
	PostMessages bool
	// If set, allows the admin to also edit messages from other admins in the channel.
	EditMessages bool
	// If set, allows the admin to also delete messages from other admins in the channel.
	DeleteMessages bool
	// If set, allows the admin to ban users from the channel/supergroup.
	BanUsers bool
	// If set, allows the admin to invite users in the channel/supergroup.
	InviteUsers bool
	// If set, allows the admin to pin messages in the channel/supergroup.
	PinMessages bool
	// If set, allows the admin to add other admins with the same (or more limited)
	// permissions in the channel/supergroup.
	AddAdmins bool
	// Whether this admin is anonymous.
	Anonymous bool
	// If set, allows the admin to change group call/livestream settings.
	ManageCall bool
	// Set this flag if none of the other flags are set, but you still want the user to be an
	// admin.
	Other bool
}

// IntoChatAdminRights converts AdminRights into tg.ChatAdminRights.
func (b AdminRights) IntoChatAdminRights() (r tg.ChatAdminRights) {
	r.ChangeInfo = b.ChangeInfo
	r.PostMessages = b.PostMessages
	r.EditMessages = b.EditMessages
	r.DeleteMessages = b.DeleteMessages
	r.BanUsers = b.BanUsers
	r.InviteUsers = b.InviteUsers
	r.PinMessages = b.PinMessages
	r.AddAdmins = b.AddAdmins
	r.Anonymous = b.Anonymous
	r.ManageCall = b.ManageCall
	r.Other = b.Other
	r.SetFlags()
	return r
}

// ParticipantRights represents participant right settings.
type ParticipantRights struct {
	// If set, does not allow a user to view messages in a supergroup/channel/chat.
	//
	// In fact, user will be kicked.
	DenyViewMessages bool
	// If set, does not allow a user to send messages in a supergroup/chat.
	DenySendMessages bool
	// If set, does not allow a user to send any media in a supergroup/chat.
	DenySendMedia bool
	// If set, does not allow a user to send stickers in a supergroup/chat.
	DenySendStickers bool
	// If set, does not allow a user to send gifs in a supergroup/chat.
	DenySendGifs bool
	// If set, does not allow a user to send games in a supergroup/chat.
	DenySendGames bool
	// If set, does not allow a user to use inline bots in a supergroup/chat.
	DenySendInline bool
	// If set, does not allow a user to embed links in the messages of a supergroup/chat.
	DenyEmbedLinks bool
	// If set, does not allow a user to send polls in a supergroup/chat.
	DenySendPolls bool
	// If set, does not allow any user to change the description of a supergroup/chat.
	DenyChangeInfo bool
	// If set, does not allow any user to invite users in a supergroup/chat.
	DenyInviteUsers bool
	// If set, does not allow any user to pin messages in a supergroup/chat.
	DenyPinMessages bool
	// Validity of said permissions (it is considered forever any value less than 30 seconds or more than 366 days).
	//
	// If value is zero, value will not be used.
	UntilDate time.Time
}

// ApplyFor sets duration of validity of set rights.
func (b *ParticipantRights) ApplyFor(d time.Duration) {
	b.UntilDate = time.Now().Add(d)
}

// IntoChatBannedRights converts ParticipantRights into tg.ChatBannedRights.
func (b ParticipantRights) IntoChatBannedRights() (r tg.ChatBannedRights) {
	r = tg.ChatBannedRights{
		ViewMessages: b.DenyViewMessages,
		SendMessages: b.DenySendMessages,
		SendMedia:    b.DenySendMedia,
		SendStickers: b.DenySendStickers,
		SendGifs:     b.DenySendGifs,
		SendGames:    b.DenySendGames,
		SendInline:   b.DenySendInline,
		EmbedLinks:   b.DenyEmbedLinks,
		SendPolls:    b.DenySendPolls,
		ChangeInfo:   b.DenyChangeInfo,
		InviteUsers:  b.DenyInviteUsers,
		PinMessages:  b.DenyPinMessages,
	}
	if !b.UntilDate.IsZero() {
		r.UntilDate = int(b.UntilDate.Unix())
	}
	r.SetFlags()
	return r
}
