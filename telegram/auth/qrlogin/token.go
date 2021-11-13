package qrlogin

import (
	"encoding/base64"
	"image"
	"net/url"
	"time"

	"github.com/go-faster/errors"
	"rsc.io/qr"
)

// Token represents Telegram QR Login token.
type Token struct {
	token   []byte
	expires time.Time
}

// ParseTokenURL creates Token from given URL.
func ParseTokenURL(u string) (Token, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return Token{}, err
	}
	switch {
	case parsed.Scheme != "tg":
		return Token{}, errors.Errorf("unexpected scheme %q", parsed.Scheme)
	case parsed.Host != "login":
		return Token{}, errors.Errorf("wrong path %q", parsed.Host)
	}

	q := parsed.Query()
	if q.Get("token") == "" {
		return Token{}, errors.New("token is empty")
	}
	token, err := base64.URLEncoding.DecodeString(q.Get("token"))
	if err != nil {
		return Token{}, err
	}

	return NewToken(token, 0), nil
}

// NewToken creates new Token.
func NewToken(token []byte, expires int) Token {
	return Token{
		token:   token,
		expires: time.Unix(int64(expires), 0),
	}
}

// Expires returns token expiration time.
func (t Token) Expires() time.Time {
	return t.expires
}

// String implements fmt.Stringer.
func (t Token) String() string {
	return base64.URLEncoding.EncodeToString(t.token)
}

// URL returns login URL.
//
// See https://core.telegram.org/api/qr-login#exporting-a-login-token.
func (t Token) URL() string {
	return "tg://login?token=" + base64.URLEncoding.EncodeToString(t.token)
}

// Image returns QR image.
func (t Token) Image(level qr.Level) (image.Image, error) {
	code, err := qr.Encode(t.URL(), level)
	if err != nil {
		return nil, errors.Wrap(err, "encode")
	}
	return code.Image(), nil
}
