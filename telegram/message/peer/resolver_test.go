package peer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/telegram/internal/rpcmock"
	"github.com/gotd/td/tg"
)

func Test_plainResolver_Resolve(t *testing.T) {
	mock := rpcmock.NewMock(t, require.New(t))
	raw := tg.NewClient(mock)

	domain := "adcd"
	mock.ExpectCall(&tg.ContactsResolveUsernameRequest{
		Username: domain,
	}).ThenResult(&tg.ContactsResolvedPeer{
		Peer: &tg.PeerUser{UserID: 10},
		Users: []tg.UserClass{
			&tg.User{ID: 10, AccessHash: 10, Username: domain},
		},
	}).ExpectCall(&tg.ContactsResolveUsernameRequest{
		Username: domain,
	}).ThenRPCErr(&mtproto.Error{
		Code:    1337,
		Message: "TEST_ERR",
		Type:    "TEST_ERR",
	})

	ctx := context.Background()
	resolver := plainResolver{raw: raw}

	r, err := resolver.Resolve(ctx, domain)
	mock.IsType(&tg.InputPeerUser{}, r)
	mock.Equal(10, r.(*tg.InputPeerUser).UserID)
	mock.NoError(err)

	_, err = resolver.Resolve(ctx, domain)
	mock.Error(err)
}
