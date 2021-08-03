package tgtest

import (
	"context"
	"crypto/rsa"
	"io"
	"net"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mtproto"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/transport"
)

// Server is a MTProto server structure.
type Server struct {
	dcID   int
	codec  func() transport.Codec // immutable,nilable
	key    *rsa.PrivateKey        // immutable
	cipher crypto.Cipher          // immutable

	dispatcher *Dispatcher
	ctx        context.Context

	clock clock.Clock             // immutable
	log   *zap.Logger             // immutable
	msgID mtproto.MessageIDSource // immutable

	users *users
	types *tmap.Map
}

// NewServer creates new Server.
func NewServer(key *rsa.PrivateKey, opts ServerOptions) *Server {
	opts.setDefaults()

	s := &Server{
		dcID:       opts.DC,
		codec:      opts.Codec,
		key:        key,
		cipher:     crypto.NewServerCipher(opts.Random),
		clock:      opts.Clock,
		log:        opts.Logger,
		dispatcher: NewDispatcher(),
		users:      newUsers(),
		msgID:      opts.MessageID,
		types:      opts.Types,
	}
	return s
}

// Key returns public key of this server.
func (s *Server) Key() *rsa.PublicKey {
	return &s.key.PublicKey
}

// Serve runs server loop using given listener.
func (s *Server) Serve(ctx context.Context, l net.Listener) error {
	s.ctx = ctx
	return s.serve(l)
}

// Dispatcher returns server RPC dispatcher.
func (s *Server) Dispatcher() *Dispatcher {
	return s.dispatcher
}

func (s *Server) serve(listener net.Listener) error {
	s.log.Info("Serving")
	defer func() {
		s.log.Info("Stopping")
	}()

	// NB: s.codec may be nil.
	server := transport.NewCustomServer(s.codec, listener)
	defer func() {
		_ = server.Close()
	}()
	return server.Serve(s.ctx, func(ctx context.Context, conn transport.Conn) error {
		err := s.serveConn(ctx, conn)
		if err != nil {
			// Client disconnected.
			var syscallErr *net.OpError
			if xerrors.Is(err, io.EOF) || xerrors.As(err, &syscallErr) &&
				(syscallErr.Op == "write" || syscallErr.Op == "read") {
				return nil
			}
			s.log.Info("Serving handler error", zap.Error(err))
		}
		return err
	})
}
