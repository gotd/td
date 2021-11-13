package qrlogin

import (
	"context"

	"github.com/gotd/td/clock"
)

// Options of QR.
type Options struct {
	Migrate func(ctx context.Context, dcID int) error
	Clock   clock.Clock
}

func (o *Options) setDefaults() {
	// It's okay to use zero value Migrate.
	if o.Clock == nil {
		o.Clock = clock.System
	}
}
