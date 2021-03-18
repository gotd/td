package telegram

import (
	"crypto/rand"
	"crypto/rsa"
	"io"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/internal/mtproto"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
)

// Transport is MTProto connection creator.
type Transport = dcs.Transport

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
	DCList []tg.DCOption

	// Resolver to use.
	Resolver dcs.Resolver
	// NoUpdates enables no updates mode.
	NoUpdates bool
	// ReconnectionBackoff configures and returns reconnection backoff object.
	ReconnectionBackoff func() backoff.BackOff
	// Random is random source. Defaults to crypto.
	Random io.Reader
	// Logger is instance of zap.Logger. No logs by default.
	Logger *zap.Logger
	// SessionStorage will be used to load and save session data.
	// NB: Very sensitive data, save with care.
	SessionStorage SessionStorage
	// UpdateHandler will be called on received update.
	UpdateHandler UpdateHandler

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
		opt.Resolver = dcs.PlainResolver(dcs.PlainOptions{})
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
	if len(opt.DCList) == 0 {
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
		opt.MessageID = proto.NewMessageIDGen(opt.Clock.Now, 100)
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
