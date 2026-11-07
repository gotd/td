package telegram

import (
	"context"
	"io"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.opentelemetry.io/otel/trace"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
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
	//
	// It is called for all connections: primary, pools and CDN, without
	// telling them apart. Prefer OnConnDead in new code.
	OnDead func(error)
	// OnConnDead is called on connection dead with the DC ID and the kind of
	// the connection, e.g. for metrics. Preferred over OnDead, which is kept
	// for compatibility; if both are set, OnDead is called first.
	//
	// Called from the connection lifecycle, so the callback must not block.
	OnConnDead func(ConnDeadEvent)
	// OnConnectionState is called when the primary connection state changes:
	// on every (re)connect start, on successful initialization and on
	// connection loss. Connection death details are reported via OnDead.
	//
	// Called synchronously from connection lifecycle, so the callback must
	// not block.
	OnConnectionState func(ConnectionState)
	// OnConnectionStateEvent is OnConnectionState with the DC ID of the
	// primary connection the state is about, which changes on migration.
	// Preferred over OnConnectionState, which is kept for compatibility; if
	// both are set, OnConnectionState is called first. Same constraints.
	OnConnectionStateEvent func(ConnectionStateEvent)
	// MigrationTimeout configures migration timeout.
	MigrationTimeout time.Duration
	// FilePoolSize is the max number of connections of each pool returned by
	// Client.FileDC. Defaults to 4.
	FilePoolSize int64

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
	// PingInterval is duration between pings (ping_delay_disconnect) of a
	// connection. Defaults to 1 minute, see mtproto.Options.
	PingInterval time.Duration
	// PingTimeout is the timeout of ping response, after which the connection
	// is considered dead. Defaults to 15 seconds, see mtproto.Options.
	PingTimeout time.Duration
	// RequestTimeout returns response deadline for a request on the primary
	// connection, zero to wait as long as the request context allows
	// (optional).
	//
	// If the deadline passes while the request context is still alive (e.g.
	// the request was acknowledged but never answered by a half-dead
	// connection), the client reconnects and sends the request once more on
	// the new connection, like Telegram Desktop does. Return zero for
	// requests that must not be sent twice.
	RequestTimeout func(input bin.Encoder) time.Duration
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

	// OnAuthLost is called when a request of an authorized client fails
	// with an error meaning that the authorization is lost (see IsAuthLost),
	// e.g. the session was terminated from another device. Called once until
	// the client is authorized again.
	//
	// The client counts as authorized after any successful request that
	// requires authorization (anything but auth.*, help.*, langpack.* and the
	// 2FA/login email account.* methods), including Self() on connect, or
	// after a sign in that returned a complete authorization. So a session
	// restored from storage is covered after its first such request, also
	// with NoUpdates.
	//
	// ctx is the context of Run, not of the failed request, so the callback
	// can make requests, e.g. to sign in again.
	//
	// Return non-nil error to stop Run with *AuthLostError, so that one
	// account of many can be stopped and reported. Return nil to keep
	// running, e.g. to sign in again.
	//
	// If the connect-time Self() of an authorized client reveals the lost
	// authorization (e.g. on reconnect), OnAuthLost is called first. If it
	// returns nil, OnSelfError is called too, as for any Self() error; if
	// it stops Run, OnSelfError is not called. Before the client is
	// authorized, such errors are reported via OnSelfError only.
	OnAuthLost func(ctx context.Context, err error) error
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
