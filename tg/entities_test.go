package tg_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

var (
	_ tg.EntitiesResult = (*tg.Updates)(nil)
	_ tg.EntitiesResult = (*tg.UpdatesCombined)(nil)
	_ tg.EntitiesResult = (*tg.UpdatesDifference)(nil)
	_ tg.EntitiesResult = (*tg.UpdatesDifferenceSlice)(nil)
	_ tg.EntitiesResult = (*tg.UpdatesChannelDifference)(nil)
	_ tg.EntitiesResult = (*tg.MessagesMessages)(nil)
	_ tg.EntitiesResult = (*tg.MessagesChannelMessages)(nil)
	_ tg.EntitiesResult = (*tg.MessagesDialogs)(nil)
	_ tg.EntitiesResult = (*tg.ContactsResolvedPeer)(nil)
)

func TestEntitiesFrom(t *testing.T) {
	users := []tg.UserClass{
		&tg.User{ID: 1},
		&tg.UserEmpty{ID: 2},
	}
	chats := []tg.ChatClass{
		&tg.Chat{ID: 10},
		&tg.ChatForbidden{ID: 11},
		&tg.Channel{ID: 20},
		&tg.ChannelForbidden{ID: 21},
	}

	check := func(t *testing.T, e tg.Entities) {
		t.Helper()
		require.False(t, e.Short)
		require.Len(t, e.Users, 1)
		require.Contains(t, e.Users, int64(1))
		require.Len(t, e.Chats, 1)
		require.Contains(t, e.Chats, int64(10))
		require.Len(t, e.Channels, 1)
		require.Contains(t, e.Channels, int64(20))
	}

	t.Run("Slices", func(t *testing.T) {
		check(t, tg.EntitiesFrom(users, chats))
	})
	t.Run("Result", func(t *testing.T) {
		check(t, tg.EntitiesFromResult(&tg.MessagesMessages{Users: users, Chats: chats}))
	})
	t.Run("Empty", func(t *testing.T) {
		e := tg.EntitiesFrom(nil, nil)
		require.NotNil(t, e.Users)
		require.NotNil(t, e.Chats)
		require.NotNil(t, e.Channels)
	})
}
