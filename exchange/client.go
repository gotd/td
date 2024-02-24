package exchange

import (
	"io"

	"go.uber.org/zap"

	"github.com/gotd/td/crypto"
)

// ClientExchange is a client-side key exchange flow.
type ClientExchange struct {
	unencryptedWriter
	rand io.Reader
	log  *zap.Logger

	keys []PublicKey
	dc   int
}

// ClientExchangeResult contains client part of key exchange result.
type ClientExchangeResult struct {
	AuthKey    crypto.AuthKey
	SessionID  int64
	ServerSalt int64
}
