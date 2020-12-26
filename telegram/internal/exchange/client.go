package exchange

import (
	"crypto/rsa"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/proto"
)

// ClientExchange is a client-side key exchange flow.
type ClientExchange struct {
	unencryptedWriter
	keys []*rsa.PublicKey
}

// NewClientExchange creates new ClientExchange.
func NewClientExchange(c Config, keys ...*rsa.PublicKey) ClientExchange {
	return ClientExchange{
		unencryptedWriter: unencryptedWriter{
			Config: c,
			input:  proto.MessageServerResponse,
			output: proto.MessageFromClient,
		},
		keys: keys,
	}
}

// ClientExchangeResult contains client part of key exchange result.
type ClientExchangeResult struct {
	AuthKey    crypto.AuthKeyWithID
	SessionID  int64
	ServerSalt int64
}
