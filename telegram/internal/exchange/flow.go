// Package exchange contains Telegram key exchange algorithm flows.
// See https://core.telegram.org/mtproto/auth_key.
package exchange

import (
	"io"
	"time"

	"go.uber.org/zap"

	"github.com/gotd/td/transport"
)

// Config contains common for server and client side
// dependencies.
type Config struct {
	clock func() time.Time
	rand  io.Reader
	conn  transport.Conn

	log *zap.Logger
}

// NewConfig creates new Config.
func NewConfig(clock func() time.Time, rand io.Reader, conn transport.Conn, log *zap.Logger) Config {
	return Config{
		clock: clock,
		rand:  rand,
		conn:  conn,
		log:   log,
	}
}
