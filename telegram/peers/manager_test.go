package peers

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgmock"
)

var _ = []interface {
	ParticipantsCount() int
	Leave(ctx context.Context) error
	SetTitle(ctx context.Context, title string) error
	SetDescription(ctx context.Context, about string) error
}{
	Chat{},
	Channel{},
}

func testManager(t *testing.T) (*tgmock.Mock, *Manager) {
	mock := tgmock.New(t)
	return mock, Options{
		Logger: zaptest.NewLogger(t),
		Cache:  &InmemoryCache{},
	}.Build(tg.NewClient(mock))
}

func getTestSelf() *tg.User {
	u := &tg.User{
		Self:       true,
		Bot:        true,
		ID:         10,
		AccessHash: 10,
		FirstName:  "Lana",
		LastName:   "Rhoades",
		Username:   "thebot",
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

func getTestUserFull() tg.UserFull {
	u := tg.UserFull{
		PhoneCallsAvailable: true,
		ID:                  11,
		About:               "hot mommy",
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

func getTestChatFull() *tg.ChatFull {
	u := &tg.ChatFull{
		CanSetUsername: false,
		HasScheduled:   true,
		ID:             10,
		About:          "garfield blog",
		Participants: &tg.ChatParticipants{
			ChatID: 10,
			Participants: []tg.ChatParticipantClass{
				&tg.ChatParticipant{
					UserID:    10,
					InviterID: 10,
					Date:      10,
				},
			},
			Version: 1,
		},
	}
	u.SetFlags()
	return u
}

func getTestChannel() *tg.Channel {
	u := &tg.Channel{
		Broadcast:           true,
		Noforwards:          true,
		ID:                  11,
		AccessHash:          0,
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
	u.SetFlags()
	return u
}

func getTestChannelFull() *tg.ChannelFull {
	u := &tg.ChannelFull{
		HasScheduled:      true,
		ID:                11,
		About:             "garfield blog",
		ParticipantsCount: 1,
		ChatPhoto:         &tg.PhotoEmpty{},
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
