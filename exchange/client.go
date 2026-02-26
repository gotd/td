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

	// mode selects permanent vs temporary auth-key generation path.
	mode ExchangeMode
	// expiresIn is only meaningful in temporary mode and forwarded into
	// p_q_inner_data_temp_dc payload.
	expiresIn int
}

// ClientExchangeResult contains client part of key exchange result.
type ClientExchangeResult struct {
	AuthKey    crypto.AuthKey
	SessionID  int64
	ServerSalt int64
	// ExpiresAt is unix timestamp for temporary keys, zero for permanent.
	ExpiresAt int64
}
