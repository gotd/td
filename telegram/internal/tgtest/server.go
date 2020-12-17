// Package tgtest provides test Telegram server for end-to-end test.
package tgtest

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/net/nettest"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/proto"
)

type TB interface {
	Log(args ...interface{})
	Logf(format string, args ...interface{})
}

type Server struct {
	Listener net.Listener

	key     *rsa.PrivateKey
	cipher  crypto.Cipher
	handler Handler

	wg sync.WaitGroup

	mux    sync.Mutex // guards closed and tb
	tb     TB
	closed bool

	conns *conns
}

func (s *Server) Key() *rsa.PublicKey {
	return &s.key.PublicKey
}

func (s *Server) Start() {
	s.wg.Add(1)
	go s.serve()
}

func (s *Server) Close() {
	s.mux.Lock()
	s.closed = true
	s.mux.Unlock()
	_ = s.Listener.Close()
	s.wg.Wait()
}

func NewServer(tb TB, h Handler) *Server {
	s := NewUnstartedServer(tb)
	s.SetHandler(h)
	s.Start()
	return s
}

func NewUnstartedServer(tb TB) *Server {
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	s := &Server{
		Listener: newLocalListener(),
		cipher:   crypto.NewServerCipher(rand.Reader),
		key:      k,
		tb:       tb,
		conns:    newConns(),
	}
	return s
}

func newLocalListener() net.Listener {
	l, err := nettest.NewLocalListener("tcp")
	if err != nil {
		panic(fmt.Sprintf("tgtest: failed to listen on a port: %v", err))
	}
	return l
}

func (s *Server) SetHandler(handler Handler) {
	s.handler = handler
}

func (s *Server) writeUnencrypted(conn net.Conn, data bin.Encoder) error {
	b := &bin.Buffer{}
	if err := data.Encode(b); err != nil {
		return err
	}
	msg := proto.UnencryptedMessage{
		MessageID:   int64(proto.NewMessageID(time.Now(), proto.MessageServerResponse)),
		MessageData: b.Copy(),
	}
	b.Reset()
	if err := msg.Encode(b); err != nil {
		return err
	}

	return proto.WriteIntermediate(conn, b)
}

func (s *Server) readUnencrypted(conn net.Conn, data bin.Decoder) error {
	b := &bin.Buffer{}
	if err := proto.ReadIntermediate(conn, b); err != nil {
		return err
	}
	var msg proto.UnencryptedMessage
	if err := msg.Decode(b); err != nil {
		return err
	}
	if err := s.checkMsgID(msg.MessageID); err != nil {
		return err
	}
	b.ResetTo(msg.MessageData)

	return data.Decode(b)
}

func (s *Server) rpcHandle(k Session, conn net.Conn) error {
	var b bin.Buffer
	for {
		b.Reset()
		if err := proto.ReadIntermediate(conn, &b); err != nil {
			return xerrors.Errorf("failed to read intermediate: %w", err)
		}

		msg, err := s.cipher.DecryptDataFrom(k.Key, 0, &b)
		if err != nil {
			return xerrors.Errorf("failed to decrypt: %w", err)
		}
		k.SessionID = msg.SessionID

		// Buffer now contains plaintext message payload.
		b.ResetTo(msg.Data())

		if err := s.handler.OnMessage(k, msg.MessageID, &b); err != nil {
			return xerrors.Errorf("failed to call handler: %w", err)
		}
	}
}

func (s *Server) serveConn(conn net.Conn) error {
	var session Session
	defer func() {
		s.conns.delete(session)
		_ = conn.Close()
	}()

	buf := make([]byte, len(proto.IntermediateClientStart))
	if _, err := conn.Read(buf); err != nil {
		return err
	}

	if !bytes.Equal(buf, proto.IntermediateClientStart) {
		return errors.New("unexpected intermediate client start")
	}

	session, err := s.exchange(conn)
	if err != nil {
		return xerrors.Errorf("key exchange failed: %w", err)
	}
	s.conns.add(session, conn)

	err = s.handler.OnNewClient(session)
	if err != nil {
		return xerrors.Errorf("OnNewClient handler failed: %w", err)
	}

	return s.rpcHandle(session, conn)
}

func (s *Server) checkMsgID(id int64) error {
	if proto.MessageID(id).Type() != proto.MessageFromClient {
		return errors.New("bad msg type")
	}
	return nil
}

func (s *Server) serve() {
	defer s.wg.Done()

	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			return
		}
		s.mux.Lock()
		closed := s.closed
		s.mux.Unlock()
		if closed {
			break
		}
		go func() {
			if err := s.serveConn(conn); err != nil {
				s.mux.Lock()
				if !s.closed {
					s.tb.Log(err)
				}
				s.mux.Unlock()
			}
		}()
	}
}
