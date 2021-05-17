package telegram

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/internal/mtproto"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/middleware"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
)

// Options of Client.
type Options struct {
	// PublicKeys of telegram.
	//
	// If not provided, embedded public keys will be used.
	PublicKeys []*rsa.PublicKey

	// DC ID to connect.
	//
	// If not provided, 2 will be used by default.
	DC int

	// DCList is initial list of addresses to connect.
	DCList dcs.DCList

	// Resolver to use.
	Resolver dcs.Resolver
	// NoUpdates enables no updates mode.
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

	// Middleware to use for RPC calls.
	//
	// If not provided, no middleware is used.
	Middleware middleware.Middleware

	// AckBatchSize is maximum ack-s to buffer.
	AckBatchSize int
	// AckInterval is maximum time to buffer ack.
	AckInterval   time.Duration
	RetryInterval time.Duration
	MaxRetries    int

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
		opt.Random = rand.Reader
	}
	if opt.Logger == nil {
		opt.Logger = zap.NewNop()
	}
	if opt.DC == 0 {
		opt.DC = 2
	}
	if opt.DCList.Zero() {
		opt.DCList = dcs.ProdDCs()
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
	if opt.MigrationTimeout == 0 {
		opt.MigrationTimeout = time.Second * 15
	}
	if opt.MaxRetries == 0 {
		opt.MaxRetries = 5
	}
	opt.Device.SetDefaults()
	if opt.Clock == nil {
		opt.Clock = clock.System
	}
	if opt.ReconnectionBackoff == nil {
		opt.ReconnectionBackoff = defaultBackoff(opt.Clock)
	}
	if opt.MessageID == nil {
		opt.MessageID = proto.NewMessageIDGen(opt.Clock.Now)
	}
	if opt.UpdateHandler == nil {
		// opt.NoUpdates = true
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
