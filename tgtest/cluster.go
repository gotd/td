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

// Cluster is a cluster of multiple servers, representing multiple Telegram datacenters.
type Cluster struct {
	servers map[int]*Server
	keys    []*rsa.PublicKey

	// DCs config state.
	cfg tg.Config

	// Signal for readiness.
	ready *tdsync.Ready

	// RPC dispatcher.
	common *Dispatcher

	// Listen is used to create server listener.
	listen func(ctx context.Context, dc int) (net.Listener, error)

	log    *zap.Logger
	random io.Reader
	codec  func() transport.Codec // nilable
}

// NewCluster creates new server Cluster.
func NewCluster(opts ClusterOptions) *Cluster {
	opts.setDefaults()

	q := &Cluster{
		servers: map[int]*Server{},
		cfg:     opts.Config,
		ready:   tdsync.NewReady(),
		common:  NewDispatcher(),
		listen:  opts.Listen,
		log:     opts.Logger,
		random:  opts.Random,
		codec:   opts.Codec,
	}
	q.common.Fallback(q.fallback())

	return q
}

// Config returns config for client.
func (c *Cluster) Config() tg.Config {
	return c.cfg
}

// Common returns common dispatcher.
func (c *Cluster) Common() *Dispatcher {
	return c.common
}

// DC registers new server and returns it.
func (c *Cluster) DC(id int, name string) *Server {
	key, err := rsa.GenerateKey(c.random, crypto.RSAKeyBits)
	if err != nil {
		// TODO(tdakkota): Return error instead.
		panic(err)
	}

	logger := c.log.Named(name).With(zap.Int("dc_id", id))
	server := NewServer(key, ServerOptions{
		DC:     id,
		Logger: logger,
		Codec:  c.codec,
	})
	c.servers[id] = server
	c.keys = append(c.keys, server.Key())

	// We set server fallback handler to dispatch request in order
	// 1) Explicit DC handler
	// 2) Explicit common handler
	// 3) Common fallback
	server.Dispatcher().Fallback(c.Common())
	return server
}

// Dispatch registers new server and returns its dispatcher.
func (c *Cluster) Dispatch(id int, name string) *Dispatcher {
	return c.DC(id, name).Dispatcher()
}

type typeIDObject struct {
	TypeID uint32
}

func (t *typeIDObject) Decode(b *bin.Buffer) error {
	id, err := b.PeekID()
	if err != nil {
		return xerrors.Errorf("peek id: %w", err)
	}
	t.TypeID = id
	return nil
}

func (t *typeIDObject) Encode(*bin.Buffer) error {
	return xerrors.New("typeIDObject must not be encoded")
}

func (c *Cluster) fallback() HandlerFunc {
	return func(srv *Server, req *Request) error {
		cfg := c.Config()
		cfg.ThisDC = req.DC

		id, err := req.Buf.PeekID()
		if err != nil {
			return err
		}

		switch id {
		case tg.InvokeWithLayerRequestTypeID:
			obj := typeIDObject{}
			r := &tg.InvokeWithLayerRequest{
				Query: &obj,
			}
			if err := r.Decode(req.Buf); err != nil {
				return err
			}
			req.Session.Layer.Store(int32(r.Layer))

			return c.common.OnMessage(srv, req)
		case tg.InitConnectionRequestTypeID:
			obj := typeIDObject{}
			r := &tg.InitConnectionRequest{
				Query: &obj,
			}
			if err := r.Decode(req.Buf); err != nil {
				return err
			}
			c.log.Debug("Init connection call", zap.Inline(req.Session))

			return c.common.OnMessage(srv, req)
		case tg.HelpGetConfigRequestTypeID:
			var r tg.HelpGetConfigRequest
			if err := r.Decode(req.Buf); err != nil {
				return err
			}

			return srv.SendResult(req, &cfg)
		default:
			return xerrors.Errorf("unexpected TypeID %x call", id)
		}
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

	for dcID := range c.servers {
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

		server := c.servers[dcID]
		g.Go(func(ctx context.Context) error {
			return server.Serve(ctx, l)
		})
	}
	c.ready.Signal()

	return g.Wait()
}
