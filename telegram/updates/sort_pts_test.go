package updates

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func Test_sortUpdatesByPts(t *testing.T) {
	channelNewMessage := func(pts int, id int64) *tg.UpdateNewChannelMessage {
		return &tg.UpdateNewChannelMessage{
			Message: &tg.Message{
				PeerID: &tg.PeerChannel{ChannelID: id},
			},
			Pts:      pts - 1,
			PtsCount: 1,
		}
	}
	newMessage := func(pts int) *tg.UpdateNewMessage {
		return &tg.UpdateNewMessage{
			Message: &tg.Message{
				PeerID: &tg.PeerUser{UserID: 10},
			},
			Pts:      pts - 1,
			PtsCount: 1,
		}
	}
	encryptedNewMessage := func(pts int) *tg.UpdateNewEncryptedMessage {
		return &tg.UpdateNewEncryptedMessage{
			Qts: pts,
		}
	}
	channelReadInbox := func(pts int, id int64) *tg.UpdateReadChannelInbox {
		return &tg.UpdateReadChannelInbox{ChannelID: id, MaxID: 25, Pts: pts}
	}

	tests := []struct {
		input  []tg.UpdateClass
		result []tg.UpdateClass
	}{
		{
			[]tg.UpdateClass{
				channelReadInbox(26, 1),
				channelNewMessage(25, 1),
			},
			[]tg.UpdateClass{
				channelNewMessage(25, 1),
				channelReadInbox(26, 1),
			},
		},
		{
			[]tg.UpdateClass{
				channelReadInbox(26, 1),
				channelNewMessage(25, 2),
				newMessage(26),
				encryptedNewMessage(26),
				newMessage(25),
				encryptedNewMessage(25),
				encryptedNewMessage(27),
				channelNewMessage(25, 1),
			},
			[]tg.UpdateClass{
				newMessage(25),
				newMessage(26),
				encryptedNewMessage(25),
				encryptedNewMessage(26),
				encryptedNewMessage(27),
				channelNewMessage(25, 1),
				channelReadInbox(26, 1),
				channelNewMessage(25, 2),
			},
		},
		{
			[]tg.UpdateClass{
				channelReadInbox(26, 1),
				&tg.UpdateConfig{},
				channelNewMessage(25, 1),
			},
			[]tg.UpdateClass{
				&tg.UpdateConfig{},
				channelNewMessage(25, 1),
				channelReadInbox(26, 1),
			},
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			sortUpdatesByPts(tt.input)
			require.Equal(t, tt.result, tt.input)
		})
	}
}
