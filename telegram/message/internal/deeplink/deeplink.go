// Package deeplink contains deeplink parsing helpers.
package deeplink

import (
	"net/url"
	"strings"

	"github.com/go-faster/errors"
)

// Type is an enum type of Telegram deeplinks types.
type Type string

const (
	// Resolve is deeplink like
	//
	// 	tg:resolve?domain={domain}
	// 	tg://resolve?domain={domain}
	// 	https://t.me/{domain}
	// 	https://telegram.me/{domain}
	//
	Resolve Type = "resolve"

	// Join is deeplink like
	//
	// 	tg:join?invite={hash}
	// 	tg://join?invite={hash}
	// 	https://t.me/joinchat/{hash}
	// 	https://telegram.me/joinchat/{hash}
	// 	t.me/+{hash}
	//
	Join Type = "join"
)

// DeepLink represents Telegram deeplink.
type DeepLink struct {
	Type Type
	Args url.Values
}

func ensureParam(query url.Values, key string) error {
	if query.Get(key) == "" {
		return errors.Errorf("should have %q query parameter", key)
	}
	return nil
}

func (d DeepLink) validate() error {
	switch d.Type {
	case Resolve:
		return ensureParam(d.Args, "domain")
	case Join:
		return ensureParam(d.Args, "invite")
	default:
		return errors.Errorf("unsupported deeplink %q", d.Type)
	}
}

func parseTg(u *url.URL) (DeepLink, error) {
	query := u.Query()
	switch Type(u.Hostname()) {
	case Resolve:
		return DeepLink{
			Type: Resolve,
			Args: query,
		}, nil
	case Join:
		return DeepLink{
			Type: Join,
			Args: query,
		}, nil
	}

	return DeepLink{}, errors.Errorf("unsupported deeplink %q", u.String())
}

func parseHTTPS(u *url.URL) (DeepLink, error) {
	cleanInviteHash := func(root string) string {
		hash := strings.Trim(root, "+ ")
		if u.RawPath == "" {
			hash = url.PathEscape(hash)
		}
		return hash
	}

	query := url.Values{}
	p := strings.TrimPrefix(u.Path, "/")
	p = strings.TrimSuffix(p, "/")
	split := strings.Split(p, "/")
	var (
		root = split[0]
		base string
	)
	if len(split) > 1 {
		base = split[1]
	}

	switch root {
	case "joinchat":
		query.Set("invite", cleanInviteHash(base))
		return DeepLink{
			Type: Join,
			Args: query,
		}, nil
	case "":
		return DeepLink{}, errors.Errorf("unsupported deeplink %q", u.String())
	}

	switch root[0] {
	case ' ', '+':
		query.Set("invite", cleanInviteHash(root))
		return DeepLink{
			Type: Join,
			Args: query,
		}, nil
	default:
		if err := ValidateDomain(root); err != nil {
			return DeepLink{}, err
		}
		query.Set("domain", root)
		return DeepLink{
			Type: Resolve,
			Args: query,
		}, nil
	}
}

func hasTelegramPrefix(link string) bool {
	return strings.HasPrefix(link, "t.me") ||
		strings.HasPrefix(link, "telegram.me") ||
		strings.HasPrefix(link, "telegram.dog")
}

// IsDeeplinkLike returns true if string may be a valid deeplink.
func IsDeeplinkLike(link string) bool {
	return strings.HasPrefix(link, "tg:") ||
		hasTelegramPrefix(link) ||
		strings.HasPrefix(link, "https://")
}

// Parse parses and returns deeplink.
func Parse(link string) (DeepLink, error) {
	switch {
	// Normalize case like t.me/gotd.
	case hasTelegramPrefix(link):
		link = strings.TrimSuffix("https://"+link, "/")
	// Normalize case like tg:resolve?domain=gotd.
	case !strings.HasPrefix(link, "tg://") && strings.HasPrefix(link, "tg:"):
		link = "tg://" + strings.TrimPrefix(link, "tg:")
	}

	u, err := url.Parse(link)
	if err != nil {
		return DeepLink{}, errors.Wrapf(err, "invalid URL %q", link)
	}

	var d DeepLink
	switch {
	case u.Scheme == "https":
		switch strings.TrimPrefix(u.Hostname(), "www.") {
		case "t.me", "telegram.me", "telegram.dog":
			d, err = parseHTTPS(u)
		default:
			return DeepLink{}, errors.Errorf("invalid domain %q", link)
		}
	case u.Scheme == "tg":
		d, err = parseTg(u)
	default:
		return DeepLink{}, errors.Errorf("invalid deeplink %q", link)
	}
	if err != nil {
		return DeepLink{}, err
	}
	if err := d.validate(); err != nil {
		return DeepLink{}, err
	}

	return d, nil
}

// Expect parses deeplink and check type its type.
func Expect(link string, typ Type) (DeepLink, error) {
	l, err := Parse(link)
	if err != nil {
		return l, err
	}
	if l.Type != typ {
		return l, errors.Errorf("unexpected deeplink type %q", l.Type)
	}
	return l, nil
}
