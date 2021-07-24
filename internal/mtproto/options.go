package mtproto

import (
	"crypto/rand"
	"crypto/rsa"
	"io"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/exchange"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/rpc"
	"github.com/gotd/td/internal/tmap"
)

// Options of Conn.
type Options struct {
	// PublicKeys of telegram.
	//
	// If not provided, embedded public keys will be used.
	PublicKeys []*rsa.PublicKey
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
	// DialTimeout is timeout of creating connection.
	DialTimeout time.Duration
	// ExchangeTimeout is timeout of every key exchange request.
	ExchangeTimeout time.Duration
	// SaltFetchInterval is duration between get_future_salts request.
	SaltFetchInterval time.Duration
	// PingTimeout sets ping_delay_disconnect timeout.
	PingTimeout time.Duration
	// PingInterval is duration between ping_delay_disconnect request.
	PingInterval time.Duration
	// RequestTimeout is function which returns request timeout for given type ID.
	RequestTimeout func(req uint32) time.Duration
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
	Key crypto.AuthKey
	// Salt from server that can be used to restore previous connection.
	Salt int64
	// ReadConcurrency limits maximum concurrently handled messages.
	// Can be CPU or IO bound depending on message handlers.
	// Defaults to GOMAXPROCS if it is not less than 10.
	//
	// Deprecated: no longer used.
	ReadConcurrency int
	// NoBufferReuse disables buffer reuse for concurrently handled messages.
	// Each concurrent message handler adds around 0.5 MiB to the total used
	// memory. The default behavior is to keep the buffer for reuse, but in
	// programs that maintain many connections this quickly adds up to hundreds
	// of megabytes (buffer size multiplied by the number of connections and read
	// concurrency).
	//
	// Deprecated: no longer needed, will be removed in v0.48.
	NoBufferReuse bool
	// Cipher defines message crypto.
	Cipher Cipher

	// Private options.

	// engine for replacing RPC engine.
	engine *rpc.Engine
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

func (opt *Options) setDefaults() {
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
		opt.AckInterval = 15 * time.Second
	}
	if opt.RetryInterval == 0 {
		opt.RetryInterval = 5 * time.Second
	}
	if opt.DialTimeout == 0 {
		opt.DialTimeout = 15 * time.Second
	}
	if opt.ExchangeTimeout == 0 {
		opt.ExchangeTimeout = exchange.DefaultTimeout
	}
	if opt.SaltFetchInterval == 0 {
		opt.SaltFetchInterval = 1 * time.Hour
	}
	if opt.PingTimeout == 0 {
		opt.PingTimeout = 15 * time.Second
	}
	if opt.PingInterval == 0 {
		opt.PingInterval = 1 * time.Minute
	}
	if opt.RequestTimeout == nil {
		opt.RequestTimeout = func(req uint32) time.Duration {
			return 15 * time.Second
		}
	}
	if opt.MaxRetries == 0 {
		opt.MaxRetries = 5
	}
	if opt.Clock == nil {
		opt.Clock = clock.System
	}
	if opt.MessageID == nil {
		opt.MessageID = proto.NewMessageIDGen(opt.Clock.Now)
	}
	if len(opt.PublicKeys) == 0 {
		opt.setDefaultPublicKeys()
	}
	if opt.Handler == nil {
		opt.Handler = nopHandler{}
	}
	if opt.Cipher == nil {
		opt.Cipher = crypto.NewClientCipher(opt.Random)
	}
}
