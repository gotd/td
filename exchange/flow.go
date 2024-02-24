// Package exchange contains Telegram key exchange algorithm flows.
// See https://core.telegram.org/mtproto/auth_key.
package exchange

import (
	"io"
	"time"

	"go.uber.org/zap"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/transport"
)

// DefaultTimeout is default WithTimeout parameter value.
const DefaultTimeout = 1 * time.Minute

// Exchanger is builder for key exchangers.
type Exchanger struct {
	conn transport.Conn

	clock   clock.Clock
	rand    io.Reader
	log     *zap.Logger
	timeout time.Duration
	dc      int
}

// WithClock sets exchange flow clock.
func (e Exchanger) WithClock(c clock.Clock) Exchanger {
	e.clock = c
	return e
}

// WithRand sets exchange flow random source.
func (e Exchanger) WithRand(reader io.Reader) Exchanger {
	e.rand = reader
	return e
}

// WithLogger sets exchange flow logger.
func (e Exchanger) WithLogger(log *zap.Logger) Exchanger {
	e.log = log
	return e
}

// WithTimeout sets write/read deadline of every exchange request.
func (e Exchanger) WithTimeout(timeout time.Duration) Exchanger {
	e.timeout = timeout
	return e
}

// NewExchanger creates new Exchanger.
func NewExchanger(conn transport.Conn, dc int) Exchanger {
	return Exchanger{
		conn: conn,

		clock:   clock.System,
		rand:    crypto.DefaultRand(),
		log:     zap.NewNop(),
		timeout: DefaultTimeout,
		dc:      dc,
	}
}

func (e Exchanger) unencryptedWriter(input, output proto.MessageType) unencryptedWriter {
	return unencryptedWriter{
		clock:   e.clock,
		conn:    e.conn,
		timeout: e.timeout,
		input:   input,
		output:  output,
	}
}

// Client creates new ClientExchange using parameters from Exchanger.
func (e Exchanger) Client(keys []PublicKey) ClientExchange {
	return ClientExchange{
		unencryptedWriter: e.unencryptedWriter(
			proto.MessageServerResponse,
			proto.MessageFromClient,
		),
		rand: e.rand,
		log:  e.log,
		keys: keys,
		dc:   e.dc,
	}
}

// Server creates new ServerExchange using parameters from Exchanger.
func (e Exchanger) Server(key PrivateKey) ServerExchange {
	return ServerExchange{
		unencryptedWriter: e.unencryptedWriter(
			proto.MessageFromClient,
			proto.MessageServerResponse,
		),
		rand: e.rand,
		log:  e.log,
		rng:  TestServerRNG{rand: e.rand},
		key:  key,
		dc:   e.dc,
	}
}
