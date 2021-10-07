package telegram

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/tg"
)

func TestClient_Self(t *testing.T) {
	ctx := context.Background()
	self, err := newTestClient(func(id int64, body bin.Encoder) (bin.Encoder, error) {
		assert.IsType(t, &tg.UsersGetUsersRequest{}, body)
		u := &tg.User{ID: 1}
		u.SetBot(true)
		return &tg.UserClassVector{
			Elems: []tg.UserClass{u},
		}, nil
	}).Self(ctx)
	require.NoError(t, err)

	expected := &tg.User{ID: 1}
	expected.SetBot(true)
	require.Equal(t, expected, self)
}
