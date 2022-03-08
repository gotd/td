package members

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChatInfoUnavailableError_Error(t *testing.T) {
	require.Equal(t, (&ChatInfoUnavailableError{}).Error(), "chat members info is unavailable")
}

func TestChannelInfoUnavailableError_Error(t *testing.T) {
	require.Equal(t, (&ChannelInfoUnavailableError{}).Error(), "channel members info is unavailable")
}
