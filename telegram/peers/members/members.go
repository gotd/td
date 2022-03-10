// Package members defines interfaces for working with chat/channel members.
package members

import (
	"context"
	"time"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
)

var _ = []Member{
	ChatMember{},
	ChannelMember{},
}

// Member represents chat/channel member.
type Member interface {
	// Status returns member Status.
	Status() Status
	// JoinDate returns member join date, if it is available.
	JoinDate() (time.Time, bool)
	// InvitedBy returns user that invited this member.
	InvitedBy() (peers.User, bool)
	// User returns member User object.
	User() peers.User
	// Kick kicks this member.
	//
	// If revokeHistory is set, will delete all messages from this member.
	Kick(ctx context.Context, revokeHistory bool) error
}

// Callback is type for member iterator callback.
type Callback = func(p Member) error

var _ = []Members{
	&ChatMembers{},
	&ChannelMembers{},
}

// Members represents chat/channel members.
type Members interface {
	// ForEach calls cb for every member of chat/channel.
	ForEach(ctx context.Context, cb Callback) error
	// Count returns total count of members.
	Count(ctx context.Context) (int, error)
	// Peer returns chat object.
	Peer() peers.Peer
	// Kick kicks user member.
	//
	// If revokeHistory is set, will delete all messages from this member.
	Kick(ctx context.Context, member tg.InputUserClass, revokeHistory bool) error
	// EditRights edits rights of all members in this chat/channel.
	EditRights(ctx context.Context, options MemberRights) error
}
