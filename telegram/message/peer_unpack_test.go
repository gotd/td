package message

import (
	"context"
	"testing"

	"github.com/gotd/td/tg"
)

func TestUnpack(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	var p tg.InputPeerClass = &tg.InputPeerUser{
		UserID:     10,
		AccessHash: 10,
	}
	_, err := sender.To(p).AsInputChannel(ctx)
	mock.Error(err)

	_, err = sender.To(p).AsInputUser(ctx)
	mock.NoError(err)

	p = &tg.InputPeerChannel{
		ChannelID:  10,
		AccessHash: 10,
	}
	_, err = sender.To(p).AsInputChannel(ctx)
	mock.NoError(err)

	_, err = sender.To(p).AsInputUser(ctx)
	mock.Error(err)
}
