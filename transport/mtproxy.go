package transport

import "github.com/gotd/td/internal/proto/codec"

// MTProxy creates MTProxy obfuscated transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#transport-obfuscation
func MTProxy(d Dialer, dc int, secret []byte) *Transport {
	return NewTransport(d, func() Codec {
		return &codec.MTProxyObfuscated2{
			Codec:  codec.PaddedIntermediate{},
			DC:     dc,
			Secret: secret,
		}
	})
}
