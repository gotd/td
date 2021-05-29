package updates

import "github.com/gotd/td/tg"

type update struct {
	Value interface{}
	State int
	Count int
	Ents  *Entities
}

func (u update) start() int { return u.State - u.Count + 1 }

func (u update) end() int { return u.State }

// Entities contains update entities.
type Entities struct {
	Users             map[int]*tg.User
	Chats             map[int]*tg.Chat
	Channels          map[int]*tg.Channel
	ChannelsForbidden map[int]*tg.ChannelForbidden
}

// NewEntities creates new Entities.
func NewEntities() *Entities {
	return &Entities{
		Users:             map[int]*tg.User{},
		Chats:             map[int]*tg.Chat{},
		Channels:          map[int]*tg.Channel{},
		ChannelsForbidden: map[int]*tg.ChannelForbidden{},
	}
}

// Merge merges entities.
func (e *Entities) Merge(from *Entities) {
	if from == nil {
		return
	}

	for userID, user := range from.Users {
		e.Users[userID] = user
	}

	for chanID, chat := range from.Chats {
		e.Chats[chanID] = chat
	}

	for channelID, channel := range from.Channels {
		e.Channels[channelID] = channel
	}

	for channelID, channel := range from.ChannelsForbidden {
		e.ChannelsForbidden[channelID] = channel
	}
}

func (e *Entities) fromUpdates(u interface {
	tg.UpdatesClass
	MapUsers() tg.UserClassArray
	MapChats() tg.ChatClassArray
}) *Entities {
	u.MapChats().FillChatMap(e.Chats)
	u.MapChats().FillChannelMap(e.Channels)
	u.MapChats().FillChannelForbiddenMap(e.ChannelsForbidden)
	u.MapUsers().FillUserMap(e.Users)
	return e
}

// AsUsers returns users as tg.UserClass slice.
func (e *Entities) AsUsers() []tg.UserClass {
	var users []tg.UserClass
	for _, u := range e.Users {
		users = append(users, u)
	}
	return users
}

// AsChats returns chats as tg.ChatClass slice.
func (e *Entities) AsChats() []tg.ChatClass {
	var chats []tg.ChatClass
	for _, c := range e.Chats {
		chats = append(chats, c)
	}
	for _, c := range e.Channels {
		chats = append(chats, c)
	}
	for _, c := range e.ChannelsForbidden {
		chats = append(chats, c)
	}
	return chats
}
