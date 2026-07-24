package telegram

import (
	"context"
	"io"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.opentelemetry.io/otel/trace"

	"github.com/gotd/log"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/exchange"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
)

type (
	// PublicKey is a Telegram server public key.
	PublicKey = exchange.PublicKey
)

// Options of Client.
type Options struct {
	// PublicKeys of telegram.
	//
	// If not provided, embedded public keys will be used.
	PublicKeys []PublicKey

	// DC ID to connect.
	//
	// If not provided, 2 will be used by default.
	DC int

	// DCList is initial list of addresses to connect.
	DCList dcs.List

	// Resolver to use.
	Resolver dcs.Resolver

	// NoUpdates enables no updates mode.
	//
	// Enabled by default if no UpdateHandler is provided.
	NoUpdates bool

	// AllowCDN enables downloader CDN redirect flow for clients that support
	// downloader integration.
	//
	// If false, downloader will stay on the master DC path (legacy behavior).
	// If true and server does not return redirect, requests still go through the
	// same master path (no extra CDN round-trips).
	// Default is false.
	AllowCDN bool
	// ReconnectionBackoff configures and returns reconnection backoff object.
	ReconnectionBackoff func() backoff.BackOff
	// OnDead will be called on connection dead.
	OnDead func(error)
	// OnConnectionState is called when the primary connection state changes:
	// on every (re)connect start, on successful initialization and on
	// connection loss. Connection death details are reported via OnDead.
	//
	// Called synchronously from connection lifecycle, so the callback must
	// not block.
	OnConnectionState func(ConnectionState)
	// MigrationTimeout configures migration timeout.
	MigrationTimeout time.Duration

	// Random is random source. Defaults to crypto.
	Random io.Reader
	// Logger is the structured logger. No logs by default.
	//
	// Use github.com/gotd/log/logzap to bridge a *zap.Logger:
	//
	//	import (
	//		"go.uber.org/zap"
	//		"github.com/gotd/log/logzap"
	//	)
	//
	//	zapLog, _ := zap.NewProduction()
	//	opts := telegram.Options{
	//		Logger: logzap.New(zapLog),
	//	}
	Logger log.Logger
	// SessionStorage will be used to load and save session data.
	// NB: Very sensitive data, save with care.
	SessionStorage SessionStorage
	// UpdateHandler will be called on received update.
	UpdateHandler UpdateHandler
	// Middlewares list allows wrapping tg.Invoker. Can be useful for metrics,
	// tracing, etc. Note that order is important, see ExampleMiddleware.
	//
	// Middlewares are called in order from first to last.
	Middlewares []Middleware

	// AckBatchSize is limit of MTProto ACK buffer size.
	AckBatchSize int
	// AckInterval is maximum time to buffer MTProto ACK.
	AckInterval time.Duration
	// RetryInterval is duration between send retries.
	RetryInterval time.Duration
	// MaxRetries is limit of send retries.
	MaxRetries int
	// ExchangeTimeout is timeout of every key exchange request.
	ExchangeTimeout time.Duration
	// DialTimeout is timeout of creating connection.
	DialTimeout time.Duration
	// PingInterval is the duration between ping_delay_disconnect requests.
	//
	// Zero value means default (1 minute).
	PingInterval time.Duration
	// PingTimeout is how long to wait for a pong before considering the
	// connection dead.
	//
	// Zero value means default (15 seconds).
	PingTimeout time.Duration
	// PingDelayDisconnect is the disconnect_delay value sent to the server.
	// Must exceed PingInterval.
	//
	// Zero value means default (PingInterval + PingTimeout).
	PingDelayDisconnect time.Duration
	// IdleTimeout is the maximum duration without any received data before the
	// connection is closed and reconnected.
	//
	// Zero value means default (PingDelayDisconnect).
	IdleTimeout time.Duration
	// RetryOnWriteFailed retries a request whose transport send failed on the
	// next connection instead of returning the write error to the caller.
	//
	// Disabled by default, preserving the behavior of returning the error: a
	// caller that acts on it itself — rotating a proxy or an endpoint, for
	// example — needs to keep seeing it. Requests that were sent but not
	// acknowledged are retried regardless of this option.
	RetryOnWriteFailed bool
	// EnablePFS enables Perfect Forward Secrecy with temporary auth keys.
	EnablePFS bool
	// TempKeyTTL controls temporary key lifetime in seconds.
	// Default: 86400 (24h).
	// The value is clamped in mtproto layer to keep protocol-safe bounds.
	TempKeyTTL int

	// CompressThreshold is a threshold in bytes to determine that message
	// is large enough to be compressed using GZIP.
	// If < 0, compression will be disabled.
	// If == 0, default value will be used.
	CompressThreshold int

	// Device is device config.
	// Will be sent with session creation request.
	Device DeviceConfig

	// Layer is the schema layer to request via invokeWithLayer.
	//
	// If not provided, tg.Layer will be used, i.e. the layer of the schema
	// this package is generated from. Override it only if the server must be
	// asked for an older layer; requests are still encoded using tg.Layer
	// schema, so a mismatch may cause decoding errors.
	Layer int

	MessageID mtproto.MessageIDSource
	Clock     clock.Clock

	// OpenTelemetry.
	TracerProvider trace.TracerProvider

	// OnTransfer is called during authorization transfer.
	// See [AuthTransferHandler] for details.
	OnTransfer AuthTransferHandler

	// OnSelfError is called when client receives error calling Self() on connect.
	// Return error to stop reconnection.
	//
	// NB: this method is called immediately after connection, so it's not expected to be
	// non-nil error on first connection before auth, so it's safe to return nil until
	// first successful auth.
	OnSelfError func(ctx context.Context, err error) error

	// OnSelfSuccess is called when client get self calling Self() on connect.
	OnSelfSuccess func(self *tg.User)
}

