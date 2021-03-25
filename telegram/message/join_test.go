package message

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
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
					mock.Error(err)
				} else {
					mock.NoError(err)
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
	mock.NoError(err)

	mock.ExpectCall(&tg.ChannelsJoinChannelRequest{
		Channel: &tg.InputChannel{
			ChannelID:  10,
			AccessHash: 10,
		},
	}).ThenRPCErr(testRPCError())
	_, err = sender.To(peer).Join(ctx)
	mock.Error(err)
}

func TestRequestBuilder_Leave(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	peer := &tg.InputPeerChannel{
		ChannelID:  10,
		AccessHash: 10,
	}

	mock.ExpectCall(&tg.ChannelsLeaveChannelRequest{
		Channel: &tg.InputChannel{
			ChannelID:  10,
			AccessHash: 10,
		},
	}).ThenResult(&tg.Updates{})
	_, err := sender.To(peer).Leave(ctx)
	mock.NoError(err)

	mock.ExpectCall(&tg.ChannelsLeaveChannelRequest{
		Channel: &tg.InputChannel{
			ChannelID:  10,
			AccessHash: 10,
		},
	}).ThenRPCErr(testRPCError())
	_, err = sender.To(peer).Leave(ctx)
	mock.Error(err)
}

func Test_inputChannel(t *testing.T) {
	tests := []struct {
		input   tg.InputPeerClass
		output  tg.InputChannelClass
		wantErr bool
	}{
		{&tg.InputPeerChannel{
			ChannelID:  10,
			AccessHash: 10,
		}, &tg.InputChannel{
			ChannelID:  10,
			AccessHash: 10,
		}, false},
		{&tg.InputPeerChannelFromMessage{
			Peer:      &tg.InputPeerSelf{},
			MsgID:     10,
			ChannelID: 10,
		}, &tg.InputChannelFromMessage{
			Peer:      &tg.InputPeerSelf{},
			MsgID:     10,
			ChannelID: 10,
		}, false},
		{&tg.InputPeerChat{ChatID: 10}, nil, true},
	}

	a := require.New(t)
	for _, test := range tests {
		ch, err := inputChannel(test.input)
		if test.wantErr {
			a.Error(err)
		} else {
			a.NoError(err)
			a.Equal(test.output, ch)
		}
	}
}
