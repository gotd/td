package telegram

import (
	"bytes"
	"context"
	"testing"

	"github.com/gotd/td/telegram/internal/rpcmock"
	"github.com/gotd/td/tg"
)

func TestTransfer(t *testing.T) {
	ctx := context.Background()

	dc := 1
	mockClient(func(a *rpcmock.Mock, client *Client) {
		user := &tg.User{ID: 10, Username: "abc10"}
		auth := &tg.AuthAuthorization{
			User: user,
		}
		exported := bytes.Repeat([]byte{10}, 10)
		a.ExpectCall(&tg.AuthExportAuthorizationRequest{
			DCID: dc,
		}).ThenResult(&tg.AuthExportedAuthorization{
			ID:    user.ID,
			Bytes: exported,
		}).ExpectCall(&tg.AuthImportAuthorizationRequest{
			ID:    user.ID,
			Bytes: exported,
		}).ThenResult(&tg.AuthAuthorization{
			User: user,
		})

		r, err := client.transfer(ctx, tg.NewClient(client), dc)
		a.NoError(err)
		a.Equal(auth, r)
	})(t)
}
