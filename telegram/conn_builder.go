package telegram

import (
	"context"

	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/pool"
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

func (b connBuilder) WithOptions(dc int, addr string, opts mtproto.Options) connBuilder {
	b.conn.log = opts.Logger.Named("conn").With(
		zap.Int64("conn_id", b.conn.id),
		zap.Int("dc_id", dc),
	)

	b.conn.addr = addr
	opts.Handler = b.conn
	opts.Logger = b.conn.log.Named("mtproto").With(zap.String("addr", b.conn.addr))
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

func (c *Client) buildConn(mode connMode, session *pool.SyncSession) connBuilder {
	opts, s := session.Options(c.opts)

	var id int64
	if mode != connModeUpdates {
		id = c.connsCounter.Inc()
	}

	connection := &conn{
		id:          id,
		addr:        s.Addr,
		mode:        mode,
		appID:       c.appID,
		device:      c.device,
		clock:       opts.Clock,
		handler:     c,
		sessionInit: tdsync.NewReady(),
		gotConfig:   tdsync.NewReady(),
		dead:        tdsync.NewReady(),
	}

	return connBuilder{conn: connection}.WithOptions(s.DC, s.Addr, opts)
}
