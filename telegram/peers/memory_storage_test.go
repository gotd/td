package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestMemoryStorage(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)
	k := Key{
		Prefix: usersPrefix,
		ID:     1,
	}
	v := Value{
		AccessHash: 10,
	}
	phone := "phone"

	var m InmemoryStorage
	_, found, err := m.Find(ctx, k)
	a.NoError(err)
	a.False(found)

	a.NoError(m.Save(ctx, k, v))

	v2, found, err := m.Find(ctx, k)
	a.NoError(err)
	a.True(found)
	a.Equal(v, v2)

	_, _, found, err = m.FindPhone(ctx, phone)
	a.NoError(err)
	a.False(found)

	a.NoError(m.SavePhone(ctx, phone, k))

	k2, v2, found, err := m.FindPhone(ctx, phone)
	a.NoError(err)
	a.True(found)
	a.Equal(k, k2)
	a.Equal(v, v2)

	hash, err := m.GetContactsHash(ctx)
	a.NoError(err)
	a.Zero(hash)

	a.NoError(m.SaveContactsHash(ctx, 1))

	hash, err = m.GetContactsHash(ctx)
	a.NoError(err)
	a.Equal(int64(1), hash)
}

func TestInmemoryCache(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)

	var m InmemoryCache

	{
		value := &tg.User{ID: 10}

		_, found, err := m.FindUser(ctx, value.GetID())
		a.NoError(err)
		a.False(found)

		a.NoError(m.SaveUsers(ctx, value))

		r, found, err := m.FindUser(ctx, value.GetID())
		a.NoError(err)
		a.True(found)
		a.Equal(value, r)
	}
	{
		value := &tg.UserFull{ID: 10}

		_, found, err := m.FindUserFull(ctx, value.GetID())
		a.NoError(err)
		a.False(found)

		a.NoError(m.SaveUserFulls(ctx, value))

		r, found, err := m.FindUserFull(ctx, value.GetID())
		a.NoError(err)
		a.True(found)
		a.Equal(value, r)
	}
	{
		value := &tg.Chat{ID: 10}

		_, found, err := m.FindChat(ctx, value.GetID())
		a.NoError(err)
		a.False(found)

		a.NoError(m.SaveChats(ctx, value))

		r, found, err := m.FindChat(ctx, value.GetID())
		a.NoError(err)
		a.True(found)
		a.Equal(value, r)
	}
	{
		value := &tg.ChatFull{ID: 10}

		_, found, err := m.FindChatFull(ctx, value.GetID())
		a.NoError(err)
		a.False(found)

		a.NoError(m.SaveChatFulls(ctx, value))

		r, found, err := m.FindChatFull(ctx, value.GetID())
		a.NoError(err)
		a.True(found)
		a.Equal(value, r)
	}
	{
		value := &tg.Channel{ID: 10}

		_, found, err := m.FindChannel(ctx, value.GetID())
		a.NoError(err)
		a.False(found)

		a.NoError(m.SaveChannels(ctx, value))

		r, found, err := m.FindChannel(ctx, value.GetID())
		a.NoError(err)
		a.True(found)
		a.Equal(value, r)
	}
	{
		value := &tg.ChannelFull{ID: 10}

		_, found, err := m.FindChannelFull(ctx, value.GetID())
		a.NoError(err)
		a.False(found)

		a.NoError(m.SaveChannelFulls(ctx, value))

		r, found, err := m.FindChannelFull(ctx, value.GetID())
		a.NoError(err)
		a.True(found)
		a.Equal(value, r)
	}
}
