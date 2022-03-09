package members

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgmock"
)

func testManager(t *testing.T) (*tgmock.Mock, *peers.Manager) {
	mock := tgmock.New(t)
	return mock, peers.Options{
		Logger: zaptest.NewLogger(t),
		Cache:  &peers.InmemoryCache{},
	}.Build(tg.NewClient(mock))
}

func getTestChannel() *tg.Channel {
	return &tg.Channel{
		Broadcast:           true,
		Noforwards:          true,
		ID:                  11,
		AccessHash:          11,
		Title:               "I hate mondays",
		Username:            "",
		Photo:               &tg.ChatPhotoEmpty{},
		Date:                int(time.Now().Unix()),
		RestrictionReason:   nil,
		AdminRights:         tg.ChatAdminRights{},
		BannedRights:        tg.ChatBannedRights{},
		DefaultBannedRights: tg.ChatBannedRights{},
		ParticipantsCount:   1,
	}
}

func getTestChannelFull() *tg.ChannelFull {
	u := &tg.ChannelFull{
		CanViewParticipants: true,
		HasScheduled:        true,
		ID:                  11,
		About:               "garfield blog",
		ParticipantsCount:   1,
		ChatPhoto:           &tg.PhotoEmpty{},
	}
	u.SetFlags()
	return u
}

func getTestChat() *tg.Chat {
	u := &tg.Chat{
		Noforwards:        true,
		ID:                10,
		Title:             "I hate mondays",
		ParticipantsCount: 1,
		Date:              int(time.Now().Unix()),
		Version:           1,
		Photo:             &tg.ChatPhotoEmpty{},
	}
	u.SetFlags()
	return u
}

func getTestChatFull(participants tg.ChatParticipantsClass) *tg.ChatFull {
	u := &tg.ChatFull{
		CanSetUsername: false,
		HasScheduled:   true,
		ID:             10,
		About:          "garfield blog",
		Participants:   participants,
	}
	u.SetFlags()
	return u
}

func getTestUser() *tg.User {
	u := &tg.User{
		Self:       false,
		Bot:        false,
		ID:         11,
		AccessHash: 10,
		FirstName:  "Julia",
		LastName:   "Ann",
		Username:   "aboba",
	}
	u.SetFlags()
	return u
}

func getTestError() *tgerr.Error {
	return &tgerr.Error{
		Code:    1337,
		Message: "TEST_ERROR",
		Type:    "TEST_ERROR",
	}
}

func TestEditRights(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	rights := tg.ChatBannedRights{
		SendInline: true,
	}
	rights.SetFlags()
	req := func(p Members) *tgmock.RequestBuilder {
		return mock.ExpectCall(&tg.MessagesEditChatDefaultBannedRightsRequest{
			Peer:         p.Peer().InputPeer(),
			BannedRights: rights,
		})
	}
	for _, p := range []Members{
		Chat(m.Chat(getTestChat())),
		Channel(m.Channel(getTestChannel())),
	} {
		req(p).ThenRPCErr(getTestError())
		a.Error(p.EditRights(ctx, MemberRights{DenySendInline: true}))
		req(p).ThenResult(&tg.Updates{})
		a.NoError(p.EditRights(ctx, MemberRights{DenySendInline: true}))
	}
}
