package tg

import "context"

// BeforeHandler is called for every dispatched update before its handler,
// including updates without handler, e.g. for metrics, tracing or filtering.
//
// A non-nil error skips the update's handler and is returned by Handle,
// joined with errors of other updates of the same batch.
type BeforeHandler func(ctx context.Context, e Entities, update UpdateClass) error

// OnBefore sets hook called for every dispatched update before its handler.
func (u UpdateDispatcher) OnBefore(handler BeforeHandler) {
	u.state.setBefore(handler)
}

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

// AnyRead is read history event: messages up to and including MaxID in Peer
// (and topic, if TopMsgID or SavedPeer is set) are read.
type AnyRead struct {
	// Peer is the chat.
	Peer PeerClass
	// MaxID is the ID of the last read message.
	MaxID int
	// TopMsgID is the forum topic or discussion thread ID, zero if none.
	TopMsgID int
	// SavedPeer is the monoforum (channel direct messages) topic, nil if none.
	SavedPeer PeerClass
	// Update is the source update, for fields not mapped here.
	Update UpdateClass
}

// AnyReadHandler is a handler of read history.
type AnyReadHandler func(ctx context.Context, e Entities, read AnyRead) error

// OnAnyReadInbox sets handler of read incoming messages in any chat or topic,
// i.e. UpdateReadHistoryInbox, UpdateReadChannelInbox,
// UpdateReadChannelDiscussionInbox and UpdateReadMonoForumInbox.
//
// It replaces handlers set by OnReadHistoryInbox, OnReadChannelInbox,
// OnReadChannelDiscussionInbox and OnReadMonoForumInbox, and is replaced by
// them.
func (u UpdateDispatcher) OnAnyReadInbox(handler AnyReadHandler) {
	u.OnReadHistoryInbox(func(ctx context.Context, e Entities, update *UpdateReadHistoryInbox) error {
		return handler(ctx, e, AnyRead{
			Peer:     update.Peer,
			MaxID:    update.MaxID,
			TopMsgID: update.TopMsgID,
			Update:   update,
		})
	})
	u.OnReadChannelInbox(func(ctx context.Context, e Entities, update *UpdateReadChannelInbox) error {
		return handler(ctx, e, AnyRead{
			Peer:   &PeerChannel{ChannelID: update.ChannelID},
			MaxID:  update.MaxID,
			Update: update,
		})
	})
	u.OnReadChannelDiscussionInbox(func(ctx context.Context, e Entities, update *UpdateReadChannelDiscussionInbox) error {
		return handler(ctx, e, AnyRead{
			Peer:     &PeerChannel{ChannelID: update.ChannelID},
			MaxID:    update.ReadMaxID,
			TopMsgID: update.TopMsgID,
			Update:   update,
		})
	})
	u.OnReadMonoForumInbox(func(ctx context.Context, e Entities, update *UpdateReadMonoForumInbox) error {
		return handler(ctx, e, AnyRead{
			Peer:      &PeerChannel{ChannelID: update.ChannelID},
			MaxID:     update.ReadMaxID,
			SavedPeer: update.SavedPeerID,
			Update:    update,
		})
	})
}

// OnAnyReadOutbox sets handler of read outgoing messages in any chat or topic,
// i.e. UpdateReadHistoryOutbox, UpdateReadChannelOutbox,
// UpdateReadChannelDiscussionOutbox and UpdateReadMonoForumOutbox.
//
// It replaces handlers set by OnReadHistoryOutbox, OnReadChannelOutbox,
// OnReadChannelDiscussionOutbox and OnReadMonoForumOutbox, and is replaced by
// them.
func (u UpdateDispatcher) OnAnyReadOutbox(handler AnyReadHandler) {
	u.OnReadHistoryOutbox(func(ctx context.Context, e Entities, update *UpdateReadHistoryOutbox) error {
		return handler(ctx, e, AnyRead{
			Peer:   update.Peer,
			MaxID:  update.MaxID,
			Update: update,
		})
	})
	u.OnReadChannelOutbox(func(ctx context.Context, e Entities, update *UpdateReadChannelOutbox) error {
		return handler(ctx, e, AnyRead{
			Peer:   &PeerChannel{ChannelID: update.ChannelID},
			MaxID:  update.MaxID,
			Update: update,
		})
	})
	u.OnReadChannelDiscussionOutbox(func(ctx context.Context, e Entities, update *UpdateReadChannelDiscussionOutbox) error {
		return handler(ctx, e, AnyRead{
			Peer:     &PeerChannel{ChannelID: update.ChannelID},
			MaxID:    update.ReadMaxID,
			TopMsgID: update.TopMsgID,
			Update:   update,
		})
	})
	u.OnReadMonoForumOutbox(func(ctx context.Context, e Entities, update *UpdateReadMonoForumOutbox) error {
		return handler(ctx, e, AnyRead{
			Peer:      &PeerChannel{ChannelID: update.ChannelID},
			MaxID:     update.ReadMaxID,
			SavedPeer: update.SavedPeerID,
			Update:    update,
		})
	})
}

// AnyTyping is a chat action (typing, uploading, etc.) event.
type AnyTyping struct {
	// Peer is the chat.
	Peer PeerClass
	// From is the peer performing the action.
	From PeerClass
	// TopMsgID is the forum topic or thread ID, zero if none.
	TopMsgID int
	// Action is the chat action.
	Action SendMessageActionClass
	// Update is the source update, for fields not mapped here.
	Update UpdateClass
}

// AnyTypingHandler is a handler of chat actions.
type AnyTypingHandler func(ctx context.Context, e Entities, typing AnyTyping) error

// OnAnyTyping sets handler of chat actions in any chat, i.e. UpdateUserTyping,
// UpdateChatUserTyping and UpdateChannelUserTyping.
//
// It replaces handlers set by OnUserTyping, OnChatUserTyping and
// OnChannelUserTyping, and is replaced by them.
func (u UpdateDispatcher) OnAnyTyping(handler AnyTypingHandler) {
	u.OnUserTyping(func(ctx context.Context, e Entities, update *UpdateUserTyping) error {
		user := &PeerUser{UserID: update.UserID}
		return handler(ctx, e, AnyTyping{
			Peer:     user,
			From:     user,
			TopMsgID: update.TopMsgID,
			Action:   update.Action,
			Update:   update,
		})
	})
	u.OnChatUserTyping(func(ctx context.Context, e Entities, update *UpdateChatUserTyping) error {
		return handler(ctx, e, AnyTyping{
			Peer:   &PeerChat{ChatID: update.ChatID},
			From:   update.FromID,
			Action: update.Action,
			Update: update,
		})
	})
	u.OnChannelUserTyping(func(ctx context.Context, e Entities, update *UpdateChannelUserTyping) error {
		return handler(ctx, e, AnyTyping{
			Peer:     &PeerChannel{ChannelID: update.ChannelID},
			From:     update.FromID,
			TopMsgID: update.TopMsgID,
			Action:   update.Action,
			Update:   update,
		})
	})
}
