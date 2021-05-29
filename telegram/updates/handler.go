package updates

import "github.com/gotd/td/tg"

// DiffUpdate struct.
type DiffUpdate struct {
	NewMessages          []tg.MessageClass
	NewEncryptedMessages []tg.EncryptedMessageClass
	Users                []tg.UserClass
	Chats                []tg.ChatClass
}

// Handler interface.
type Handler interface {
	HandleDiff(DiffUpdate) error
	HandleUpdates(*Entities, []tg.UpdateClass) error
	ChannelTooLong(channelID int)
}
