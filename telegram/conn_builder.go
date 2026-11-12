package telegram

import (
	"context"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/internal/manager"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
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

// primaryHandler is the handler of the primary connection: it also recovers
// updates when the server starts a new session.
type primaryHandler struct {
	clientHandler
}

// OnNewSession implements manager.NewSessionHandler.
func (c primaryHandler) OnNewSession(int64) error {
	c.client.onPrimaryNewSession()
	return nil
}

type cdnClientHandler struct {
	client *Client
}

func (c cdnClientHandler) OnSession(cfg tg.Config, s mtproto.Session) error {
	// CDN sessions are stored separately from regular DC sessions.
	return c.client.onCDNSession(cfg, s)
}

func (cdnClientHandler) OnMessage(*bin.Buffer) error {
	// CDN connections never deliver updates.
	return nil
}

func (c *Client) asCDNHandler() manager.Handler {
	return cdnClientHandler{
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

// createPrimaryConn creates primary connection and records its DC in
// c.connDC, so callers must hold connMux (or own c exclusively).
func (c *Client) createPrimaryConn(setup manager.SetupCallback) pool.Conn {
	conn, dc := c.createConnWith(0, c.defaultMode, nil, setup, func(dc int, err error) {
		c.notifyConnDead(ConnDeadEvent{DC: dc, Kind: ConnKindPrimary, Err: err})
	})
	c.connDC = dc
	return conn
}

// createConnWith creates connection that uses primary session and returns
// it with the DC of the session it was created for.
//
// If dialer is nil, primary DC dialer is used.
func (c *Client) createConnWith(
	id int64,
	mode manager.ConnMode,
	dialer mtproto.Dialer,
	setup manager.SetupCallback,
	onDead func(dc int, err error),
) (pool.Conn, int) {
	// Load the session once, so that options, dialer and reported DC match.
	opts, s := c.session.Options(c.opts)
	opts.Logger = c.log.Named("conn").With(
		log.Int64("conn_id", id),
		log.Int("dc_id", s.DC),
	).Logger()
	if dialer == nil {
		dialer = c.primaryDC(s.DC)
	}

	dc := s.DC
	opts.SessionState = c.acquireSessionState(dc)
	handler := c.asHandler()
	if id == 0 {
		handler = primaryHandler{clientHandler: clientHandler{client: c}}
	}
	var onConnDead func(error)
	if onDead != nil {
		onConnDead = func(err error) { onDead(dc, err) }
	}
	return c.create(
		dialer, mode, c.appID,
		opts, manager.ConnOptions{
			DC:      dc,
			Test:    c.testDC,
			Layer:   c.layer,
			Device:  c.device,
			Handler: handler,
			Setup:   setup,
			OnDead:  onConnDead,
		},
	), dc
}
