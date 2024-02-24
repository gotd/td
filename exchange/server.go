package exchange

import (
	"io"

	"go.uber.org/zap"

	"github.com/gotd/td/crypto"
)

// ServerExchange is a server-side key exchange flow.
type ServerExchange struct {
	unencryptedWriter
	rand io.Reader
	log  *zap.Logger

	rng ServerRNG
	key PrivateKey
	dc  int
}

// ServerExchangeResult contains server part of key exchange result.
type ServerExchangeResult struct {
	Key        crypto.AuthKey
	ServerSalt int64
}
