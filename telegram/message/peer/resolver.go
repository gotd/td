package peer

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// Resolver is a abstraction to resolve domains and Telegram deeplinks
type Resolver interface {
	Resolve(ctx context.Context, domain string) (tg.InputPeerClass, error)
}

// ResolverFunc is a functional adapter for Resolver.
type ResolverFunc func(ctx context.Context, domain string) (tg.InputPeerClass, error)

// Resolve implements Resolver.
func (p ResolverFunc) Resolve(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	return p(ctx, domain)
}

// DefaultResolver creates and returns default resolver.
func DefaultResolver(raw *tg.Client) Resolver {
	return NewLRUResolver(&plainResolver{raw: raw}, 10)
}

type plainResolver struct {
	raw *tg.Client
}

func (p plainResolver) Resolve(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	peer, err := p.raw.ContactsResolveUsername(ctx, domain)
	if err != nil {
		return nil, xerrors.Errorf("resolve: %w", err)
	}

	return EntitiesFromResult(peer).ExtractPeer(peer.Peer)
}
