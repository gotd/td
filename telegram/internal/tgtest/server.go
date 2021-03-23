// Package tgtest provides test Telegram server for end-to-end test.
package tgtest

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

type Server struct {
	dcID   int
	codec  func() transport.Codec // immutable
	key    *rsa.PrivateKey        // immutable
	cipher crypto.Cipher          // immutable

	dispatcher *Dispatcher
	ctx        context.Context

	clock clock.Clock         // immutable
	log   *zap.Logger         // immutable
	msgID *proto.MessageIDGen // immutable

	users *users
	types *tmap.Map
}

func (s *Server) Key() *rsa.PublicKey {
	return &s.key.PublicKey
}

func (s *Server) Serve(ctx context.Context, l net.Listener) error {
	s.ctx = ctx
	return s.serve(l)
}

func (s *Server) AddSession(key crypto.AuthKey) {
	s.users.addSession(key)
}

func NewUnstartedServer(dcID int, log *zap.Logger, codec func() transport.Codec) *Server {
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	s := &Server{
		dcID:       dcID,
		codec:      codec,
		key:        k,
		cipher:     crypto.NewServerCipher(rand.Reader),
		clock:      clock.System,
		log:        log,
		dispatcher: NewDispatcher(),
		users:      newUsers(),
		msgID:      proto.NewMessageIDGen(clock.System.Now, 100),
		types: tmap.New(
			tg.TypesMap(),
			mt.TypesMap(),
			proto.TypesMap(),
		),
	}
	return s
}

func (s *Server) Dispatcher() *Dispatcher {
	return s.dispatcher
}

func (s *Server) serve(listener net.Listener) error {
	s.log.Info("Serving")
	defer func() {
		s.log.Info("Stopping")
	}()

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
