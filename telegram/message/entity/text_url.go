package entity

import (
	"net/url"
	"strconv"

	"github.com/go-faster/errors"
)

func getURLFormatter(rawURL string, resolver UserResolver) (Formatter, error) {
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

		return MentionName(user), nil
	}

	return TextURL(rawURL), nil
}
