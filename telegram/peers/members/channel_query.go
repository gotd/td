package members

import (
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
)

// ChannelQuery is builder for channel members querying.
type ChannelQuery struct {
	Channel peers.Channel
}

func (q ChannelQuery) query(filter tg.ChannelParticipantsFilterClass) *ChannelMembers {
	return &ChannelMembers{
		m:       q.Channel.Manager(),
		filter:  filter,
		channel: q.Channel,
	}
}

// Recent queries recent members.
func (q ChannelQuery) Recent() *ChannelMembers {
	return q.query(&tg.ChannelParticipantsRecent{})
}

// Admins queries admins members.
func (q ChannelQuery) Admins() *ChannelMembers {
	return q.query(&tg.ChannelParticipantsAdmins{})
}

// Kicked queries kicked members.
func (q ChannelQuery) Kicked(query string) *ChannelMembers {
	return q.query(&tg.ChannelParticipantsKicked{Q: query})
}

// Bots queries bots members.
func (q ChannelQuery) Bots() *ChannelMembers {
	return q.query(&tg.ChannelParticipantsBots{})
}

// Banned queries banned members.
func (q ChannelQuery) Banned(query string) *ChannelMembers {
	return q.query(&tg.ChannelParticipantsBanned{Q: query})
}

// Search queries members by given name.
func (q ChannelQuery) Search(query string) *ChannelMembers {
	return q.query(&tg.ChannelParticipantsSearch{Q: query})
}

// Contacts queries members that are also contacts.
func (q ChannelQuery) Contacts(query string) *ChannelMembers {
	return q.query(&tg.ChannelParticipantsContacts{Q: query})
}

// Custom creates query with custom filter.
func (q ChannelQuery) Custom(filter tg.ChannelParticipantsFilterClass) *ChannelMembers {
	return q.query(filter)
}
