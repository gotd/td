package html

import (
	"net"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/go-faster/errors"

	"github.com/gotd/td/internal/ascii"
	"github.com/gotd/td/telegram/message/entity"
)

func isIPv6(str string) bool {
	ip := net.ParseIP(str)
	return strings.Contains(str, ":") && ip != nil
}

func validateHostname(u *url.URL) error {
	// TODO(tdakkota): make sure that it is correct
	ipv6 := isIPv6(u.Host)
	if !strings.ContainsRune(u.Host, '.') && ipv6 {
		return errors.New("wrong HTTP URL")
	}
	if ipv6 {
		return nil
	}

	allowedSymbol := func(c rune) bool {
		return ascii.IsLatinLetter(c) ||
			ascii.IsDigit(c) ||
			(c >= '&' && c <= '.') ||
			c == '_' ||
			c == '!' ||
			c == '$' ||
			c == '~' ||
			c == ';' ||
			c == '=' ||
			c > utf8.RuneSelf
	}

	for _, c := range u.Host {
		if !allowedSymbol(c) {
			return errors.Errorf("disallowed character %c in URL host", c)
		}
	}

	return nil
}

func getURLFormatter(rawURL string, resolver entity.UserResolver) (entity.Formatter, error) {
	const defaultProtocol = "http"
	if rawURL == "" {
		return nil, errors.New("empty URL")
	}

	// FIXME(tdakkota): move normalization to deeplink package when it's done?
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "tg" && u.Host == "user" {
		id, err := strconv.ParseInt(u.Query().Get("id"), 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid user ID %q", id)
		}

		user, err := resolver(id)
		if err != nil {
			return nil, errors.Wrapf(err, "can't resolve user %q", id)
		}

		return entity.MentionName(user), nil
	}
	if u.Scheme == "" {
		u.Scheme = defaultProtocol
		u.Host = u.Path
		u.Path = "/"
		rawURL = u.String()
	}

	if err := validateHostname(u); err != nil {
		return nil, err
	}

	return entity.TextURL(rawURL), nil
}
