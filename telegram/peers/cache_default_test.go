package peers

import (
	"context"
	"testing"
	"time"

	"github.com/gotd/neo"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgmock"
)

// PEER-1: with default options a peer already seen (e.g. in an update) is
// resolved from the cache, with no users.getUsers.
func TestDefaultCacheNoRPC(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.New(t)
	m := Options{}.Build(tg.NewClient(mock))
	u := getTestUser()
	require.NoError(t, m.Apply(ctx, []tg.UserClass{u}, nil))
	for i := 0; i < 2; i++ {
		p, err := m.ResolvePeer(ctx, &tg.PeerUser{UserID: u.ID})
		require.NoError(t, err) // the mock fails on any call
		require.Equal(t, u.ID, p.ID())
	}
}

// NoopCache stays available and fetches every time.
func TestNoopCacheFetches(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.New(t)
	m := Options{Cache: NoopCache{}}.Build(tg.NewClient(mock))
	u := getTestUser()
	require.NoError(t, m.Apply(ctx, []tg.UserClass{u}, nil))
	mock.ExpectCall(&tg.UsersGetUsersRequest{ID: []tg.InputUserClass{&tg.InputUser{UserID: u.ID, AccessHash: u.AccessHash}}}).
		ThenResult(&tg.UserClassVector{Elems: []tg.UserClass{u}})
	_, err := m.ResolvePeer(ctx, &tg.PeerUser{UserID: u.ID})
	require.NoError(t, err)
}

func resolvedUser(u *tg.User) *tg.ContactsResolvedPeer {
	return &tg.ContactsResolvedPeer{Peer: &tg.PeerUser{UserID: u.ID}, Users: []tg.UserClass{u}}
}

// PEER-2: spellings of one username make one contacts.resolveUsername.
func TestResolveDomainCached(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.New(t)
	m := Options{}.Build(tg.NewClient(mock))
	u := getTestUser()
	u.Username = "Aboba"
	mock.ExpectCall(&tg.ContactsResolveUsernameRequest{Username: "aboba"}).ThenResult(resolvedUser(u))
	for _, name := range []string{"aboba", "@aboba", "Aboba", "ABOBA"} {
		p, err := m.ResolveDomain(ctx, name)
		require.NoError(t, err, name)
		require.Equal(t, u.ID, p.ID())
	}
}

// A username of an applied peer (seen in an update) needs no RPC at all,
// including its active additional usernames; inactive ones are not indexed.
func TestResolveDomainFromApplied(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.New(t)
	m := Options{}.Build(tg.NewClient(mock))
	u := getTestUser()
	u.Usernames = []tg.Username{{Username: "collectible", Active: true}, {Username: "disabled"}}
	require.NoError(t, m.Apply(ctx, []tg.UserClass{u}, nil))
	for _, name := range []string{"aboba", "collectible"} {
		p, err := m.ResolveDomain(ctx, name)
		require.NoError(t, err, name)
		require.Equal(t, u.ID, p.ID())
	}
	mock.ExpectCall(&tg.ContactsResolveUsernameRequest{Username: "disabled"}).
		ThenRPCErr(&tgerr.Error{Code: 400, Message: "USERNAME_NOT_OCCUPIED", Type: "USERNAME_NOT_OCCUPIED"})
	_, err := m.ResolveDomain(ctx, "disabled")
	require.True(t, tgerr.Is(err, "USERNAME_NOT_OCCUPIED"))
}

// A rename drops the old name: it is resolved again.
func TestResolveDomainRename(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.New(t)
	m := Options{}.Build(tg.NewClient(mock))
	u := getTestUser()
	require.NoError(t, m.Apply(ctx, []tg.UserClass{u}, nil))
	renamed := getTestUser()
	renamed.Username = "newname"
	require.NoError(t, m.Apply(ctx, []tg.UserClass{renamed}, nil))

	p, err := m.ResolveDomain(ctx, "newname")
	require.NoError(t, err)
	require.Equal(t, u.ID, p.ID())

	mock.ExpectCall(&tg.ContactsResolveUsernameRequest{Username: "aboba"}).
		ThenRPCErr(&tgerr.Error{Code: 400, Message: "USERNAME_NOT_OCCUPIED", Type: "USERNAME_NOT_OCCUPIED"})
	_, err = m.ResolveDomain(ctx, "aboba")
	require.Error(t, err)
}

// Cached usernames expire after UsernameTTL; USERNAME_NOT_OCCUPIED drops them.
func TestResolveDomainExpiry(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.New(t)
	clk := neo.NewTime(time.Unix(1_700_000_000, 0))
	m := Options{Clock: clk}.Build(tg.NewClient(mock))
	u := getTestUser()
	mock.ExpectCall(&tg.ContactsResolveUsernameRequest{Username: "aboba"}).ThenResult(resolvedUser(u))
	_, err := m.ResolveDomain(ctx, "aboba")
	require.NoError(t, err)

	clk.Travel(UsernameTTL - time.Second)
	_, err = m.ResolveDomain(ctx, "aboba")
	require.NoError(t, err, "still cached")

	clk.Travel(2 * time.Second)
	mock.ExpectCall(&tg.ContactsResolveUsernameRequest{Username: "aboba"}).
		ThenRPCErr(&tgerr.Error{Code: 400, Message: "USERNAME_NOT_OCCUPIED", Type: "USERNAME_NOT_OCCUPIED"})
	_, err = m.ResolveDomain(ctx, "aboba")
	require.Error(t, err, "expired: resolved again")

	mock.ExpectCall(&tg.ContactsResolveUsernameRequest{Username: "aboba"}).
		ThenRPCErr(&tgerr.Error{Code: 400, Message: "USERNAME_NOT_OCCUPIED", Type: "USERNAME_NOT_OCCUPIED"})
	_, err = m.ResolveDomain(ctx, "aboba")
	require.Error(t, err, "dropped after USERNAME_NOT_OCCUPIED")
}

func TestNormalizeUsername(t *testing.T) {
	for in, want := range map[string]string{
		"aboba":    "aboba",
		"@Aboba":   "aboba",
		" @ABOBA ": "aboba",
		"a.bo.ba":  "aboba",
		"":         "",
	} {
		require.Equal(t, want, NormalizeUsername(in), in)
	}
}
