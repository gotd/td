package calls

import (
	"crypto/rand"
	"io"

	"github.com/gotd/log"
)

// Options configures a Client.
type Options struct {
	// Logger is used for diagnostics. Defaults to a no-op logger.
	Logger log.Logger
	// Random is the entropy source for DH exponents. Defaults to crypto/rand.
	Random io.Reader
}

func (o *Options) setDefaults() {
	if o.Logger == nil {
		o.Logger = log.Nop
	}
	if o.Random == nil {
		o.Random = rand.Reader
	}
}
