package message

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

type testResolver struct {
	t *testing.T

	expectedDomain string
	expected       tg.InputPeerClass
}

func (r *testResolver) ResolveDomain(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	return r.expectResolve(ctx, domain)
}

func (r *testResolver) ResolvePhone(ctx context.Context, phone string) (tg.InputPeerClass, error) {
	return r.expectResolve(ctx, phone)
}

func (r *testResolver) expectResolve(_ context.Context, domain string) (tg.InputPeerClass, error) {
	if domain != r.expectedDomain {
		err := fmt.Errorf("expected domain %q, got %q", r.expectedDomain, domain)
		r.t.Error(err)
		return nil, err
	}
	return r.expected, nil
}

func resolver(t *testing.T, expectedDomain string, expected tg.InputPeerClass) peer.Resolver {
	return &testResolver{t, expectedDomain, expected}
}

type answerable struct {
	ID     int
	UserID int
}

func (a answerable) GetMessage() tg.MessageClass {
	return &tg.Message{
		ID:     a.ID,
		PeerID: &tg.PeerUser{UserID: a.UserID},
	}
}

func (a answerable) GetPts() int {
	return -1
}

func TestResolve(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	expected := &tg.InputPeerUser{UserID: 10, AccessHash: 10}
	s := NewSender(nil).WithResolver(resolver(t, "durov", expected))

	check := func(req *RequestBuilder, expected tg.InputPeerClass) {
		p, err := req.AsInputPeer(ctx)
		a.NoError(err)
		a.Equal(expected, p)
	}

	check(s.Self(), &tg.InputPeerSelf{})
	check(s.Peer(expected), expected)
	check(s.Resolve("durov"), expected)
	check(s.ResolveDomain("@durov"), expected)
	check(s.ResolveDeeplink("https://t.me/durov"), expected)

	uctx := tg.Entities{Users: map[int]*tg.User{
		expected.UserID: {ID: expected.UserID, AccessHash: expected.AccessHash, Username: "durov"},
	}}
	check(s.Answer(uctx, answerable{ID: 10, UserID: expected.UserID}), expected)
}

func TestSender_ResolvePhone(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	expected := &tg.InputPeerChannel{ChannelID: 10, AccessHash: 10}
	s := NewSender(nil).WithResolver(resolver(t, "13115552368", expected))

	check := func(req *RequestBuilder, expected tg.InputPeerClass) {
		p, err := req.AsInputPeer(ctx)
		a.NoError(err)
		a.Equal(expected, p)
	}

	// If there's somethin' strange
	// in your neighborhood
	check(s.Resolve("+13115552368"), expected)
	// Who ya gonna call
	// Ghostb...!
	check(s.Resolve("+1 (311) 555-2368"), expected)
}
