package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func TestUnpack(t *testing.T) {
	ctx := context.Background()
	sender, _ := testSender(t)

	var p tg.InputPeerClass = &tg.InputPeerUser{
		UserID:     10,
		AccessHash: 10,
	}
	_, err := sender.To(p).AsInputChannel(ctx)
	require.Error(t, err)

	_, err = sender.To(p).AsInputUser(ctx)
	require.NoError(t, err)

	p = &tg.InputPeerChannel{
		ChannelID:  10,
		AccessHash: 10,
	}
	_, err = sender.To(p).AsInputChannel(ctx)
	require.NoError(t, err)

	_, err = sender.To(p).AsInputUser(ctx)
	require.Error(t, err)
}
