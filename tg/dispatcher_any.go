package tg

import "context"

// AnyMessageHandler is a handler of a message from any chat: private chat,
// basic group or channel (supergroup).
type AnyMessageHandler func(ctx context.Context, e Entities, msg MessageClass) error

// OnAnyNewMessage sets handler of new messages from any chat, i.e. both
// UpdateNewMessage and UpdateNewChannelMessage.
//
// It replaces handlers set by OnNewMessage and OnNewChannelMessage, and is
// replaced by them.
func (u UpdateDispatcher) OnAnyNewMessage(handler AnyMessageHandler) {
	u.OnNewMessage(func(ctx context.Context, e Entities, update *UpdateNewMessage) error {
		return handler(ctx, e, update.Message)
	})
	u.OnNewChannelMessage(func(ctx context.Context, e Entities, update *UpdateNewChannelMessage) error {
		return handler(ctx, e, update.Message)
	})
}

// OnAnyEditMessage sets handler of edited messages from any chat, i.e. both
// UpdateEditMessage and UpdateEditChannelMessage.
//
// It replaces handlers set by OnEditMessage and OnEditChannelMessage, and is
// replaced by them.
func (u UpdateDispatcher) OnAnyEditMessage(handler AnyMessageHandler) {
	u.OnEditMessage(func(ctx context.Context, e Entities, update *UpdateEditMessage) error {
		return handler(ctx, e, update.Message)
	})
	u.OnEditChannelMessage(func(ctx context.Context, e Entities, update *UpdateEditChannelMessage) error {
		return handler(ctx, e, update.Message)
	})
}

// AnyDeleteMessagesHandler is a handler of deleted messages.
//
// The channelID is zero for private chats and basic groups: message IDs are
// common for them, and the update does not contain the peer.
type AnyDeleteMessagesHandler func(ctx context.Context, e Entities, channelID int64, ids []int) error

// OnAnyDeleteMessages sets handler of deleted messages from any chat, i.e.
// both UpdateDeleteMessages and UpdateDeleteChannelMessages.
//
// It replaces handlers set by OnDeleteMessages and OnDeleteChannelMessages,
// and is replaced by them.
func (u UpdateDispatcher) OnAnyDeleteMessages(handler AnyDeleteMessagesHandler) {
	u.OnDeleteMessages(func(ctx context.Context, e Entities, update *UpdateDeleteMessages) error {
		return handler(ctx, e, 0, update.Messages)
	})
	u.OnDeleteChannelMessages(func(ctx context.Context, e Entities, update *UpdateDeleteChannelMessages) error {
		return handler(ctx, e, update.ChannelID, update.Messages)
	})
}

// AnyReadHandler is a handler of read history: messages up to and including
// maxID in peer are read.
type AnyReadHandler func(ctx context.Context, e Entities, peer PeerClass, maxID int) error

// OnAnyReadInbox sets handler of read incoming messages in any chat, i.e. both
// UpdateReadHistoryInbox and UpdateReadChannelInbox.
//
// It replaces handlers set by OnReadHistoryInbox and OnReadChannelInbox, and
// is replaced by them.
func (u UpdateDispatcher) OnAnyReadInbox(handler AnyReadHandler) {
	u.OnReadHistoryInbox(func(ctx context.Context, e Entities, update *UpdateReadHistoryInbox) error {
		return handler(ctx, e, update.Peer, update.MaxID)
	})
	u.OnReadChannelInbox(func(ctx context.Context, e Entities, update *UpdateReadChannelInbox) error {
		return handler(ctx, e, &PeerChannel{ChannelID: update.ChannelID}, update.MaxID)
	})
}

// OnAnyReadOutbox sets handler of read outgoing messages in any chat, i.e.
// both UpdateReadHistoryOutbox and UpdateReadChannelOutbox.
//
// It replaces handlers set by OnReadHistoryOutbox and OnReadChannelOutbox, and
// is replaced by them.
func (u UpdateDispatcher) OnAnyReadOutbox(handler AnyReadHandler) {
	u.OnReadHistoryOutbox(func(ctx context.Context, e Entities, update *UpdateReadHistoryOutbox) error {
		return handler(ctx, e, update.Peer, update.MaxID)
	})
	u.OnReadChannelOutbox(func(ctx context.Context, e Entities, update *UpdateReadChannelOutbox) error {
		return handler(ctx, e, &PeerChannel{ChannelID: update.ChannelID}, update.MaxID)
	})
}

// AnyTypingHandler is a handler of typing (and other chat actions) in peer by
// from.
type AnyTypingHandler func(ctx context.Context, e Entities, peer, from PeerClass, action SendMessageActionClass) error

// OnAnyTyping sets handler of chat actions in any chat, i.e. UpdateUserTyping,
// UpdateChatUserTyping and UpdateChannelUserTyping.
//
// It replaces handlers set by OnUserTyping, OnChatUserTyping and
// OnChannelUserTyping, and is replaced by them.
func (u UpdateDispatcher) OnAnyTyping(handler AnyTypingHandler) {
	u.OnUserTyping(func(ctx context.Context, e Entities, update *UpdateUserTyping) error {
		user := &PeerUser{UserID: update.UserID}
		return handler(ctx, e, user, user, update.Action)
	})
	u.OnChatUserTyping(func(ctx context.Context, e Entities, update *UpdateChatUserTyping) error {
		return handler(ctx, e, &PeerChat{ChatID: update.ChatID}, update.FromID, update.Action)
	})
	u.OnChannelUserTyping(func(ctx context.Context, e Entities, update *UpdateChannelUserTyping) error {
		return handler(ctx, e, &PeerChannel{ChannelID: update.ChannelID}, update.FromID, update.Action)
	})
}
