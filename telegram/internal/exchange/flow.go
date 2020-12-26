package exchange

import (
	"io"
	"time"

	"go.uber.org/zap"

	"github.com/gotd/td/transport"
)

type Config struct {
	clock func() time.Time
	rand  io.Reader
	conn  transport.Conn

	log *zap.Logger
}

func NewConfig(clock func() time.Time, rand io.Reader, conn transport.Conn, log *zap.Logger) Config {
	return Config{clock: clock, rand: rand, conn: conn, log: log}
}
