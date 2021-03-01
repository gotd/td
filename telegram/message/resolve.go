package message

import (
	"context"
	"net/url"
	"strings"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// PeerResolver is a abstraction to resolve domains and Telegram deeplinks
type PeerResolver interface {
	Resolve(ctx context.Context, domain string) (tg.InputPeerClass, error)
}

// PeerResolverFunc is a functional adapter for PeerResolver.
type PeerResolverFunc func(ctx context.Context, domain string) (tg.InputPeerClass, error)

// Resolve implements PeerResolver.
func (p PeerResolverFunc) Resolve(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	return p(ctx, domain)
}

// DefaultPeerResolver creates and returns default resolver.
func DefaultPeerResolver(raw *tg.Client) PeerResolver {
	return &plainResolver{raw: raw}
}

type plainResolver struct {
	raw *tg.Client
}

func (p plainResolver) Resolve(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	peer, err := p.raw.ContactsResolveUsername(ctx, domain)
	if err != nil {
		return nil, xerrors.Errorf("resolve: %w", err)
	}

	return findPeer(entities{
		Users:    peer.MapUsers().UserToMap(),
		Chats:    peer.MapChats().ChatToMap(),
		Channels: peer.MapChats().ChannelToMap(),
	}, peer.Peer)
}

// Resolve uses given text to create new message builder.
// It resolves peer of message using Sender's PeerResolver.
// Input examples:
//
//	@telegram
//	telegram
//	t.me/telegram
//	https://t.me/telegram
//	tg:resolve?domain=telegram
//	tg://resolve?domain=telegram
//
func (s *Sender) Resolve(from string) *Builder {
	from = strings.TrimSpace(from)

	if strings.HasPrefix(from, "tg:") ||
		strings.HasPrefix(from, "t.me") ||
		strings.HasPrefix(from, "https://") {
		return s.ResolveDeeplink(from)
	}

	return s.ResolveDomain(from)
}

// ResolveDomain uses given domain to create new message builder.
// It resolves peer of message using Sender's PeerResolver.
// Can has prefix with @ or not.
// Input examples:
//
//	@telegram
//	telegram
//
func (s *Sender) ResolveDomain(domain string) *Builder {
	if strings.HasPrefix(domain, "@") {
		domain = strings.TrimPrefix(domain, "@")
	}

	return s.builder(func(ctx context.Context) (tg.InputPeerClass, error) {
		if err := validateDomain(domain); err != nil {
			return nil, xerrors.Errorf("validate domain: %w", err)
		}

		return s.resolver.Resolve(ctx, domain)
	})
}

func validateDomain(domain string) error {
	const minDomainLength = 5
	if len(domain) < minDomainLength {
		return xerrors.Errorf("domain %q is too small", domain)
	}

	if err := checkDomainSymbols(domain); err != nil {
		return err
	}

	return nil
}

func isLatinLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// checkDomainSymbols check that domain contains only a-z, A-Z, 0-9 and '_'
// symbols.
func checkDomainSymbols(domain string) error {
	last := len(domain) - 1
	for i, r := range domain {
		if isLatinLetter(r) {
			continue
		}

		switch {
		case i == 0:
			return xerrors.Errorf("domain should start with latin letter, got %c in %q", r, domain)
		case i == last && r == '_':
			return xerrors.Errorf("domain should end with latin letter or digit, got %c in %q", r, domain)
		case !isDigit(r) && r != '_':
			return xerrors.Errorf("unexpected rune %[1]c (%[1]U) in %[2]q domain", r, domain)
		}
	}

	return nil
}

// ResolveDeeplink uses given deeplink to create new message builder.
// Deeplink is a URL like https://t.me/telegram.
// It resolves peer of message using Sender's PeerResolver.
// Input examples:
//
//	t.me/telegram
//	https://t.me/telegram
//	tg:resolve?domain=telegram
//	tg://resolve?domain=telegram
//
func (s *Sender) ResolveDeeplink(deeplink string) *Builder {
	return s.builder(func(ctx context.Context) (tg.InputPeerClass, error) {
		domain, err := parseDeeplink(deeplink)
		if err != nil {
			return nil, err
		}

		if err := validateDomain(domain); err != nil {
			return nil, xerrors.Errorf("validate domain: %w", err)
		}

		return s.resolver.Resolve(ctx, domain)
	})
}

func parseDeeplink(deeplink string) (string, error) {
	switch {
	// Normalize case like t.me/gotd.
	case strings.HasPrefix(deeplink, "t.me"):
		deeplink = "https://" + deeplink
	// Normalize case like tg:resolve?domain=gotd.
	case !strings.HasPrefix(deeplink, "tg://") && strings.HasPrefix(deeplink, "tg:"):
		deeplink = "tg://" + strings.TrimPrefix(deeplink, "tg:")
	}

	u, err := url.Parse(deeplink)
	if err != nil {
		return "", xerrors.Errorf("invalid URL %q: %w", deeplink, err)
	}

	var domain string
	switch {
	case u.Scheme == "https" && u.Hostname() == "t.me":
		domain = strings.TrimSuffix(u.Path, "/")
		domain = strings.TrimPrefix(domain, "/")

	case u.Scheme == "tg" && u.Hostname() == "resolve":
		domain = u.Query().Get("domain")
		if domain == "" {
			return "", xerrors.Errorf("deeplink %q should have domain query parameter", deeplink)
		}
	default:
		return "", xerrors.Errorf("invalid deeplink %q", deeplink)
	}

	return strings.TrimSpace(domain), nil
}
