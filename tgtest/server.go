package tgtest

import (
	"context"
	"io"
	"net"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"nhooyr.io/websocket"

	"github.com/nnqq/td/clock"
	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/internal/exchange"
	"github.com/nnqq/td/internal/mtproto"
	"github.com/nnqq/td/internal/tdsync"
	"github.com/nnqq/td/internal/tmap"
	"github.com/nnqq/td/transport"
)

// Server is a MTProto server structure.
type Server struct {
	// DC ID of this server.
	dcID int
	// Key pair of this server.
	key exchange.PrivateKey // immutable

	// Codec constructor. May be nil.
	codec func() transport.Codec // immutable,nilable
	// Server-side message cipher.
	cipher crypto.Cipher // immutable
	// Clock to use in key exchange and message ID generation.
	clock clock.Clock // immutable
	// MessageID generator
	msgID mtproto.MessageIDSource // immutable

	readTimeout  time.Duration
	writeTimeout time.Duration

	// RPC handler.
	handler Handler // immutable

	// users stores session info.
	users *users

	// type map for logging.
	types *tmap.Map   // immutable
	log   *zap.Logger // immutable
}

// NewServer creates new Server.
func NewServer(key exchange.PrivateKey, handler Handler, opts ServerOptions) *Server {
	opts.setDefaults()

	s := &Server{
		dcID:         opts.DC,
		key:          key,
		codec:        opts.Codec,
		cipher:       crypto.NewServerCipher(opts.Random),
		clock:        opts.Clock,
		msgID:        opts.MessageID,
		readTimeout:  opts.ReadTimeout,
		writeTimeout: opts.WriteTimeout,
		handler:      handler,
		users:        newUsers(),
		types:        opts.Types,
		log:          opts.Logger,
	}
	return s
}

// Key returns public key of this server.
func (s *Server) Key() exchange.PublicKey {
	return s.key.Public()
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
				if err := s.serveConn(ctx, conn); err != nil {
					// Client disconnected.
					var syscallErr *net.OpError
					switch {
					case xerrors.Is(err, io.EOF):
						return nil
					case xerrors.As(err, &syscallErr) &&
						(syscallErr.Op == "write" || syscallErr.Op == "read"):
						return nil
					}
					// TODO(tdakkota): emulate errors too?
					if code := websocket.CloseStatus(err); code >= 0 {
						return nil
					}

					s.log.Info("Serving handler error", zap.Error(err))
				}
				return nil
			})
		}
	})
	grp.Go(func(ctx context.Context) error {
		<-ctx.Done()
		return l.Close()
	})
	return grp.Wait()
}
