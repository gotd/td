package message

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

func resolver(t *testing.T, expectedDomain string, expected tg.InputPeerClass) peer.ResolverFunc {
	return func(ctx context.Context, domain string) (tg.InputPeerClass, error) {
		if domain != expectedDomain {
			err := fmt.Errorf("expected domain %q, got %q", expectedDomain, domain)
			t.Error(err)
			return nil, err
		}
		return expected, nil
	}
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

	uctx := tg.UpdateContext{Users: map[int]*tg.User{
		expected.UserID: {ID: expected.UserID, AccessHash: expected.AccessHash, Username: "durov"},
	}}
	check(s.Answer(uctx, answerable{ID: 10, UserID: expected.UserID}), expected)
}
