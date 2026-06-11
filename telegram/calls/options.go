package calls

import (
	"crypto/rand"
	"io"

	"go.uber.org/zap"
)

// Options configures a Client.
type Options struct {
	// Logger is used for diagnostics. Defaults to a no-op logger.
	Logger *zap.Logger
	// Random is the entropy source for DH exponents. Defaults to crypto/rand.
	Random io.Reader
}

func (o *Options) setDefaults() {
	if o.Logger == nil {
		o.Logger = zap.NewNop()
	}
	if o.Random == nil {
		o.Random = rand.Reader
	}
}
