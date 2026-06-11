package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestChannel_SetUsername(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	const username = "gotd_channel"
	ch := m.Channel(getTestChannel())

	mock.ExpectCall(&tg.ChannelsUpdateUsernameRequest{
		Channel:  ch.InputChannel(),
		Username: username,
	}).ThenRPCErr(getTestError())
	a.Error(ch.SetUsername(ctx, username))

	mock.ExpectCall(&tg.ChannelsUpdateUsernameRequest{
		Channel:  ch.InputChannel(),
		Username: username,
	}).ThenTrue()
	a.NoError(ch.SetUsername(ctx, username))
}

func TestChannel_CheckUsername(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	const username = "gotd_channel"
	ch := m.Channel(getTestChannel())

	mock.ExpectCall(&tg.ChannelsCheckUsernameRequest{
		Channel:  ch.InputChannel(),
		Username: username,
	}).ThenRPCErr(getTestError())
	_, err := ch.CheckUsername(ctx, username)
	a.Error(err)

	mock.ExpectCall(&tg.ChannelsCheckUsernameRequest{
		Channel:  ch.InputChannel(),
		Username: username,
	}).ThenTrue()
	ok, err := ch.CheckUsername(ctx, username)
	a.NoError(err)
	a.True(ok)
}

func TestChannel_DeactivateAllUsernames(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	ch := m.Channel(getTestChannel())

	mock.ExpectCall(&tg.ChannelsDeactivateAllUsernamesRequest{
		Channel: ch.InputChannel(),
	}).ThenRPCErr(getTestError())
	a.Error(ch.DeactivateAllUsernames(ctx))

	mock.ExpectCall(&tg.ChannelsDeactivateAllUsernamesRequest{
		Channel: ch.InputChannel(),
	}).ThenTrue()
	a.NoError(ch.DeactivateAllUsernames(ctx))
}

func TestChannel_SetPhoto(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	ch := m.Channel(getTestChannel())
	photo := &tg.InputChatUploadedPhoto{}
	photo.SetFlags()

	mock.ExpectCall(&tg.ChannelsEditPhotoRequest{
		Channel: ch.InputChannel(),
		Photo:   photo,
	}).ThenRPCErr(getTestError())
	a.Error(ch.SetPhoto(ctx, photo))

	mock.ExpectCall(&tg.ChannelsEditPhotoRequest{
		Channel: ch.InputChannel(),
		Photo:   photo,
	}).ThenResult(&tg.Updates{})
	a.NoError(ch.SetPhoto(ctx, photo))

	mock.ExpectCall(&tg.ChannelsEditPhotoRequest{
		Channel: ch.InputChannel(),
		Photo:   &tg.InputChatPhotoEmpty{},
	}).ThenResult(&tg.Updates{})
	a.NoError(ch.DeletePhoto(ctx))
}

func TestSupergroup_TogglePreHistoryHidden(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	s, ok := m.Channel(getTestSuperGroup()).ToSupergroup()
	a.True(ok)

	mock.ExpectCall(&tg.ChannelsTogglePreHistoryHiddenRequest{
		Channel: s.InputChannel(),
		Enabled: true,
	}).ThenRPCErr(getTestError())
	a.Error(s.TogglePreHistoryHidden(ctx, true))

	mock.ExpectCall(&tg.ChannelsTogglePreHistoryHiddenRequest{
		Channel: s.InputChannel(),
		Enabled: true,
	}).ThenResult(&tg.Updates{})
	a.NoError(s.TogglePreHistoryHidden(ctx, true))
}

func TestSupergroup_ToggleJoinToSend(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	s, ok := m.Channel(getTestSuperGroup()).ToSupergroup()
	a.True(ok)

	mock.ExpectCall(&tg.ChannelsToggleJoinToSendRequest{
		Channel: s.InputChannel(),
		Enabled: true,
	}).ThenRPCErr(getTestError())
	a.Error(s.ToggleJoinToSend(ctx, true))

	mock.ExpectCall(&tg.ChannelsToggleJoinToSendRequest{
		Channel: s.InputChannel(),
		Enabled: true,
	}).ThenResult(&tg.Updates{})
	a.NoError(s.ToggleJoinToSend(ctx, true))
}

func TestSupergroup_ToggleJoinRequest(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	s, ok := m.Channel(getTestSuperGroup()).ToSupergroup()
	a.True(ok)

	mock.ExpectCall(&tg.ChannelsToggleJoinRequestRequest{
		Channel: s.InputChannel(),
		Enabled: true,
	}).ThenRPCErr(getTestError())
	a.Error(s.ToggleJoinRequest(ctx, true))

	mock.ExpectCall(&tg.ChannelsToggleJoinRequestRequest{
		Channel: s.InputChannel(),
		Enabled: true,
	}).ThenResult(&tg.Updates{})
	a.NoError(s.ToggleJoinRequest(ctx, true))
}

func TestChannel_AdminLog(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	ch := m.Channel(getTestChannel())

	filter := tg.ChannelAdminLogEventsFilter{Ban: true, Promote: true}
	filter.SetFlags()

	expect := &tg.ChannelsGetAdminLogRequest{
		Channel: ch.InputChannel(),
		Q:       "spam",
		MaxID:   0,
		MinID:   0,
		Limit:   100,
	}
	expect.SetEventsFilter(filter)

	mock.ExpectCall(expect).ThenRPCErr(getTestError())
	a.Error(ch.AdminLog().Search("spam").Filter(filter).ForEach(ctx, func(event tg.ChannelAdminLogEvent) error {
		return nil
	}))

	// First page returns one event, second page is empty (terminates iteration).
	mock.ExpectCall(expect).ThenResult(&tg.ChannelsAdminLogResults{
		Events: []tg.ChannelAdminLogEvent{{
			ID:     42,
			Action: &tg.ChannelAdminLogEventActionToggleSlowMode{},
		}},
	})
	next := &tg.ChannelsGetAdminLogRequest{
		Channel: ch.InputChannel(),
		Q:       "spam",
		MaxID:   42,
		MinID:   0,
		Limit:   100,
	}
	next.SetEventsFilter(filter)
	mock.ExpectCall(next).ThenResult(&tg.ChannelsAdminLogResults{})

	var got []int64
	a.NoError(ch.AdminLog().Search("spam").Filter(filter).ForEach(ctx, func(event tg.ChannelAdminLogEvent) error {
		got = append(got, event.ID)
		return nil
	}))
	a.Equal([]int64{42}, got)
}
