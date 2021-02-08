package telegram

import (
	"context"

	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/tg"
)

type connBuilder struct {
	conn *conn
}

func (b connBuilder) WithSetup(f func(ctx context.Context, invoker tg.Invoker) error) connBuilder {
	b.conn.setup = f
	return b
}

func (b connBuilder) WithAddr(addr string) connBuilder {
	b.conn.addr = addr
	return b
}

func (b connBuilder) WithOptions(opts mtproto.Options) connBuilder {
	opts.Handler = b.conn
	opts.Logger = b.conn.log.Named("mtproto")
	b.conn.proto = mtproto.New(b.conn.addr, opts)
	return b
}

type noopHandler struct{}

func (n noopHandler) onSession(addr string, cfg tg.Config, s mtproto.Session) error {
	return nil
}

func (n noopHandler) onMessage(b *bin.Buffer) error {
	return nil
}

func (b connBuilder) WithNoopHandler() connBuilder {
	b.conn.handler = noopHandler{}
	return b
}

func (b connBuilder) WithHandler(handler connHandler) connBuilder {
	b.conn.handler = handler
	return b
}

func (b connBuilder) Build() *conn {
	return b.conn
}

func (c *Client) buildConn(mode connMode) connBuilder {
	opts, s := c.session.Options(c.opts)

	var id int64
	if mode != connModeUpdates {
		id = c.connsCounter.Inc()
	}

	logger := opts.Logger.Named("conn").With(
		zap.Int64("conn_id", id),
		zap.Int("dc_id", s.DC),
	)
	connection := &conn{
		addr:        s.Addr,
		mode:        mode,
		appID:       c.appID,
		device:      c.device,
		clock:       opts.Clock,
		log:         logger,
		handler:     c,
		sessionInit: tdsync.NewReady(),
		gotConfig:   tdsync.NewReady(),
		dead:        tdsync.NewReady(),
	}

	return connBuilder{conn: connection}.WithOptions(opts)
}
