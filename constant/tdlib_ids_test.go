package constant

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsID(t *testing.T) {
	tests := []struct {
		id      TDLibPeerID
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
		t.Run(strconv.FormatInt(int64(tt.id), 10), func(t *testing.T) {
			a := require.New(t)

			a.Equal(tt.user, tt.id.IsUser())
			a.Equal(tt.chat, tt.id.IsChat())
			a.Equal(tt.channel, tt.id.IsChannel())
		})
	}
}

func TestTDLibPeerID_ToPlain(t *testing.T) {
	tests := []struct {
		id    TDLibPeerID
		wantR int64
	}{
		{309570373, 309570373},
		{140267078, 140267078},
		{-365219918, 365219918},
		{-1001228418968, 1228418968},
	}
	for _, tt := range tests {
		t.Run(strconv.FormatInt(int64(tt.id), 10), func(t *testing.T) {
			a := require.New(t)
			plain := tt.id.ToPlain()
			a.Equal(tt.wantR, plain)

			switch {
			case tt.id.IsUser():
				var tdlibID TDLibPeerID
				tdlibID.User(plain)
				a.Equal(tt.id, tdlibID)
			case tt.id.IsChat():
				var tdlibID TDLibPeerID
				tdlibID.Chat(plain)
				a.Equal(tt.id, tdlibID)
			case tt.id.IsChannel():
				var tdlibID TDLibPeerID
				tdlibID.Channel(plain)
				a.Equal(tt.id, tdlibID)
			}
		})
	}
}
