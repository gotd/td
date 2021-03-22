package peer

import (
	"context"
	"sort"
	"sync"

	"go.uber.org/atomic"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/query/hasher"
	"github.com/gotd/td/tg"
)

// Resolver is a abstraction to resolve domains and Telegram deeplinks.
type Resolver interface {
	Resolve(ctx context.Context, domain string) (tg.InputPeerClass, error)
	ResolvePhone(ctx context.Context, phone string) (tg.InputPeerClass, error)
}

// ResolverFunc is functional adapter for Resolver.
type ResolverFunc func(ctx context.Context, phone bool, s string) (tg.InputPeerClass, error)

// Resolve implements Resolver.
func (r ResolverFunc) Resolve(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	return r(ctx, false, domain)
}

// ResolvePhone implements Resolver.
func (r ResolverFunc) ResolvePhone(ctx context.Context, phone string) (tg.InputPeerClass, error) {
	return r(ctx, false, phone)
}

// DefaultResolver creates and returns default resolver.
func DefaultResolver(raw *tg.Client) Resolver {
	return NewLRUResolver(&plainResolver{raw: raw}, 10)
}

type plainResolver struct {
	raw *tg.Client

	contactsMux  sync.Mutex
	contacts     *tg.ContactsContacts
	contactsHash atomic.Int32
}

func (p *plainResolver) Resolve(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	peer, err := p.raw.ContactsResolveUsername(ctx, domain)
	if err != nil {
		return nil, xerrors.Errorf("resolve: %w", err)
	}

	return EntitiesFromResult(peer).ExtractPeer(peer.Peer)
}

func (p *plainResolver) ResolvePhone(ctx context.Context, phone string) (tg.InputPeerClass, error) {
	r, err := p.raw.ContactsGetContacts(ctx, int(p.contactsHash.Load()))
	if err != nil {
		return nil, xerrors.Errorf("get contacts: %w", err)
	}

	p.contactsMux.Lock()
	defer p.contactsMux.Unlock()

	switch c := r.(type) {
	case *tg.ContactsContacts:
		p.contacts = c
		cts := c.Contacts

		sort.SliceStable(cts, func(i, j int) bool {
			return cts[i].UserID < cts[j].UserID
		})
		h := hasher.Hasher{}
		for _, contact := range cts {
			h.Update(uint32(contact.UserID))
		}

		p.contactsHash.Store(h.Sum())
	case *tg.ContactsContactsNotModified:
		if p.contacts == nil {
			return nil, xerrors.Errorf("got unexpected %T result", r)
		}
	default:
		return nil, xerrors.Errorf("unexpected type %T", r)
	}

	for _, u := range p.contacts.Users {
		user, ok := u.AsNotEmpty()
		if !ok {
			continue
		}

		if user.Phone == phone {
			return &tg.InputPeerUser{
				UserID:     user.ID,
				AccessHash: user.AccessHash,
			}, nil
		}
	}

	return nil, xerrors.Errorf("can't resolve phone %q", phone)
}
