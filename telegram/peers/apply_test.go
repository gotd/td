package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestManager_applyChats(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)
	_, m := testManager(t)

	chats := []tg.ChatClass{
		&tg.ChatEmpty{ID: 1},
		&tg.Chat{ID: 2},
		&tg.ChatForbidden{ID: 3},
		&tg.Channel{ID: 4, AccessHash: 14},
		&tg.ChannelForbidden{ID: 5, AccessHash: 15},
	}

	a.NoError(m.applyChats(ctx, chats...))

	_, ok, err := m.storage.Find(ctx, Key{ID: 2, Prefix: chatsPrefix})
	a.NoError(err)
	a.True(ok)

	_, ok, err = m.storage.Find(ctx, Key{ID: 3, Prefix: chatsPrefix})
	a.NoError(err)
	a.True(ok)

	v, ok, err := m.storage.Find(ctx, Key{ID: 4, Prefix: channelPrefix})
	a.NoError(err)
	a.True(ok)
	a.Equal(int64(14), v.AccessHash)

	v, ok, err = m.storage.Find(ctx, Key{ID: 5, Prefix: channelPrefix})
	a.NoError(err)
	a.True(ok)
	a.Equal(int64(15), v.AccessHash)
}
