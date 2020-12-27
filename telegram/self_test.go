package telegram

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestClient_Self(t *testing.T) {
	ctx := context.Background()

	inv := mockInvoker(func(input *bin.Buffer) (bin.Encoder, error) {
		var req tg.UsersGetUsersRequest
		if err := req.Decode(input); err != nil {
			return nil, err
		}

		u := &tg.User{ID: 1}
		u.SetBot(true)
		return &tg.UserClassVector{
			Elems: []tg.UserClass{u},
		}, nil
	})

	client := &Client{
		RPC:     tg.NewClient(inv),
		mtp:     inv,
		appID:   1,
		appHash: "foo",
	}

	self, err := client.Self(ctx)
	require.NoError(t, err)

	expected := &tg.User{ID: 1}
	expected.SetBot(true)
	require.Equal(t, expected, self)
}
