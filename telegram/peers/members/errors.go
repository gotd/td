package members

import (
	"github.com/gotd/td/tg"
)

// ChatInfoUnavailableError reports that chat members info is not available.
type ChatInfoUnavailableError struct {
	Info *tg.ChatParticipantsForbidden
}

// Error implements error.
func (c *ChatInfoUnavailableError) Error() string {
	return "chat members info is unavailable"
}

// ChannelInfoUnavailableError reports that channel members info is not available.
type ChannelInfoUnavailableError struct {
}

// Error implements error.
func (c *ChannelInfoUnavailableError) Error() string {
	return "channel members info is unavailable"
}
