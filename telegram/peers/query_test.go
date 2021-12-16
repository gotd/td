package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestManager_getUser(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)
	testSelf := getTestSelf()
	testUser := getTestUser()

	mock.ExpectCall(&tg.UsersGetUsersRequest{
		ID: []tg.InputUserClass{&tg.InputUserSelf{}},
	}).ThenResult(&tg.UserClassVector{Elems: []tg.UserClass{testSelf}})

	v, err := m.getUser(ctx, &tg.InputUserSelf{})
	a.NoError(err)
	a.Equal(testSelf, v)

	v, err = m.getUser(ctx, &tg.InputUser{UserID: testSelf.ID, AccessHash: testSelf.AccessHash})
	a.NoError(err)
	a.Equal(testSelf, v)

	v, err = m.getUser(ctx, &tg.InputUserFromMessage{UserID: testSelf.ID})
	a.NoError(err)
	a.Equal(testSelf, v)

	mock.ExpectCall(&tg.UsersGetUsersRequest{
		ID: []tg.InputUserClass{&tg.InputUser{
			UserID:     testUser.ID,
			AccessHash: testUser.AccessHash,
		}},
	}).ThenResult(&tg.UserClassVector{Elems: []tg.UserClass{testUser}})

	v, err = m.getUser(ctx, &tg.InputUser{UserID: testUser.ID, AccessHash: testUser.AccessHash})
	a.NoError(err)
	a.Equal(testUser, v)

	v, err = m.getUser(ctx, &tg.InputUser{UserID: testUser.ID})
	a.NoError(err)
	a.Equal(testUser, v)

	v, err = m.getUser(ctx, &tg.InputUserFromMessage{UserID: testUser.ID})
	a.NoError(err)
	a.Equal(testUser, v)
}

func TestManager_updateUser(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)
	input := &tg.InputUser{
		UserID:     10,
		AccessHash: 10,
	}

	mock.ExpectCall(&tg.UsersGetUsersRequest{
		ID: []tg.InputUserClass{input},
	}).ThenRPCErr(getTestError())

	_, err := m.updateUser(ctx, input)
	a.Error(err)

	mock.ExpectCall(&tg.UsersGetUsersRequest{
		ID: []tg.InputUserClass{input},
	}).ThenResult(&tg.UserClassVector{})

	_, err = m.updateUser(ctx, input)
	a.Error(err)

	mock.ExpectCall(&tg.UsersGetUsersRequest{
		ID: []tg.InputUserClass{input},
	}).ThenResult(&tg.UserClassVector{Elems: []tg.UserClass{&tg.UserEmpty{}}})

	_, err = m.updateUser(ctx, input)
	a.Error(err)
}

func TestManager_getChat(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)
	testChat := getTestChat()

	mock.ExpectCall(&tg.MessagesGetChatsRequest{
		ID: []int64{testChat.ID},
	}).ThenResult(&tg.MessagesChats{Chats: []tg.ChatClass{testChat}})

	v, err := m.getChat(ctx, testChat.ID)
	a.NoError(err)
	a.Equal(testChat, v)

	v, err = m.getChat(ctx, testChat.ID)
	a.NoError(err)
	a.Equal(testChat, v)
}

func TestManager_updateChat(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)
	var input int64 = 10

	mock.ExpectCall(&tg.MessagesGetChatsRequest{
		ID: []int64{input},
	}).ThenRPCErr(getTestError())

	_, err := m.updateChat(ctx, input)
	a.Error(err)

	mock.ExpectCall(&tg.MessagesGetChatsRequest{
		ID: []int64{input},
	}).ThenResult(&tg.MessagesChats{})

	_, err = m.updateChat(ctx, input)
	a.Error(err)

	mock.ExpectCall(&tg.MessagesGetChatsRequest{
		ID: []int64{input},
	}).ThenResult(&tg.MessagesChats{Chats: []tg.ChatClass{&tg.ChatEmpty{}}})

	_, err = m.updateChat(ctx, input)
	a.Error(err)
}

func TestManager_getChannel(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)
	testChannel := getTestChannel()

	input := &tg.InputChannel{
		ChannelID:  testChannel.ID,
		AccessHash: testChannel.AccessHash,
	}
	mock.ExpectCall(&tg.ChannelsGetChannelsRequest{
		ID: []tg.InputChannelClass{input},
	}).ThenResult(&tg.MessagesChats{Chats: []tg.ChatClass{testChannel}})

	v, err := m.getChannel(ctx, input)
	a.NoError(err)
	a.Equal(testChannel, v)

	v, err = m.getChannel(ctx, input)
	a.NoError(err)
	a.Equal(testChannel, v)
}

func TestManager_updateChannel(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)
	input := &tg.InputChannel{
		ChannelID:  10,
		AccessHash: 10,
	}

	mock.ExpectCall(&tg.ChannelsGetChannelsRequest{
		ID: []tg.InputChannelClass{input},
	}).ThenRPCErr(getTestError())

	_, err := m.updateChannel(ctx, input)
	a.Error(err)

	mock.ExpectCall(&tg.ChannelsGetChannelsRequest{
		ID: []tg.InputChannelClass{input},
	}).ThenResult(&tg.MessagesChats{})

	_, err = m.updateChannel(ctx, input)
	a.Error(err)

	mock.ExpectCall(&tg.ChannelsGetChannelsRequest{
		ID: []tg.InputChannelClass{input},
	}).ThenResult(&tg.MessagesChats{Chats: []tg.ChatClass{&tg.Chat{}}})

	_, err = m.updateChannel(ctx, input)
	a.Error(err)
}
