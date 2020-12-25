// Package tgtest provides test Telegram server for end-to-end test.
package tgtest

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"net"
	"time"

	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/transport"
)

type Server struct {
	server *transport.Server

	key     *rsa.PrivateKey
	cipher  crypto.Cipher
	handler Handler

	clock  func() time.Time
	ctx    context.Context
	cancel context.CancelFunc
	log    *zap.Logger

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

func (s *Server) Start() {
	go func() {
		_ = s.Serve()
	}()
}

func (s *Server) Close() {
	if s.cancel != nil {
		s.cancel()
	}

	_ = s.server.Close()
}

func NewServer(suite Suite, codec func() transport.Codec, h Handler) *Server {
	s := NewUnstartedServer(suite, codec)
	s.SetHandler(h)
	s.Start()
	return s
}

func NewUnstartedServer(suite Suite, codec func() transport.Codec) *Server {
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(suite.Ctx)
	s := &Server{
		server: transport.NewCustomServer(codec, newLocalListener()),
		key:    k,
		cipher: crypto.NewServerCipher(rand.Reader),
		clock:  time.Now,
		ctx:    ctx,
		cancel: cancel,
		log:    suite.Log.Named("server"),
		users:  newUsers(),
	}
	return s
}

func (s *Server) SetHandler(handler Handler) {
	s.handler = handler
}

func (s *Server) SetHandlerFunc(handler func(s Session, msgID int64, in *bin.Buffer) error) {
	s.handler = HandlerFunc(handler)
}

func (s *Server) checkMsgID(id int64) error {
	if proto.MessageID(id).Type() != proto.MessageFromClient {
		return errors.New("bad msg type")
	}
	return nil
}

func (s *Server) serve() error {
	return s.server.Serve(s.ctx, s.serveConn)
}
