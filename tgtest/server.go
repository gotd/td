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
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/transport"
)

// Server is a MTProto server structure.
type Server struct {
	// DC ID of this server.
	dcID int
	// Key pair of this server.
	key *rsa.PrivateKey // immutable

	// Codec constructor. May be nil.
	codec func() transport.Codec // immutable,nilable
	// Server-side message cipher.
	cipher crypto.Cipher // immutable
	// Clock to use in key exchange and message ID generation.
	clock clock.Clock // immutable
	// MessageID generator
	msgID mtproto.MessageIDSource // immutable

	// RPC handler.
	handler Handler // immutable

	// users stores session info.
	users *users

	// type map for logging.
	types *tmap.Map   // immutable
	log   *zap.Logger // immutable
}

// NewServer creates new Server.
func NewServer(key *rsa.PrivateKey, handler Handler, opts ServerOptions) *Server {
	opts.setDefaults()

	s := &Server{
		dcID:    opts.DC,
		codec:   opts.Codec,
		key:     key,
		cipher:  crypto.NewServerCipher(opts.Random),
		clock:   opts.Clock,
		log:     opts.Logger,
		users:   newUsers(),
		handler: handler,
		msgID:   opts.MessageID,
		types:   opts.Types,
	}
	return s
}

// Key returns public key of this server.
func (s *Server) Key() *rsa.PublicKey {
	return &s.key.PublicKey
}

// Serve runs server loop using given listener.
func (s *Server) Serve(ctx context.Context, l transport.Listener) error {
	return s.serve(ctx, l)
}

func (s *Server) serve(ctx context.Context, l transport.Listener) error {
	s.log.Info("Serving")
	defer func() {
		s.log.Info("Stopping")
	}()

	grp := tdsync.NewCancellableGroup(ctx)
	grp.Go(func(context.Context) error {
		for {
			conn, err := l.Accept()
			if err != nil {
				if xerrors.Is(err, net.ErrClosed) {
					return nil
				}
				return xerrors.Errorf("accept: %w", err)
			}

			grp.Go(func(ctx context.Context) error {
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
	})
	grp.Go(func(ctx context.Context) error {
		<-ctx.Done()
		return l.Close()
	})
	return grp.Wait()
}
