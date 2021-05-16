package tgtest

import (
	"context"
	"crypto/rsa"
	"net"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

// Cluster creates cluster of multiple servers, representing telegram multiple
// datacenters.
type Cluster struct {
	servers map[int]*Server
	keys    []*rsa.PublicKey

	// DCs config state.
	cfg    tg.Config
	cfgMux sync.RWMutex

	// Signal for readiness.
	ready *tdsync.Ready

	common *Dispatcher
	log    *zap.Logger
	codec  func() transport.Codec
}

// NewCluster creates new server Cluster.
func NewCluster(codec func() transport.Codec) *Cluster {
	q := &Cluster{
		servers: map[int]*Server{},
		ready:   tdsync.NewReady(),
		common:  NewDispatcher(),
		log:     zap.NewNop(),
		codec:   codec,
	}
	q.common.Fallback(q.fallback())

	return q
}

// WithLogger sets logger.
func (c *Cluster) WithLogger(log *zap.Logger) *Cluster {
	c.log = log
	return c
}

// Config returns config for client.
func (c *Cluster) Config() tg.Config {
	c.cfgMux.RLock()
	defer c.cfgMux.RUnlock()

	return c.cfg
}

// Common returns common dispatcher.
func (c *Cluster) Common() *Dispatcher {
	return c.common
}

// DC registers new server and returns it.
func (c *Cluster) DC(id int, name string) *Server {
	logger := c.log.Named(name).With(zap.Int("dc_id", id))
	server := NewUnstartedServer(id, logger, c.codec)
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
			layerInvoke := tg.InvokeWithLayerRequest{
				Query: &tg.InitConnectionRequest{
					Query: &tg.HelpGetConfigRequest{},
				},
			}

			if err := layerInvoke.Decode(req.Buf); err != nil {
				return err
			}

			return srv.SendResult(req, &cfg)
		case tg.HelpGetConfigRequestTypeID:
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

	for id := range c.servers {
		l, err := newLocalListener(ctx)
		if err != nil {
			return xerrors.Errorf("tgtest: failed to listen on a port: %w", err)
		}

		addr, ok := l.Addr().(*net.TCPAddr)
		if !ok {
			return xerrors.Errorf("unexpected addr type %T", l.Addr())
		}

		c.cfgMux.Lock()
		c.cfg.DCOptions = append(c.cfg.DCOptions, tg.DCOption{
			Ipv6:      addr.IP.To16() != nil,
			Static:    true,
			ID:        id,
			IPAddress: addr.IP.String(),
			Port:      addr.Port,
		})
		c.cfgMux.Unlock()

		server := c.servers[id]
		g.Go(func(ctx context.Context) error {
			return server.Serve(ctx, l)
		})
	}
	c.ready.Signal()

	return g.Wait()
}
