package peer

import (
	"context"
	"strings"

	"github.com/go-faster/errors"

	"github.com/gotd/td/internal/ascii"
	"github.com/gotd/td/telegram/internal/deeplink"
	"github.com/gotd/td/tg"
)

// Promise is a peer promise.
type Promise func(ctx context.Context) (tg.InputPeerClass, error)

// Resolve uses given string to create new peer promise.
// It resolves peer of message using given Resolver.
// Input examples:
//
//	@telegram
//	telegram
//	t.me/telegram
//	https://t.me/telegram
//	tg:resolve?domain=telegram
//	tg://resolve?domain=telegram
//	+13115552368
//	+1 (311) 555-0123
//	+1 311 555-6162
//	13115556162
//
func Resolve(r Resolver, from string) Promise {
	from = strings.TrimSpace(from)

	if deeplink.IsDeeplinkLike(from) {
		return ResolveDeeplink(r, from)
	}
	if isPhoneNumber(from) {
		return ResolvePhone(r, from)
	}

	return ResolveDomain(r, from)
}

func isPhoneNumber(s string) bool {
	if s == "" {
		return false
	}
	r := rune(s[0])
	return r == '+' || ascii.IsDigit(r)
}

func cleanupPhone(phone string) string {
	clean := strings.Builder{}
	clean.Grow(len(phone) + 1)

	for _, ch := range phone {
		if ascii.IsDigit(ch) {
			clean.WriteRune(ch)
		}
	}

	return clean.String()
}

// ResolvePhone uses given phone to create new peer promise.
// It resolves peer of message using given Resolver.
// Input example:
//
//	+13115552368
//	+1 (311) 555-0123
//	+1 311 555-6162
//	13115556162
//
// Note that Telegram represents phone numbers according to the E.164 standard
// without the plus sign (”+”) prefix. The resolver therefore takes an easy
// route and just deletes any non-digit symbols from phone number string.
func ResolvePhone(r Resolver, phone string) Promise {
	return func(ctx context.Context) (tg.InputPeerClass, error) {
		return r.ResolvePhone(ctx, cleanupPhone(phone))
	}
}

// ResolveDomain uses given domain to create new peer promise.
// It resolves peer of message using given Resolver.
// Can has prefix with @ or not.
// Input examples:
//
//	@telegram
//	telegram
//
func ResolveDomain(r Resolver, domain string) Promise {
	return func(ctx context.Context) (tg.InputPeerClass, error) {
		domain = strings.TrimPrefix(domain, "@")

		if err := validateDomain(domain); err != nil {
			return nil, errors.Wrap(err, "validate domain")
		}

		return r.ResolveDomain(ctx, domain)
	}
}

func validateDomain(domain string) error {
	return deeplink.ValidateDomain(domain)
}

// ResolveDeeplink uses given deeplink to create new peer promise.
// Deeplink is a URL like https://t.me/telegram.
// It resolves peer of message using given Resolver.
// Input examples:
//
//	t.me/telegram
//	https://t.me/telegram
//	tg:resolve?domain=telegram
//	tg://resolve?domain=telegram
//
func ResolveDeeplink(r Resolver, u string) Promise {
	return func(ctx context.Context) (tg.InputPeerClass, error) {
		link, err := deeplink.Expect(u, deeplink.Resolve)
		if err != nil {
			return nil, err
		}
		domain := link.Args.Get("domain")

		if err := validateDomain(domain); err != nil {
			return nil, errors.Wrap(err, "validate domain")
		}

		return r.ResolveDomain(ctx, domain)
	}
}
