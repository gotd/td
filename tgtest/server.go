package tgtest

import (
	"context"
	"crypto/rsa"
	"io"
	"net"
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"nhooyr.io/websocket"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/exchange"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/tmap"
	"github.com/gotd/td/transport"
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

// NewPrivateKey creates new private key from RSA private key.
func NewPrivateKey(k *rsa.PrivateKey) exchange.PrivateKey {
	return exchange.PrivateKey{
		RSA: k,
	}
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
				if errors.Is(err, net.ErrClosed) {
					return nil
				}
				return errors.Wrap(err, "accept")
			}

			grp.Go(func(ctx context.Context) error {
				if err := s.serveConn(ctx, conn); err != nil {
					// Client disconnected.
					var syscallErr *net.OpError
					switch {
					case errors.Is(err, io.EOF):
						return nil
					case errors.As(err, &syscallErr) &&
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
