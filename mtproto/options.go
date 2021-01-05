package mtproto

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"runtime"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/transport"
)

// Transport is MTProto connection creator.
type Transport interface {
	Codec() transport.Codec
	DialContext(ctx context.Context, network, address string) (transport.Conn, error)
}

// Options of Conn.
type Options struct {
	// PublicKeys of telegram.
	//
	// If not provided, embedded public keys will be used.
	PublicKeys []*rsa.PublicKey
	// Transport to use. Default dialer will be used if not provided.
	Transport Transport
	// Network to use. Defaults to tcp.
	Network string
	// Random is random source. Defaults to crypto.
	Random io.Reader
	// Logger is instance of zap.Logger. No logs by default.
	Logger *zap.Logger
	// Handler will be called on received message.
	Handler Handler
	// AckBatchSize is maximum ack-s to buffer.
	AckBatchSize int
	// AckInterval is maximum time to buffer ack.
	AckInterval time.Duration
	// RetryInterval is duration between retries.
	RetryInterval time.Duration
	// MaxRetries is max retry count until rpc request failure.
	MaxRetries int
	// MessageID is message id source. Share source between connection to
	// reduce collision probability.
	MessageID MessageIDSource
	// Clock is current time source. Defaults to system time.
	Clock clock.Clock
	// Types map, used in verbose logging of incoming message.
	Types *tmap.Map
	// Key that can be used to restore previous connection.
	Key crypto.AuthKeyWithID
	// Salt from server that can be used to restore previous connection.
	Salt int64
	// ReadConcurrency limits maximum concurrently handled messages.
	// Can be CPU or IO bound depending on message handlers.
	// Defaults to GOMAXPROCS if it is not less than 10.
	ReadConcurrency int
}

type nopHandler struct{}

func (nopHandler) OnMessage(b *bin.Buffer) error   { return nil }
func (nopHandler) OnSession(session Session) error { return nil }

func (opt *Options) setDefaultPublicKeys() {
	// Using public keys that are included with distribution if not
	// provided.
	//
	// This should never fail and keys should be valid for recent
	// library versions.
	keys, err := vendoredKeys()
	if err != nil {
		panic(xerrors.Errorf("load vendored keys: %w", err))
	}
	opt.PublicKeys = keys
}

func (opt *Options) setDefaultConcurrency() {
	opt.ReadConcurrency = runtime.GOMAXPROCS(0)

	// In container environment small GOMAXPROCS are common (like 1 or 2),
	// but such low concurrency is unfortunate, because most calls will
	// be io bound.
	const minConcurrency = 10

	if opt.ReadConcurrency < minConcurrency {
		opt.ReadConcurrency = minConcurrency
	}
}

func (opt *Options) setDefaults() {
	if opt.Transport == nil {
		opt.Transport = transport.Intermediate(nil)
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
	if opt.AckBatchSize == 0 {
		opt.AckBatchSize = 20
	}
	if opt.AckInterval == 0 {
		opt.AckInterval = time.Second * 15
	}
	if opt.RetryInterval == 0 {
		opt.RetryInterval = time.Second * 5
	}
	if opt.MaxRetries == 0 {
		opt.MaxRetries = 5
	}
	if opt.Clock == nil {
		opt.Clock = clock.System
	}
	if opt.MessageID == nil {
		opt.MessageID = proto.NewMessageIDGen(opt.Clock.Now, 100)
	}
	if len(opt.PublicKeys) == 0 {
		opt.setDefaultPublicKeys()
	}
	if opt.Handler == nil {
		opt.Handler = nopHandler{}
	}
	if opt.ReadConcurrency == 0 {
		opt.setDefaultConcurrency()
	}
}
