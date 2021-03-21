package peer

import (
	"context"
	"strings"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/ascii"
	"github.com/gotd/td/telegram/message/internal"
	"github.com/gotd/td/tg"
)

// Promise is a peer promise.
type Promise func(ctx context.Context) (tg.InputPeerClass, error)

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
func Resolve(r Resolver, from string) Promise {
	from = strings.TrimSpace(from)

	if internal.IsDeeplinkLike(from) {
		return ResolveDeeplink(r, from)
	}

	return ResolveDomain(r, from)
}

// ResolveDomain uses given domain to create new message builder.
// It resolves peer of message using Sender's PeerResolver.
// Can has prefix with @ or not.
// Input examples:
//
//	@telegram
//	telegram
//
func ResolveDomain(r Resolver, domain string) Promise {
	if strings.HasPrefix(domain, "@") {
		domain = strings.TrimPrefix(domain, "@")
	}

	return func(ctx context.Context) (tg.InputPeerClass, error) {
		if err := validateDomain(domain); err != nil {
			return nil, xerrors.Errorf("validate domain: %w", err)
		}

		return r.Resolve(ctx, domain)
	}
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

// checkDomainSymbols check that domain contains only a-z, A-Z, 0-9 and '_'
// symbols.
func checkDomainSymbols(domain string) error {
	last := len(domain) - 1
	for i, r := range domain {
		if ascii.IsLatinLetter(r) {
			continue
		}

		switch {
		case i == 0:
			return xerrors.Errorf("domain should start with latin letter, got %c in %q", r, domain)
		case i == last && r == '_':
			return xerrors.Errorf("domain should end with latin letter or digit, got %c in %q", r, domain)
		case !ascii.IsDigit(r) && r != '_':
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
func ResolveDeeplink(r Resolver, deeplink string) Promise {
	return func(ctx context.Context) (tg.InputPeerClass, error) {
		link, err := internal.ExpectDeeplink(deeplink, internal.Resolve)
		if err != nil {
			return nil, err
		}
		domain := link.Args.Get("domain")

		if err := validateDomain(domain); err != nil {
			return nil, xerrors.Errorf("validate domain: %w", err)
		}

		return r.Resolve(ctx, domain)
	}
}
