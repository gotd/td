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

// Quorum is a helper struct to create DC quorum.
type Quorum struct {
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

// NewQuorum creates new Quorum.
func NewQuorum(codec func() transport.Codec) *Quorum {
	q := &Quorum{
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
func (q *Quorum) WithLogger(log *zap.Logger) *Quorum {
	q.log = log
	return q
}

// Config returns config for client.
func (q *Quorum) Config() tg.Config {
	q.cfgMux.RLock()
	defer q.cfgMux.RUnlock()

	return q.cfg
}

// Common returns common dispatcher.
func (q *Quorum) Common() *Dispatcher {
	return q.common
}

// DC registers new server and returns it.
func (q *Quorum) DC(id int, name string) *Server {
	logger := q.log.Named(name).With(zap.Int("dc_id", id))
	server := NewUnstartedServer(id, logger, q.codec)
	q.servers[id] = server
	q.keys = append(q.keys, server.Key())

	// We set server fallback handler to dispatch request in order
	// 1) Explicit DC handler
	// 2) Explicit common handler
	// 3) Common fallback
	server.Dispatcher().Fallback(q.Common())
	return server
}

// Dispatch registers new server and returns its dispatcher.
func (q *Quorum) Dispatch(id int, name string) *Dispatcher {
	return q.DC(id, name).Dispatcher()
}

func (q *Quorum) fallback() HandlerFunc {
	return func(srv *Server, req *Request) error {
		cfg := q.Config()
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
func (q *Quorum) Keys() []*rsa.PublicKey {
	return q.keys
}

// Ready returns signal channel to await readiness.
func (q *Quorum) Ready() <-chan struct{} {
	return q.ready.Ready()
}

// Up runs all servers in a quorum.
func (q *Quorum) Up(ctx context.Context) error {
	grp := tdsync.NewCancellableGroup(ctx)

	for dcID, server := range q.servers {
		l, err := newLocalListener(ctx)
		if err != nil {
			return xerrors.Errorf("tgtest: failed to listen on a port: %w", err)
		}

		addr, ok := l.Addr().(*net.TCPAddr)
		if !ok {
			return xerrors.Errorf("unexpected type %T", l.Addr())
		}

		q.cfgMux.Lock()
		q.cfg.DCOptions = append(q.cfg.DCOptions, tg.DCOption{
			Ipv6:      addr.IP.To16() != nil,
			Static:    true,
			ID:        dcID,
			IPAddress: addr.IP.String(),
			Port:      addr.Port,
		})
		q.cfgMux.Unlock()

		s := server
		grp.Go(func(ctx context.Context) error {
			return s.Serve(ctx, l)
		})
	}
	q.ready.Signal()

	return grp.Wait()
}
