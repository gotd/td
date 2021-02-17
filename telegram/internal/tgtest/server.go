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

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/transport"
)

type Server struct {
	server *transport.Server

	key     *rsa.PrivateKey
	cipher  crypto.Cipher
	handler Handler

	ctx   context.Context
	clock clock.Clock
	log   *zap.Logger
	msgID *proto.MessageIDGen

	users *users
}

func (s *Server) Key() *rsa.PublicKey {
	return &s.key.PublicKey
}

func (s *Server) Addr() net.Addr {
	return s.server.Addr()
}

func (s *Server) Serve() error {
	return s.serve()
}

func (s *Server) AddSession(key crypto.AuthKey) {
	s.users.addSession(key)
}

func (s *Server) Start() {
	go func() {
		_ = s.Serve()
	}()
}

func (s *Server) Close() {
	_ = s.server.Close()
}

func NewServer(name string, suite Suite, codec func() transport.Codec, h Handler) *Server {
	s := NewUnstartedServer(name, suite, codec)
	s.SetHandler(h)
	s.Start()
	return s
}

func NewUnstartedServer(name string, suite Suite, codec func() transport.Codec) *Server {
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	log := suite.Log.Named(name)

	s := &Server{
		server: transport.NewCustomServer(codec, newLocalListener(suite.Ctx)),
		key:    k,
		cipher: crypto.NewServerCipher(rand.Reader),
		ctx:    suite.Ctx,
		clock:  clock.System,
		log:    log,
		users:  newUsers(),
		msgID:  proto.NewMessageIDGen(clock.System.Now, 100),
	}
	return s
}

func (s *Server) SetHandler(handler Handler) {
	s.handler = handler
}

func (s *Server) SetHandlerFunc(handler func(s Session, msgID int64, in *bin.Buffer) error) {
	s.handler = HandlerFunc(handler)
}

func (s *Server) serve() error {
	s.log.Info("Serving", zap.String("addr", s.Addr().String()))
	defer func() {
		l := s.log
		if err := s.ctx.Err(); err != nil {
			l = s.log.With(zap.Error(err))
		}
		l.Info("Stopping")
	}()

	return s.server.Serve(s.ctx, func(ctx context.Context, conn transport.Conn) error {
		err := s.serveConn(ctx, conn)
		if err != nil {
			// Client disconnected.
			var syscallErr *net.OpError
			if xerrors.Is(err, io.EOF) ||
				xerrors.As(err, &syscallErr) && syscallErr.Op == "write" ||
				syscallErr.Op == "read" {
				return nil
			}
			s.log.Info("Serving handler error", zap.Error(err))
		}
		return err
	})
}
