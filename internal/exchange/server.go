package exchange

import (
	"crypto/rsa"
	"io"

	"go.uber.org/zap"

	"github.com/gotd/td/internal/crypto"
)

// ServerExchange is a server-side key exchange flow.
type ServerExchange struct {
	unencryptedWriter
	rand io.Reader
	log  *zap.Logger

	rng ServerRNG
	key *rsa.PrivateKey
}

// ServerExchangeResult contains server part of key exchange result.
type ServerExchangeResult struct {
	Key        crypto.AuthKeyWithID
	ServerSalt int64
}
