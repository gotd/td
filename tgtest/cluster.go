package tgtest

import (
	"context"
	"crypto/rsa"
	"io"
	"net"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

type setup struct {
	srv      *Server
	dispatch *Dispatcher
}

// Cluster is a cluster of multiple servers, representing multiple Telegram datacenters.
type Cluster struct {
	setups map[int]setup
	keys   []*rsa.PublicKey

	// DCs config state.
	cfg    tg.Config
	cdnCfg tg.CDNConfig

	// Signal for readiness.
	ready *tdsync.Ready

	// RPC dispatcher.
	common *Dispatcher

	// Listen is used to create server transport.Listener.
	listen ListenFunc

	log    *zap.Logger
	random io.Reader
	codec  func() transport.Codec // nilable
}

// NewCluster creates new server Cluster.
func NewCluster(opts ClusterOptions) *Cluster {
	opts.setDefaults()

	q := &Cluster{
		setups: map[int]setup{},
		cfg:    opts.Config,
		cdnCfg: opts.CDNConfig,
		ready:  tdsync.NewReady(),
		common: NewDispatcher(),
		listen: opts.Listen,
		log:    opts.Logger,
		random: opts.Random,
		codec:  opts.Codec,
	}
	q.common.Fallback(q.fallback())

	return q
}

// Config returns config.
func (c *Cluster) Config() tg.Config {
	return c.cfg
}

// CDNConfig returns CDN config.
func (c *Cluster) CDNConfig() tg.CDNConfig {
	return c.cdnCfg
}

// Common returns common dispatcher.
func (c *Cluster) Common() *Dispatcher {
	return c.common
}

// DC registers new server and returns it.
func (c *Cluster) DC(id int, name string) (*Server, *Dispatcher) {
	if s, ok := c.setups[id]; ok {
		return s.srv, s.dispatch
	}

	key, err := rsa.GenerateKey(c.random, crypto.RSAKeyBits)
	if err != nil {
		// TODO(tdakkota): Return error instead.
		panic(err)
	}

	d := NewDispatcher()
	logger := c.log.Named(name).With(zap.Int("dc_id", id))
	server := NewServer(key, UnpackInvoke(d), ServerOptions{
		DC:     id,
		Logger: logger,
		Codec:  c.codec,
	})
	c.setups[id] = setup{
		srv:      server,
		dispatch: d,
	}
	c.keys = append(c.keys, server.Key())

	// We set server fallback handler to dispatch request in order
	// 1) Explicit DC handler
	// 2) Explicit common handler
	// 3) Common fallback
	d.Fallback(c.Common())
	return server, d
}

// Dispatch registers new server and returns its dispatcher.
func (c *Cluster) Dispatch(id int, name string) *Dispatcher {
	_, d := c.DC(id, name)
	return d
}

func (c *Cluster) fallback() HandlerFunc {
	return func(srv *Server, req *Request) error {
		id, err := req.Buf.PeekID()
		if err != nil {
			return err
		}

		var (
			decode bin.Decoder
			result bin.Encoder
		)
		switch id {
		case tg.HelpGetCDNConfigRequestTypeID:
			cfg := c.CDNConfig()

			decode = &tg.HelpGetCDNConfigRequest{}
			result = &cfg
		case tg.HelpGetConfigRequestTypeID:
			cfg := c.Config()
			cfg.ThisDC = req.DC

			decode = &tg.HelpGetConfigRequest{}
			result = &cfg
		default:
			return xerrors.Errorf("unexpected TypeID %x call", id)
		}

		if err := decode.Decode(req.Buf); err != nil {
			return err
		}
		return srv.SendResult(req, result)
	}
}

// Keys returns all servers public keys.
func (c *Cluster) Keys() []*rsa.PublicKey {
	return c.keys
}

// Ready returns signal channel to await readiness.
func (c *Cluster) Ready() <-chan struct{} {
	return c.ready.Ready()
}

// Up runs all servers in a cluster.
func (c *Cluster) Up(ctx context.Context) error {
	g := tdsync.NewCancellableGroup(ctx)

	for dcID := range c.setups {
		l, err := c.listen(ctx, dcID)
		if err != nil {
			return xerrors.Errorf("tgtest, DC %d: listen port: %w", dcID, err)
		}

		// Add TCP listeners to config.
		if addr, ok := l.Addr().(*net.TCPAddr); ok {
			c.cfg.DCOptions = append(c.cfg.DCOptions, tg.DCOption{
				Ipv6:      addr.IP.To16() != nil,
				Static:    true,
				ID:        dcID,
				IPAddress: addr.IP.String(),
				Port:      addr.Port,
			})
		}

		s := c.setups[dcID]
		g.Go(func(ctx context.Context) error {
			return s.srv.Serve(ctx, l)
		})
	}
	c.ready.Signal()

	return g.Wait()
}
