package constant

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsID(t *testing.T) {
	tests := []struct {
		id      int64
		user    bool
		chat    bool
		channel bool
	}{
		{id: 309570373, user: true},
		{id: 140267078, user: true},
		{id: -365219918, chat: true},
		{id: -1001228418968, channel: true},
	}
	for _, tt := range tests {
		t.Run(strconv.FormatInt(tt.id, 10), func(t *testing.T) {
			a := require.New(t)

			a.Equal(tt.user, IsUserTDLibID(tt.id))
			a.Equal(tt.chat, IsChatTDLibID(tt.id))
			a.Equal(tt.channel, IsChannelTDLibID(tt.id))
		})
	}
}
