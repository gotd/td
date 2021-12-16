package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestManager_getUserFull(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)
	testUserFull := getTestUserFull()
	input := &tg.InputUser{
		UserID:     testUserFull.ID,
		AccessHash: 10,
	}

	mock.ExpectCall(&tg.UsersGetFullUserRequest{ID: input}).ThenResult(&tg.UsersUserFull{
		FullUser: testUserFull,
	})

	v, err := m.getUserFull(ctx, input)
	a.NoError(err)
	a.Equal(&testUserFull, v)

	// Test caching.
	v, err = m.getUserFull(ctx, input)
	a.NoError(err)
	a.Equal(&testUserFull, v)
}

func TestManager_updateUserFull(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)
	testUserFull := getTestUserFull()
	input := &tg.InputUser{
		UserID:     testUserFull.ID,
		AccessHash: 10,
	}

	mock.ExpectCall(&tg.UsersGetFullUserRequest{ID: input}).ThenRPCErr(getTestError())
	_, err := m.updateUserFull(ctx, input)
	a.Error(err)
}

func TestManager_getChatFull(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)
	testChatFull := getTestChatFull()
	input := testChatFull.ID

	mock.ExpectCall(&tg.MessagesGetFullChatRequest{ChatID: input}).ThenResult(&tg.MessagesChatFull{
		FullChat: testChatFull,
	})

	v, err := m.getChatFull(ctx, input)
	a.NoError(err)
	a.Equal(testChatFull, v)

	// Test caching.
	v, err = m.getChatFull(ctx, input)
	a.NoError(err)
	a.Equal(testChatFull, v)
}

func TestManager_updateChatFull(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	testChatFull := getTestChatFull()
	input := testChatFull.ID

	mock.ExpectCall(&tg.MessagesGetFullChatRequest{ChatID: input}).ThenRPCErr(getTestError())
	_, err := m.updateChatFull(ctx, input)
	a.Error(err)
}

func TestManager_getChannelFull(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)
	testChannelFull := getTestChannelFull()
	input := &tg.InputChannel{
		ChannelID:  testChannelFull.ID,
		AccessHash: 10,
	}

	mock.ExpectCall(&tg.ChannelsGetFullChannelRequest{Channel: input}).ThenResult(&tg.MessagesChatFull{
		FullChat: testChannelFull,
	})

	v, err := m.getChannelFull(ctx, input)
	a.NoError(err)
	a.Equal(testChannelFull, v)

	// Test caching.
	v, err = m.getChannelFull(ctx, input)
	a.NoError(err)
	a.Equal(testChannelFull, v)
}

func TestManager_updateChannelFull(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	testChannelFull := getTestChannelFull()
	input := &tg.InputChannel{
		ChannelID:  testChannelFull.ID,
		AccessHash: 10,
	}

	mock.ExpectCall(&tg.ChannelsGetFullChannelRequest{Channel: input}).ThenRPCErr(getTestError())
	_, err := m.updateChannelFull(ctx, input)
	a.Error(err)
}
