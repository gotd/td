package transport

import (
	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/proto/codec"
)

// MTProxy creates MTProxy obfuscated transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#transport-obfuscation
func MTProxy(d Dialer, dc int, secret []byte) (*Transport, error) {
	if len(secret) != 16 {
		return nil, xerrors.Errorf("invalid secret: secret must have length of 16 bytes, got %d", len(secret))
	}

	return NewTransport(d, func() Codec {
		return &codec.MTProxyObfuscated2{
			Codec:  codec.PaddedIntermediate{},
			DC:     dc,
			Secret: secret,
		}
	}), nil
}
