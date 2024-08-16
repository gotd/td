package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func getTestBroadcast() *tg.Channel {
	testChannel := getTestChannel()
	testChannel.ID *= 3
	testChannel.Broadcast = true
	testChannel.Megagroup = false
	return testChannel
}

func TestBroadcast_SetDiscussionGroup(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	b := m.Channel(getTestSuperGroup())

	s, ok := m.Channel(getTestBroadcast()).ToBroadcast()
	a.True(ok)

	mock.ExpectCall(&tg.ChannelsSetDiscussionGroupRequest{
		Broadcast: s.InputChannel(),
		Group:     b.InputChannel(),
	}).ThenRPCErr(getTestError())
	a.Error(s.SetDiscussionGroup(ctx, b.InputChannel()))

	mock.ExpectCall(&tg.ChannelsSetDiscussionGroupRequest{
		Broadcast: s.InputChannel(),
		Group:     b.InputChannel(),
	}).ThenTrue()
	a.NoError(s.SetDiscussionGroup(ctx, b.InputChannel()))
}

func TestBroadcast_ToggleSignatures(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	ch := m.Channel(getTestBroadcast())

	s, ok := ch.ToBroadcast()
	a.True(ok)

	mock.ExpectCall(&tg.ChannelsToggleSignaturesRequest{
		Channel:           s.InputChannel(),
		SignaturesEnabled: true,
	}).ThenRPCErr(getTestError())
	a.Error(s.ToggleSignatures(ctx, true))

	mock.ExpectCall(&tg.ChannelsToggleSignaturesRequest{
		Channel:           s.InputChannel(),
		SignaturesEnabled: true,
	}).ThenResult(&tg.Updates{})
	a.NoError(s.ToggleSignatures(ctx, true))
}

func TestBroadcast_DiscussionGroup(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	linkedChat := getTestChat()
	linkedSupergroup := getTestSuperGroup()
	linkedChat.SetMigratedTo(linkedSupergroup.AsInput())
	ch := m.Channel(getTestBroadcast())

	s, ok := ch.ToBroadcast()
	a.True(ok)

	mock.ExpectCall(&tg.ChannelsGetFullChannelRequest{
		Channel: s.InputChannel(),
	}).ThenRPCErr(getTestError())
	_, ok, err := s.DiscussionGroup(ctx)
	a.False(ok)
	a.Error(err)

	full := getTestChannelFull()
	full.SetLinkedChatID(linkedChat.ID)

	mock.ExpectCall(&tg.ChannelsGetFullChannelRequest{
		Channel: s.InputChannel(),
	}).ThenResult(&tg.MessagesChatFull{
		FullChat: full,
		Chats:    []tg.ChatClass{linkedChat, linkedSupergroup},
	})
	d, ok, err := s.DiscussionGroup(ctx)
	a.True(ok)
	a.NoError(err)
	a.Equal(linkedSupergroup.ID, d.ID())
}
