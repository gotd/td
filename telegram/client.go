package telegram

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/crypto"
	"github.com/ernado/td/internal/proto"
	"github.com/ernado/td/tg"
)

// Client represents a MTProto client to Telegram.
type Client struct {
	conn      net.Conn
	clock     func() time.Time
	authKey   crypto.AuthKey
	authKeyID [8]byte
	salt      int64
	session   int64
	rand      io.Reader
	seq       int
	log       *zap.Logger

	updateHandler func(u *tg.Updates) error

	// callbacks for rpc requests
	rpcMux sync.Mutex
	rpc    map[crypto.MessageID]func(b *bin.Buffer, rpcErr error)

	// callbacks for ping results protected by pingMux
	pingMux sync.Mutex
	ping    map[int64]func()

	rsaPublicKeys []*rsa.PublicKey
}

const defaultTimeout = time.Second * 10

func (c *Client) startIntermediateMode(deadline time.Time) error {
	if err := c.conn.SetDeadline(deadline); err != nil {
		return xerrors.Errorf("failed to set deadline: %w", err)
	}
	if _, err := c.conn.Write(proto.IntermediateClientStart); err != nil {
		return xerrors.Errorf("failed to write start: %w", err)
	}
	if err := c.conn.SetDeadline(time.Time{}); err != nil {
		return xerrors.Errorf("failed to reset connection deadline: %w", err)
	}
	return nil
}

func (c *Client) resetDeadline() error {
	return c.conn.SetDeadline(time.Time{})
}

func (c *Client) deadline(ctx context.Context) time.Time {
	if deadline, ok := ctx.Deadline(); ok {
		return deadline
	}
	return c.clock().Add(defaultTimeout)
}

func (c *Client) newUnencryptedMessage(payload bin.Encoder, b *bin.Buffer) error {
	b.Reset()
	if err := payload.Encode(b); err != nil {
		return err
	}
	msg := proto.UnencryptedMessage{
		MessageID:   crypto.NewMessageID(c.clock(), crypto.MessageFromClient),
		MessageData: b.Copy(),
	}
	b.Reset()
	return msg.Encode(b)
}

func (c *Client) AuthKey() crypto.AuthKey {
	return c.authKey
}

// Options of Client.
type Options struct {
	// Required options:

	// PublicKeys of telegram.
	PublicKeys []*rsa.PublicKey
	// Addr to connect.
	Addr string

	// Optional:

	// Dialer to use. Default dialer will be used if not provided.
	Dialer *net.Dialer
	// Network to use. Defaults to tcp.
	Network string
	// Random is random source. Defaults to crypto.
	Random io.Reader
	// Logger is instance of zap.Logger. No logs by default.
	Logger *zap.Logger
}

// Dial initializes Client and creates connection to Telegram.
//
// Note that no data is send or received during this process.
func Dial(ctx context.Context, opt Options) (*Client, error) {
	if opt.Dialer == nil {
		opt.Dialer = &net.Dialer{}
	}
	if opt.Network == "" {
		opt.Network = "tcp"
	}
	if opt.Random == nil {
		opt.Random = rand.Reader
	}
	if opt.Logger == nil {
		opt.Logger = zap.NewNop()
	}
	if len(opt.PublicKeys) == 0 {
		// Using public keys that are included with distribution if not
		// provided.
		//
		// This should never fail and keys should be valid for recent
		// library versions.
		keys, err := vendoredKeys()
		if err != nil {
			return nil, xerrors.Errorf("failed to load vendored keys: %w", err)
		}
		opt.PublicKeys = keys
	}
	conn, err := opt.Dialer.DialContext(ctx, "tcp", opt.Addr)
	if err != nil {
		return nil, xerrors.Errorf("failed to dial: %w", err)
	}
	client := &Client{
		conn:  conn,
		clock: time.Now,
		rand:  opt.Random,
		log:   opt.Logger,
		ping:  map[int64]func(){},
		rpc:   map[crypto.MessageID]func(b *bin.Buffer, rpcErr error){},

		rsaPublicKeys: opt.PublicKeys,
	}
	return client, nil
}

// Connect establishes connection in intermediate mode, creating new auth key
// if needed.
func (c *Client) Connect(ctx context.Context) error {
	deadline := c.clock().Add(defaultTimeout)
	if ctxDeadline, ok := ctx.Deadline(); ok {
		deadline = ctxDeadline
	}
	if err := c.startIntermediateMode(deadline); err != nil {
		return xerrors.Errorf("failed to initialize intermediate protocol: %w", err)
	}
	if c.authKey.Zero() {
		c.log.Info("Generating new auth key")
		start := c.clock()
		if err := c.createAuthKey(ctx); err != nil {
			return xerrors.Errorf("unable to create auth key: %w", err)
		}
		c.log.With(zap.Duration("duration", c.clock().Sub(start))).Info("Auth key generated")
	}

	// Spawning reading goroutine.
	// Probably we should use another ctx here.
	go c.readLoop(ctx)

	// Simple way to test that everything is ok.
	// Blocks until pong is received.
	if err := c.Ping(ctx); err != nil {
		return err
	}

	return nil
}
