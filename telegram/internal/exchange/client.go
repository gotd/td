package exchange

import (
	"crypto/rsa"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/proto"
)

type ClientExchange struct {
	unencryptedWriter
	keys []*rsa.PublicKey
}

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

type ClientExchangeResult struct {
	AuthKey    crypto.AuthKeyWithID
	SessionID  int64
	ServerSalt int64
}
