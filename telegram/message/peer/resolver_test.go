package peer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgmock"
)

func Test_plainResolver_Resolve(t *testing.T) {
	mock := tgmock.NewMock(t, require.New(t))
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
	mock.IsType(&tg.InputPeerUser{}, r)
	mock.Equal(10, r.(*tg.InputPeerUser).UserID)
	mock.NoError(err)

	_, err = resolver.ResolveDomain(ctx, domain)
	mock.Error(err)
}

func Test_plainResolver_ResolvePhone(t *testing.T) {
	mock := tgmock.NewMock(t, require.New(t))
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
	mock.NoError(err)
	mock.IsType(&tg.InputPeerUser{}, r)
	mock.Equal(10, r.(*tg.InputPeerUser).UserID)

	_, err = resolver.ResolvePhone(ctx, phone)
	mock.Error(err)
}
