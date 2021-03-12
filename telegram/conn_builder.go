package telegram

import (
	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mtproto"
	"github.com/gotd/td/internal/pool"
	"github.com/gotd/td/telegram/internal/manager"
	"github.com/gotd/td/tg"
)

type clientHandler struct {
	client *Client
}

func (c clientHandler) OnSession(addr string, cfg tg.Config, s mtproto.Session) error {
	return c.client.onSession(addr, cfg, s)
}

func (c clientHandler) OnMessage(b *bin.Buffer) error {
	return c.client.onMessage(b)
}

func (c *Client) asHandler() manager.Handler {
	return clientHandler{
		client: c,
	}
}

type connConstructor func(
	id int64,
	mode manager.ConnMode,
	appID int,
	addr string,
	opts mtproto.Options,
	connOpts manager.ConnOptions,
) pool.Conn

func defaultConstructor() connConstructor {
	return func(
		id int64,
		mode manager.ConnMode,
		appID int,
		addr string,
		opts mtproto.Options,
		connOpts manager.ConnOptions,
	) pool.Conn {
		return manager.CreateConn(id, mode, appID, addr, opts, connOpts)
	}
}

func (c *Client) createConn(id int64, mode manager.ConnMode, setup manager.SetupCallback) pool.Conn {
	opts, s := c.session.Options(c.opts)

	return c.create(
		id, mode, c.appID,
		s.Addr, opts, manager.ConnOptions{
			DC:      s.DC,
			Device:  c.device,
			Handler: c.asHandler(),
			Setup:   setup,
		},
	)
}
