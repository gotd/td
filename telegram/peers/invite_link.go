package peers

import (
	"context"
	"time"

	"github.com/gotd/td/tg"
)

// InviteLink represents invite link.
type InviteLink struct {
	peer      Peer
	m         *Manager
	raw       tg.ChatInviteExported
	newInvite tg.ChatInviteExported
}

func (e InviteLinks) inviteLink(raw tg.ChatInviteExported) InviteLink {
	return InviteLink{
		peer: e.peer,
		m:    e.m,
		raw:  raw,
	}
}

func (e InviteLinks) replacedLink(raw, newInvite tg.ChatInviteExported) InviteLink {
	link := e.inviteLink(raw)
	link.newInvite = newInvite
	return link
}

// ReplacedWith returns new InviteLink, if any.
func (l InviteLink) ReplacedWith() (InviteLink, bool) {
	return InviteLink{
		peer: l.peer,
		m:    l.m,
		raw:  l.newInvite,
	}, !l.newInvite.Zero()
}

// Raw returns raw tg.ChatInviteExported.
func (l InviteLink) Raw() *tg.ChatInviteExported {
	return &l.raw
}

// Revoked whether this chat invite was revoked
func (l InviteLink) Revoked() bool {
	return l.raw.GetRevoked()
}

// Permanent whether this chat invite has no expiration
func (l InviteLink) Permanent() bool {
	return l.raw.GetPermanent()
}

// RequestNeeded whether users joining the chat via the link need to be approved by chat administrators.
func (l InviteLink) RequestNeeded() bool {
	return l.raw.GetRequestNeeded()
}

// Link returns chat invitation link.
func (l InviteLink) Link() string {
	return l.raw.GetLink()
}

// Creator returns link creator.
func (l InviteLink) Creator(ctx context.Context) (User, error) {
	return l.m.GetUser(ctx, &tg.InputUser{
		UserID: l.raw.AdminID,
	})
}

func telegramDate(date int) time.Time {
	return time.Unix(int64(date), 0)
}

// CreatedAt returns time when was this chat invite created.
func (l InviteLink) CreatedAt() time.Time {
	return telegramDate(l.raw.GetDate())
}

// StartDate returns time when was this chat invite last modified.
func (l InviteLink) StartDate() (time.Time, bool) {
	v, ok := l.raw.GetStartDate()
	if !ok {
		return time.Time{}, false
	}
	return telegramDate(v), true
}

// ExpireDate returns time when does this chat invite expire.
func (l InviteLink) ExpireDate() (time.Time, bool) {
	v, ok := l.raw.GetExpireDate()
	if !ok {
		return time.Time{}, false
	}
	return telegramDate(v), true
}

// UsageLimit returns maximum number of users that can join using this link.
func (l InviteLink) UsageLimit() (int, bool) {
	return l.raw.GetUsageLimit()
}

// Usage returns how many users joined using this link.
func (l InviteLink) Usage() (int, bool) {
	return l.raw.GetUsage()
}

// Requested returns number of pending join requests.
func (l InviteLink) Requested() (int, bool) {
	return l.raw.GetRequested()
}

// Title of this link.
func (l InviteLink) Title() (string, bool) {
	return l.raw.GetTitle()
}
