package message

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func TestSender_JoinLink(t *testing.T) {
	formats := []struct {
		fmt     string
		wantErr bool
	}{
		{`t.me/joinchat/%s`, false},
		{`t.me/joinchat/%s/`, false},
		{`https://t.me/joinchat/%s`, false},
		{`https://t.me/joinchat/%s/`, false},
		{`tg:join?invite=%s`, false},
		{`tg://join?invite=%s`, false},
	}
	inputs := []struct {
		value   string
		wantErr bool
	}{
		{"AAAAAAAAAAAAAAAAAA", false},
		{"", true},
	}

	for _, format := range formats {
		for _, input := range inputs {
			link := fmt.Sprintf(format.fmt, input.value)
			t.Run(link, func(t *testing.T) {
				ctx := context.Background()
				sender, mock := testSender(t)

				wantErr := format.wantErr || input.wantErr
				if !wantErr {
					mock.ExpectCall(&tg.MessagesImportChatInviteRequest{
						Hash: input.value,
					}).ThenResult(&tg.Updates{})
				}

				_, err := sender.JoinLink(ctx, link)
				if wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			})
		}
	}
}

func TestRequestBuilder_Join(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	peer := &tg.InputPeerChannel{
		ChannelID:  10,
		AccessHash: 10,
	}

	mock.ExpectCall(&tg.ChannelsJoinChannelRequest{
		Channel: &tg.InputChannel{
			ChannelID:  10,
			AccessHash: 10,
		},
	}).ThenResult(&tg.Updates{})
	_, err := sender.To(peer).Join(ctx)
	require.NoError(t, err)

	mock.ExpectCall(&tg.ChannelsJoinChannelRequest{
		Channel: &tg.InputChannel{
			ChannelID:  10,
			AccessHash: 10,
		},
	}).ThenRPCErr(testRPCError())
	_, err = sender.To(peer).Join(ctx)
	require.Error(t, err)
}

func TestRequestBuilder_Leave(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	peer := &tg.InputPeerChannel{
		ChannelID:  10,
		AccessHash: 10,
	}
	ch := &tg.InputChannel{
		ChannelID:  peer.ChannelID,
		AccessHash: peer.AccessHash,
	}

	mock.ExpectCall(&tg.ChannelsLeaveChannelRequest{
		Channel: ch,
	}).ThenResult(&tg.Updates{})
	_, err := sender.To(peer).Leave(ctx)
	require.NoError(t, err)

	mock.ExpectCall(&tg.ChannelsLeaveChannelRequest{
		Channel: ch,
	}).ThenRPCErr(testRPCError())
	_, err = sender.To(peer).Leave(ctx)
	require.Error(t, err)
}
