package members

import (
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
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
