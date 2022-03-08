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
