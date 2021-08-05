package tgtest

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

type setup struct {
	srv      *Server
	dispatch *Dispatcher
}

// Cluster is a cluster of multiple servers, representing multiple Telegram datacenters.
type Cluster struct {
	// denotes to use websocket listener
	web bool

	setups map[int]setup
	keys   []*rsa.PublicKey

	// DCs config state.
	cfg     tg.Config
	cdnCfg  tg.CDNConfig
	domains map[int]string

	// Signal for readiness.
	ready *tdsync.Ready

	// RPC dispatcher.
	common *Dispatcher

	log    *zap.Logger
	random io.Reader
	codec  func() transport.Codec // nilable
}

// NewCluster creates new server Cluster.
func NewCluster(opts ClusterOptions) *Cluster {
	opts.setDefaults()

	q := &Cluster{
		web:     opts.Web,
		setups:  map[int]setup{},
		keys:    nil,
		cfg:     opts.Config,
		cdnCfg:  opts.CDNConfig,
		domains: map[int]string{},
		ready:   tdsync.NewReady(),
		common:  NewDispatcher(),
		log:     opts.Logger,
		random:  opts.Random,
		codec:   opts.Codec,
	}
	q.common.Fallback(q.fallback())

	return q
}

// List returns DCs list.
func (c *Cluster) List() dcs.List {
	return dcs.List{
		Options: c.cfg.DCOptions,
		Domains: c.domains,
	}
}

// Resolver returns dcs.Resolver to use.
func (c *Cluster) Resolver() dcs.Resolver {
	if c.web {
		return dcs.Websocket(dcs.WebsocketOptions{})
	}

	return dcs.Plain(dcs.PlainOptions{})
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
			cfg := c.cdnCfg

			decode = &tg.HelpGetCDNConfigRequest{}
			result = &cfg
		case tg.HelpGetConfigRequestTypeID:
			cfg := c.cfg
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

func newLocalListener(ctx context.Context) (net.Listener, error) {
	cfg := net.ListenConfig{}
	l, err := cfg.Listen(ctx, "tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, xerrors.Errorf("listen: %w", err)
	}
	return l, nil
}

// Up runs all servers in a cluster.
func (c *Cluster) Up(ctx context.Context) error {
	g := tdsync.NewCancellableGroup(ctx)

	listen := func(ctx context.Context, _ int) (net.Listener, error) {
		return newLocalListener(ctx)
	}
	if c.web {
		// Create local random listener
		l, err := newLocalListener(ctx)
		if err != nil {
			return err
		}

		mux := http.NewServeMux()
		srv := http.Server{
			Handler: mux,
			BaseContext: func(net.Listener) context.Context {
				return ctx
			},
		}
		g.Go(func(ctx context.Context) error {
			if err := srv.Serve(l); err != nil && !xerrors.Is(err, http.ErrServerClosed) {
				return xerrors.Errorf("serve: %w", err)
			}
			return nil
		})
		g.Go(func(ctx context.Context) error {
			<-ctx.Done()

			return srv.Close()
		})

		baseURL := url.URL{
			Scheme: "http",
			Host:   l.Addr().String(),
		}
		listen = func(ctx context.Context, dc int) (net.Listener, error) {
			listener, handler := transport.WebsocketListener(baseURL.Host)

			path := fmt.Sprintf("/dc/%d", dc)
			mux.Handle(path, handler)

			dcURL := baseURL
			dcURL.Path = path
			c.domains[dc] = dcURL.String()
			return listener, nil
		}
	}

	for dcID, s := range c.setups {
		l, err := listen(ctx, dcID)
		if err != nil {
			return xerrors.Errorf("DC %d: listen port: %w", dcID, err)
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

		// Copy iteration value.
		srv := s.srv
		g.Go(func(ctx context.Context) error {
			return srv.Serve(ctx, transport.ListenCodec(nil, l))
		})
	}
	c.ready.Signal()

	return g.Wait()
}
