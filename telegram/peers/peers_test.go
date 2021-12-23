package peers

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestReport(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	reason := &tg.InputReportReasonSpam{}
	message := "message"
	peers := []Peer{
		m.User(getTestUser()),
		m.Chat(getTestChat()),
		m.Channel(getTestChannel()),
	}

	for _, p := range peers {
		req := &tg.AccountReportPeerRequest{
			Peer:    p.InputPeer(),
			Reason:  reason,
			Message: message,
		}

		mock.ExpectCall(req).ThenRPCErr(getTestError())
		a.Error(p.Report(ctx, reason, message))

		mock.ExpectCall(req).ThenTrue()
		a.NoError(p.Report(ctx, reason, message))
	}
}

func TestManager_FromInputPeer(t *testing.T) {
	testUser := getTestUser()
	testChat := getTestChat()
	testChannel := getTestChannel()

	getUser := func(input tg.InputUserClass) *tg.UsersGetUsersRequest {
		return &tg.UsersGetUsersRequest{
			ID: []tg.InputUserClass{input},
		}
	}
	getChannel := func(input tg.InputChannelClass) *tg.ChannelsGetChannelsRequest {
		return &tg.ChannelsGetChannelsRequest{
			ID: []tg.InputChannelClass{input},
		}
	}
	var tests = []struct {
		input   tg.InputPeerClass
		expect  bin.Encoder
		result  bin.Encoder
		wantErr bool
	}{
		{
			&tg.InputPeerSelf{},
			getUser(&tg.InputUserSelf{}),
			&tg.UserClassVector{Elems: []tg.UserClass{getTestSelf()}},
			false,
		},
		{
			testUser.AsInputPeer(),
			getUser(testUser.AsInput()),
			&tg.UserClassVector{Elems: []tg.UserClass{testUser}},
			false,
		},
		{
			&tg.InputPeerUserFromMessage{
				Peer:   getTestChannel().AsInputPeer(),
				MsgID:  10,
				UserID: testUser.ID,
			},
			getUser(&tg.InputUserFromMessage{
				Peer:   getTestChannel().AsInputPeer(),
				MsgID:  10,
				UserID: testUser.ID,
			}),
			&tg.UserClassVector{Elems: []tg.UserClass{testUser}},
			false,
		},
		{
			testChannel.AsInputPeer(),
			getChannel(testChannel.AsInput()),
			&tg.MessagesChats{Chats: []tg.ChatClass{testChannel}},
			false,
		},
		{
			&tg.InputPeerChannelFromMessage{
				Peer:      getTestChannel().AsInputPeer(),
				MsgID:     10,
				ChannelID: testChannel.ID,
			},
			getChannel(&tg.InputChannelFromMessage{
				Peer:      getTestChannel().AsInputPeer(),
				MsgID:     10,
				ChannelID: testChannel.ID,
			}),
			&tg.MessagesChats{Chats: []tg.ChatClass{testChannel}},
			false,
		},
		{
			&tg.InputPeerChat{
				ChatID: testChat.ID,
			},
			&tg.MessagesGetChatsRequest{
				ID: []int64{testChat.ID},
			},
			&tg.MessagesChats{Chats: []tg.ChatClass{testChat}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T", tt.input), func(t *testing.T) {
			a := require.New(t)
			ctx := context.Background()
			mock, m := testManager(t)

			mock.ExpectCall(tt.expect).ThenRPCErr(getTestError())
			_, err := m.FromInputPeer(ctx, tt.input)
			a.Error(err)

			mock.ExpectCall(tt.expect).ThenResult(tt.result)
			p, err := m.FromInputPeer(ctx, tt.input)
			if tt.wantErr {
				a.Error(err)
				return
			}
			a.NoError(err)
			a.NotZero(p)
		})
	}
}
