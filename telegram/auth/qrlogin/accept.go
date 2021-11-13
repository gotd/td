package qrlogin

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// AcceptQR accepts given token.
//
// See https://core.telegram.org/api/qr-login#accepting-a-login-token.
func AcceptQR(ctx context.Context, raw *tg.Client, t Token) (*tg.Authorization, error) {
	auth, err := raw.AuthAcceptLoginToken(ctx, t.token)
	if err != nil {
		return nil, errors.Wrap(err, "accept")
	}
	return auth, nil
}
