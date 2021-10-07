package telegram

import (
	"context"
	"io"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"

	"github.com/nnqq/td/clock"
	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/internal/exchange"
	"github.com/nnqq/td/internal/mtproto"
	"github.com/nnqq/td/internal/proto"
	"github.com/nnqq/td/telegram/dcs"
	"github.com/nnqq/td/tg"
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

	// ReconnectionBackoff configures and returns reconnection backoff object.
	ReconnectionBackoff func() backoff.BackOff
	// MigrationTimeout configures migration timeout.
	MigrationTimeout time.Duration

	// Random is random source. Defaults to crypto.
	Random io.Reader
	// Logger is instance of zap.Logger. No logs by default.
	Logger *zap.Logger
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

	// CompressThreshold is a threshold in bytes to determine that message
	// is large enough to be compressed using GZIP.
	// If < 0, compression will be disabled.
	// If == 0, default value will be used.
	CompressThreshold int

	// Device is device config.
	// Will be sent with session creation request.
	Device DeviceConfig

	MessageID mtproto.MessageIDSource
	Clock     clock.Clock
}

func (opt *Options) setDefaults() {
	if opt.Resolver == nil {
		opt.Resolver = dcs.DefaultResolver()
	}
	if opt.Random == nil {
		opt.Random = crypto.DefaultRand()
	}
	if opt.Logger == nil {
		opt.Logger = zap.NewNop()
	}
	if opt.DC == 0 {
		opt.DC = 2
	}
	if opt.DCList.Zero() {
		opt.DCList = dcs.Prod()
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
}

func defaultBackoff(c clock.Clock) func() backoff.BackOff {
	return func() backoff.BackOff {
		b := backoff.NewExponentialBackOff()
		b.Clock = c
		b.MaxElapsedTime = 0
		return b
	}
}
