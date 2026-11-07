package qrlogin

import (
	"context"

	"github.com/gotd/td/clock"
)

// Options of QR.
type Options struct {
	Migrate func(ctx context.Context, dcID int) error
	Clock   clock.Clock
	// Password returns 2FA password of the account (optional).
	//
	// If set, Import and Auth complete 2FA (auth.checkPassword) when the
	// account has a password, instead of returning SESSION_PASSWORD_NEEDED.
	Password func(ctx context.Context) (string, error)
}

func (o *Options) setDefaults() {
	// It's okay to use zero value Migrate.
	if o.Clock == nil {
		o.Clock = clock.System
	}
}
