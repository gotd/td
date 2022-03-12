package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestManager_Search(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	mock.ExpectCall(&tg.ContactsSearchRequest{
		Q:     "q",
		Limit: 10,
	}).ThenRPCErr(getTestError())
	_, err := m.Search(ctx, "q")
	a.Error(err)

	mock.ExpectCall(&tg.ContactsSearchRequest{
		Q:     "q",
		Limit: 10,
	}).ThenResult(&tg.ContactsFound{
		MyResults: []tg.PeerClass{&tg.PeerUser{UserID: getTestUser().GetID()}},
		Results:   []tg.PeerClass{&tg.PeerChat{ChatID: getTestChat().GetID()}},
		Chats: []tg.ChatClass{
			getTestChat(),
		},
		Users: []tg.UserClass{
			getTestUser(),
		},
	})
	r, err := m.Search(ctx, "q")
	a.NoError(err)

	a.Len(r.MyResults, 1)
	a.Equal(getTestUser().GetID(), r.MyResults[0].ID())

	a.Len(r.Results, 1)
	a.Equal(getTestChat().GetID(), r.Results[0].ID())
}
