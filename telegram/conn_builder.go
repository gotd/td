package telegram

import (
	"context"

	"go.uber.org/zap"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/mtproto"
	"github.com/nnqq/td/internal/pool"
	"github.com/nnqq/td/telegram/dcs"
	"github.com/nnqq/td/telegram/internal/manager"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/transport"
)

type clientHandler struct {
	client *Client
}

func (c clientHandler) OnSession(cfg tg.Config, s mtproto.Session) error {
	return c.client.onSession(cfg, s)
}

func (c clientHandler) OnMessage(b *bin.Buffer) error {
	return c.client.handleUpdates(b)
}

func (c *Client) asHandler() manager.Handler {
	return clientHandler{
		client: c,
	}
}

type connConstructor func(
	create mtproto.Dialer,
	mode manager.ConnMode,
	appID int,
	opts mtproto.Options,
	connOpts manager.ConnOptions,
) pool.Conn

func defaultConstructor() connConstructor {
	return func(
		create mtproto.Dialer,
		mode manager.ConnMode,
		appID int,
		opts mtproto.Options,
		connOpts manager.ConnOptions,
	) pool.Conn {
		return manager.CreateConn(create, mode, appID, opts, connOpts)
	}
}

func (c *Client) dcList() dcs.List {
	cfg := c.cfg.Load()
	return dcs.List{
		Options: cfg.DCOptions,
		Domains: c.domains,
		Test:    c.testDC,
	}
}

func (c *Client) primaryDC(dc int) mtproto.Dialer {
	return func(ctx context.Context) (transport.Conn, error) {
		return c.resolver.Primary(ctx, dc, c.dcList())
	}
}

func (c *Client) createPrimaryConn(setup manager.SetupCallback) pool.Conn {
	return c.createConn(0, c.defaultMode, setup)
}

func (c *Client) createConn(
	id int64,
	mode manager.ConnMode,
	setup manager.SetupCallback,
) pool.Conn {
	opts, s := c.session.Options(c.opts)
	opts.Logger = c.log.Named("conn").With(
		zap.Int64("conn_id", id),
		zap.Int("dc_id", s.DC),
	)

	return c.create(
		c.primaryDC(s.DC), mode, c.appID,
		opts, manager.ConnOptions{
			DC:      s.DC,
			Test:    c.testDC,
			Device:  c.device,
			Handler: c.asHandler(),
			Setup:   setup,
		},
	)
}
