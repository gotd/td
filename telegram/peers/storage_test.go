package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoopCache(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	c := NoopCache{}

	_, ok, err := c.FindUser(ctx, 1)
	a.NoError(err)
	a.False(ok)

	_, ok, err = c.FindUserFull(ctx, 1)
	a.NoError(err)
	a.False(ok)

	_, ok, err = c.FindChat(ctx, 1)
	a.NoError(err)
	a.False(ok)

	_, ok, err = c.FindChatFull(ctx, 1)
	a.NoError(err)
	a.False(ok)

	_, ok, err = c.FindChannel(ctx, 1)
	a.NoError(err)
	a.False(ok)

	_, ok, err = c.FindChannelFull(ctx, 1)
	a.NoError(err)
	a.False(ok)

	a.NoError(c.SaveUsers(ctx, nil))
	a.NoError(c.SaveUserFulls(ctx, nil))
	a.NoError(c.SaveChats(ctx, nil))
	a.NoError(c.SaveChatFulls(ctx, nil))
	a.NoError(c.SaveChannels(ctx, nil))
	a.NoError(c.SaveChannelFulls(ctx, nil))
}