func (opt *Options) setDefaults() {
	if opt.Resolver == nil {
		opt.Resolver = dcs.DefaultResolver()
	}
	if opt.Random == nil {
		opt.Random = crypto.DefaultRand()
	}
	if opt.Logger == nil {
		opt.Logger = log.Nop
	}
	if opt.DC == 0 {
		opt.DC = 2
	}
	if opt.DCList.Zero() {
		opt.DCList = dcs.Prod()
	}
	if opt.Layer == 0 {
		opt.Layer = tg.Layer
	}
	// It's okay to use zero value AckBatchSize, mtproto.Options will set defaults.
	// It's okay to use zero value AckInterval, mtproto.Options will set defaults.
	// It's okay to use zero value RetryInterval, mtproto.Options will set defaults.
	// It's okay to use zero value MaxRetries, mtproto.Options will set defaults.
	// It's okay to use zero value CompressThreshold, mtproto.Options will set defaults.
	opt.Device.SetDefaults()
	if opt.Clock == nil {
		opt.Clock = clock.System
	}
	if opt.ReconnectionBackoff == nil {
		opt.ReconnectionBackoff = defaultBackoff(opt.Clock)
	}
	if opt.MigrationTimeout == 0 {
		opt.MigrationTimeout = time.Second * 15
	}
	if opt.MessageID == nil {
		opt.MessageID = proto.NewMessageIDGen(opt.Clock.Now)
	}
	if opt.UpdateHandler == nil {
		// No updates handler passed, so no sense to subscribe for updates.
		// User should explicitly ignore updates using custom UpdateHandler.
		opt.NoUpdates = true

		// Using no-op handler.
		opt.UpdateHandler = UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
			return nil
		})
	}
	if opt.OnTransfer == nil {
		opt.OnTransfer = noopOnTransfer
	}
}

func defaultBackoff(c clock.Clock) func() backoff.BackOff {
	return func() backoff.BackOff {
		b := backoff.NewExponentialBackOff()
		b.Clock = c
		b.MaxElapsedTime = 0
		b.MaxInterval = time.Second * 5
		b.InitialInterval = time.Millisecond * 100
		return b
	}
}
