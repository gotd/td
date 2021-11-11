package qrlogin

import "context"

// Options of QR.
type Options struct {
	Migrate func(ctx context.Context, dcID int) error
}

func (o *Options) setDefaults() {
	// It's okay to use zero value Migrate.
}
