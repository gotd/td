package peers

import (
	"context"

	"github.com/go-faster/errors"
)

// GetAllChatsResult is result structure for GetAllChats query.
type GetAllChatsResult struct {
	Chats    []Chat
	Channels []Channel
}

// GetAllChats gets all chats.
func (m *Manager) GetAllChats(ctx context.Context, exceptIDs ...int64) (r GetAllChatsResult, _ error) {
	all, err := m.api.MessagesGetAllChats(ctx, exceptIDs)
	if err != nil {
		return r, errors.Wrap(err, "get all chats")
	}

	chats, channel, err := m.applyMessagesChats(ctx, all)
	if err != nil {
		return r, errors.Wrap(err, "apply messages chats")
	}

	return GetAllChatsResult{
		Chats:    chats,
		Channels: channel,
	}, nil
}
