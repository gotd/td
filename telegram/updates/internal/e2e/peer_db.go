package e2e

import "github.com/nnqq/td/tg"

type messageDatabase struct {
	common   []tg.MessageClass
	secret   []tg.EncryptedMessageClass
	channels map[int64][]tg.MessageClass
}

type peerDatabase struct {
	users    map[int64]*tg.User
	chats    map[int64]*tg.Chat
	channels map[int64]*tg.Channel

	id int64
}

func (p *peerDatabase) createUser(username string) *tg.PeerUser {
	p.users[p.id] = &tg.User{
		ID:       p.id,
		Username: username,
	}

	defer func() { p.id++ }()
	return &tg.PeerUser{UserID: p.id}
}

func (p *peerDatabase) createChat(title string) *tg.PeerChat {
	p.chats[p.id] = &tg.Chat{
		ID:    p.id,
		Title: title,
	}

	defer func() { p.id++ }()
	return &tg.PeerChat{ChatID: p.id}
}

func (p *peerDatabase) createChannel(username string) *tg.PeerChannel {
	p.channels[p.id] = &tg.Channel{
		ID:       p.id,
		Username: username,
	}
	p.channels[p.id].SetAccessHash(p.id * 2)

	defer func() { p.id++ }()
	return &tg.PeerChannel{ChannelID: p.id}
}
