// Package deeplink contains deeplink parsing helpers.
package deeplink

import (
	"net/url"
	"path"
	"strings"

	"golang.org/x/xerrors"
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
		return xerrors.Errorf("should have %q query parameter", key)
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
		return xerrors.Errorf("unsupported deeplink %q", d.Type)
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

	return DeepLink{}, xerrors.Errorf("unsupported deeplink %q", u.String())
}

func parseHTTPS(u *url.URL) (DeepLink, error) {
	query := u.Query()
	root, base := path.Split(path.Clean(u.Path))
	root = strings.TrimPrefix(root, "/")
	root = strings.TrimSuffix(root, "/")

	switch root {
	case "":
		if len(base) > 0 && base[0] == '+' {
			query.Set("invite", base[1:])
			return DeepLink{
				Type: Join,
				Args: query,
			}, nil
		}
		query.Set("domain", base)
		return DeepLink{
			Type: Resolve,
			Args: query,
		}, nil
	case "joinchat":
		query.Set("invite", base)
		return DeepLink{
			Type: Join,
			Args: query,
		}, nil
	}

	return DeepLink{}, xerrors.Errorf("unsupported deeplink %q", u.String())
}

// IsDeeplinkLike returns true if string may be a valid deeplink.
func IsDeeplinkLike(link string) bool {
	return strings.HasPrefix(link, "tg:") ||
		strings.HasPrefix(link, "t.me") ||
		strings.HasPrefix(link, "https://")
}

// Parse parses and returns deeplink.
func Parse(link string) (DeepLink, error) {
	switch {
	// Normalize case like t.me/gotd.
	case strings.HasPrefix(link, "t.me"):
		link = strings.TrimSuffix("https://"+link, "/")
	// Normalize case like tg:resolve?domain=gotd.
	case !strings.HasPrefix(link, "tg://") && strings.HasPrefix(link, "tg:"):
		link = "tg://" + strings.TrimPrefix(link, "tg:")
	}

	u, err := url.Parse(link)
	if err != nil {
		return DeepLink{}, xerrors.Errorf("invalid URL %q: %w", link, err)
	}

	var d DeepLink
	switch {
	case u.Scheme == "https" && u.Hostname() == "t.me":
		d, err = parseHTTPS(u)
	case u.Scheme == "tg":
		d, err = parseTg(u)
	default:
		return DeepLink{}, xerrors.Errorf("invalid deeplink %q", link)
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
		return l, xerrors.Errorf("unexpected deeplink type %q", l.Type)
	}
	return l, nil
}
