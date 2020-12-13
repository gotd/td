// Package tgtest provides test Telegram server for end-to-end test.
package tgtest

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
)

type TB interface {
	Log(args ...interface{})
	Logf(format string, args ...interface{})
}

type Server struct {
	Listener net.Listener

	key *rsa.PrivateKey

	wg sync.WaitGroup

	mux    sync.Mutex // guards closed and tb
	tb     TB
	closed bool
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

func NewServer(tb TB) *Server {
	s := NewUnstartedServer(tb)
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

		key: k,
		tb:  tb,
	}
	return s
}

func newLocalListener() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("tgtest: failed to listen on a port: %v", err))
		}
	}
	return l
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

func (s *Server) serveConn(conn net.Conn) error {
	defer func() { _ = conn.Close() }()

	buf := make([]byte, len(proto.IntermediateClientStart))
	if _, err := conn.Read(buf); err != nil {
		return err
	}

	if !bytes.Equal(buf, proto.IntermediateClientStart) {
		return errors.New("unexpected inermediate client start")
	}

	var pqReq mt.ReqPqMulti
	if err := s.readUnencrypted(conn, &pqReq); err != nil {
		return err
	}

	pq := big.NewInt(0x17ED48941A08F981)
	if err := s.writeUnencrypted(conn, &mt.ResPQ{
		Pq:    pq.Bytes(),
		Nonce: pqReq.Nonce,
		ServerPublicKeyFingerprints: []int64{
			crypto.RSAFingerprint(s.Key()),
		},
	}); err != nil {
		return err
	}

	// TODO(ernado): make actual crypto here
	var dhParams mt.ReqDHParams
	if err := s.readUnencrypted(conn, &dhParams); err != nil {
		return err
	}
	if err := s.writeUnencrypted(conn, &mt.ServerDHParamsOk{}); err != nil {
		return err
	}

	return nil
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
