package peer

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// Resolver is an abstraction to resolve domains and Telegram deeplinks.
type Resolver interface {
	ResolveDomain(ctx context.Context, domain string) (tg.InputPeerClass, error)
	ResolvePhone(ctx context.Context, phone string) (tg.InputPeerClass, error)
}

// DefaultResolver creates and returns default resolver.
func DefaultResolver(raw *tg.Client) Resolver {
	return NewLRUResolver(SingleflightResolver(Plain(raw)), 10)
}

// Plain creates plain resolver.
func Plain(raw *tg.Client) Resolver {
	return plainResolver{raw: raw}
}

type plainResolver struct {
	raw *tg.Client
}

func (p plainResolver) ResolveDomain(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	peer, err := p.raw.ContactsResolveUsername(ctx, domain)
	if err != nil {
		return nil, errors.Wrap(err, "resolve")
	}

	return EntitiesFromResult(peer).ExtractPeer(peer.Peer)
}

func (p plainResolver) ResolvePhone(ctx context.Context, phone string) (tg.InputPeerClass, error) {
	r, err := p.raw.ContactsGetContacts(ctx, 0)
	if err != nil {
		return nil, errors.Wrap(err, "get contacts")
	}

	switch c := r.(type) {
	case *tg.ContactsContacts:
		for _, u := range c.Users {
			user, ok := u.AsNotEmpty()
			if !ok {
				continue
			}
			if user.Phone == phone {
				return user.AsInputPeer(), nil
			}
		}

		return nil, errors.Errorf("can't resolve phone %q", phone)
	default:
		return nil, errors.Errorf("unexpected type %T", r)
	}
}
