package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/testutil"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func TestClient_self(t *testing.T) {
	ctx := context.Background()
	mockTest(func(a *require.Assertions, mock *tgmock.Mock, client *Client) {
		mock.ExpectCall(&tg.UsersGetUsersRequest{
			ID: []tg.InputUserClass{&tg.InputUserSelf{}},
		}).ThenErr(testutil.TestError())
		_, err := client.self(ctx)
		a.Error(err)

		mock.ExpectCall(&tg.UsersGetUsersRequest{
			ID: []tg.InputUserClass{&tg.InputUserSelf{}},
		}).ThenResult(&tg.UserClassVector{Elems: []tg.UserClass{&tg.UserEmpty{
			ID: 10,
		}}})
		_, err = client.self(ctx)
		a.Error(err)

		mock.ExpectCall(&tg.UsersGetUsersRequest{
			ID: []tg.InputUserClass{&tg.InputUserSelf{}},
		}).ThenResult(&tg.UserClassVector{Elems: []tg.UserClass{&tg.User{
			Self:       true,
			ID:         10,
			AccessHash: 10,
		}}})
		r, err := client.self(ctx)
		a.NoError(err)
		a.Equal(int64(10), r.ID)
	})(t)
}
