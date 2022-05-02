package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestManager_GetAllChats(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	mock.ExpectCall(&tg.MessagesGetAllChatsRequest{}).
		ThenRPCErr(getTestError())
	_, err := m.GetAllChats(ctx)
	a.Error(err)

	mock.ExpectCall(&tg.MessagesGetAllChatsRequest{}).
		ThenResult(&tg.MessagesChats{
			Chats: []tg.ChatClass{
				getTestChat(),
				getTestSuperGroup(),
				getTestBroadcast(),
			},
		})
	r, err := m.GetAllChats(ctx)
	a.NoError(err)
	a.Len(r.Chats, 1)
	a.Len(r.Channels, 2)
}
