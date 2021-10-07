package peer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgerr"
	"github.com/nnqq/td/tgmock"
)

func Test_plainResolver_Resolve(t *testing.T) {
	mock := tgmock.NewRequire(t)
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
	}).ThenRPCErr(&tgerr.Error{
		Code:    1337,
		Message: "TEST_ERROR",
		Type:    "TEST_ERROR",
	})

	ctx := context.Background()
	resolver := plainResolver{raw: raw}

	r, err := resolver.ResolveDomain(ctx, domain)
	require.IsType(t, &tg.InputPeerUser{}, r)
	require.Equal(t, int64(10), r.(*tg.InputPeerUser).UserID)
	require.NoError(t, err)

	_, err = resolver.ResolveDomain(ctx, domain)
	require.Error(t, err)
}

func Test_plainResolver_ResolvePhone(t *testing.T) {
	mock := tgmock.New(t)
	raw := tg.NewClient(mock)

	phone := "adcd"
	mock.ExpectCall(&tg.ContactsGetContactsRequest{
		Hash: 0,
	}).ThenResult(&tg.ContactsContacts{
		Contacts: []tg.Contact{{
			UserID: 10,
			Mutual: false,
		}},
		SavedCount: 1,
		Users: []tg.UserClass{
			&tg.User{ID: 10, AccessHash: 10, Username: "rustmustdie", Phone: phone},
		},
	}).Expect().ThenRPCErr(&tgerr.Error{
		Code:    1337,
		Message: "TEST_ERROR",
		Type:    "TEST_ERROR",
	})

	ctx := context.Background()
	resolver := plainResolver{raw: raw}

	r, err := resolver.ResolvePhone(ctx, phone)
	require.NoError(t, err)
	require.IsType(t, &tg.InputPeerUser{}, r)
	require.Equal(t, int64(10), r.(*tg.InputPeerUser).UserID)

	_, err = resolver.ResolvePhone(ctx, phone)
	require.Error(t, err)
}
