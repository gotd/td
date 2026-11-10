package mtproto

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log/logzap"
	"github.com/gotd/neo"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/exchange"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/testutil"
	"github.com/gotd/td/transport"
)

// TestHandleAuthKeyNotFoundResetsSession checks that regenerating the auth key
// in place after transport -404 starts a clean session: a pending request is
// sent again with a new msg_id and the first seqno of the new session, and the
// received msg_id buffer of the old session is forgotten.
func TestHandleAuthKeyNotFoundResetsSession(t *testing.T) {
	a := require.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	privateKey, err := rsa.GenerateKey(rand.Reader, crypto.RSAKeyBits)
	a.NoError(err)
	serverKey := exchange.PrivateKey{RSA: privateKey}

	var staleKey crypto.Key
	_, err = io.ReadFull(rand.Reader, staleKey[:])
	a.NoError(err)

	const retryInterval = time.Second
	clk := neo.NewTime(testutil.Date())
	logger := logzap.New(zaptest.NewLogger(t))
	client, server := transport.Intermediate.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	c := New(nil, Options{
		DC:            2,
		PublicKeys:    []exchange.PublicKey{serverKey.Public()},
		Random:        rand.Reader,
		Clock:         clk,
		Logger:        logger,
		RetryInterval: retryInterval,
		Key:           staleKey.WithID(),
		Salt:          10,
	})
	defer c.rpc.ForceClose()
	c.conn = client
	c.sessionID = 1
	// State of the old session.
	c.sentContentMessages = 5
	ids := proto.NewMessageIDGen(clk.Now)
	receivedID := ids.New(proto.MessageServerResponse)
	a.True(c.messageIDBuf.Consume(receivedID))

	serverCipher := crypto.NewServerCipher(rand.Reader)
	recv := func(key crypto.AuthKey) *crypto.EncryptedMessageData {
		t.Helper()
		var b bin.Buffer
		a.NoError(server.Recv(ctx, &b))
		msg, err := serverCipher.DecryptFromBuffer(key, &b)
		a.NoError(err)
		return msg
	}

	invokeDone := make(chan error, 1)
	go func() {
		invokeDone <- c.Invoke(ctx, &mt.PingRequest{PingID: 1}, &mt.Pong{})
	}()

	// Request sent in the old session, server does not know the key.
	old := recv(c.session().Key)
	a.Equal(int64(1), old.SessionID)
	a.Equal(int32(11), old.SeqNo)

	// Server replied -404, client regenerates the key in place.
	handled := make(chan error, 1)
	go func() { handled <- c.handleAuthKeyNotFound(ctx) }()
	res, err := exchange.NewExchanger(server, 2).
		WithLogger(logger).
		WithRand(rand.Reader).
		Server(serverKey).
		Run(ctx)
	a.NoError(err)
	a.NoError(<-handled)
	newSession := c.session()
	a.Equal(res.Key.ID, newSession.Key.ID)
	a.NotEqual(int64(1), newSession.ID)
	// Let a retry of the old message fire, if it is still pending.
	clk.Travel(retryInterval * 2)

	msg := recv(res.Key)
	a.Equal(newSession.ID, msg.SessionID)
	a.NotEqual(old.MessageID, msg.MessageID, "request must be sent with a new msg_id")
	a.Equal(int32(1), msg.SeqNo, "first content message of the new session")
	var ping mt.PingRequest
	a.NoError(ping.Decode(&bin.Buffer{Buf: msg.Data()}))
	a.Equal(int64(1), ping.PingID)

	a.True(c.messageIDBuf.Consume(receivedID), "msg_id buffer of the old session must be reset")

	cancel()
	_ = client.Close()
	a.ErrorIs(<-invokeDone, context.Canceled)
}

// sentConn is a transport.Conn that passes sent frames to a channel.
type sentConn struct {
	sent chan *bin.Buffer
}

func (s *sentConn) Send(ctx context.Context, b *bin.Buffer) error {
	select {
	case s.sent <- &bin.Buffer{Buf: append([]byte(nil), b.Buf...)}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *sentConn) Recv(ctx context.Context, b *bin.Buffer) error {
	<-ctx.Done()
	return ctx.Err()
}

func (s *sentConn) Close() error { return nil }

// notifyIDs is a MessageIDSource that reports every allocated id.
type notifyIDs struct {
	gen       *proto.MessageIDGen
	allocated chan int64
}

func (n notifyIDs) New(t proto.MessageType) int64 {
	id := n.gen.New(t)
	select {
	case n.allocated <- id:
	default:
	}
	return id
}

func newResetTestConn(t *testing.T) (*Conn, *sentConn, notifyIDs) {
	t.Helper()
	clk := neo.NewTime(testutil.Date())
	ids := notifyIDs{
		gen:       proto.NewMessageIDGen(clk.Now),
		allocated: make(chan int64, 1),
	}
	var key crypto.Key
	_, err := io.ReadFull(rand.Reader, key[:])
	require.NoError(t, err)

	c := New(nil, Options{
		Random:    rand.Reader,
		Clock:     clk,
		Logger:    logzap.New(zaptest.NewLogger(t)),
		Key:       key.WithID(),
		MessageID: ids,
	})
	t.Cleanup(c.rpc.ForceClose)
	conn := &sentConn{sent: make(chan *bin.Buffer, 10)}
	c.conn = conn
	c.sessionID = 1
	c.sentContentMessages = 5
	return c, conn, ids
}

// replaceSession emulates createAuthKey: swaps session and resets its state.
// Must be called with exchangeLock held.
func replaceSession(c *Conn) {
	c.sessionMux.Lock()
	c.sessionID = 2
	c.sessionMux.Unlock()
	c.resetSession()
}

// TestServiceMessageDuringExchange checks that a service message waiting for
// the exchange lock while the session is reset is numbered in the new session,
// otherwise its stale seqno makes the server reject the next content messages
// with "msg_seqno too low".
func TestServiceMessageDuringExchange(t *testing.T) {
	a := require.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	c, conn, ids := newResetTestConn(t)

	c.exchangeLock.Lock()
	written := make(chan error, 1)
	go func() {
		written <- c.writeServiceMessage(ctx, &mt.PingRequest{PingID: 2})
	}()
	// Give the writer a chance to allocate its ids before the reset, as it did
	// when ids were allocated outside of the lock.
	select {
	case <-ids.allocated:
	case <-time.After(100 * time.Millisecond):
	}
	replaceSession(c)
	c.exchangeLock.Unlock()
	a.NoError(<-written)

	msg, err := crypto.NewServerCipher(rand.Reader).DecryptFromBuffer(c.session().Key, <-conn.sent)
	a.NoError(err)
	a.Equal(int64(2), msg.SessionID)
	a.Equal(int32(0), msg.SeqNo, "service message must be numbered in the new session")
}

// TestStaleContentMessageAfterReset checks that a content message whose msg_id
// and seqno were allocated in the old session is not sent in the new one.
func TestStaleContentMessageAfterReset(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	c, conn, _ := newResetTestConn(t)

	msgID, seqNo := c.nextMsgSeq(true)
	a.Equal(int32(11), seqNo)

	c.exchangeLock.Lock()
	replaceSession(c)
	c.exchangeLock.Unlock()

	err := c.writeContentMessage(ctx, msgID, seqNo, &mt.PingRequest{PingID: 3})
	a.ErrorIs(err, errSessionReset)
	a.Empty(conn.sent)

	_, seqNo = c.nextMsgSeq(true)
	a.Equal(int32(1), seqNo)
}
