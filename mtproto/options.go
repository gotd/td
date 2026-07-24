package mtproto

import (
	"context"
	"io"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/exchange"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/rpc"
	"github.com/gotd/td/tmap"
)

// Options of Conn.
type Options struct {
	// DC is datacenter ID for key exchange.
	// Defaults to 2.
	DC int

	// PublicKeys of telegram.
	//
	// If not provided, embedded public keys will be used.
	PublicKeys []exchange.PublicKey

	// Random is random source. Defaults to crypto.
	Random io.Reader
	// Logger is the structured logger. No logs by default.
	Logger log.Logger
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

	// DialTimeout is timeout of creating connection.
	DialTimeout time.Duration
	// ExchangeTimeout is timeout of every key exchange request.
	ExchangeTimeout time.Duration
	// SaltFetchInterval is duration between get_future_salts request.
	SaltFetchInterval time.Duration
	// PingTimeout is how long to wait for a pong before considering the
	// ping failed.
	PingTimeout time.Duration
	// PingInterval is duration between ping_delay_disconnect request.
	PingInterval time.Duration
	// PingDelayDisconnect is disconnect_delay value sent to the server in
	// ping_delay_disconnect: the server drops the connection if it receives no
	// ping within this duration. Must be greater than PingInterval.
	//
	// Defaults to PingInterval + PingTimeout.
	PingDelayDisconnect time.Duration
	// IdleTimeout is the maximum duration without any received data before the
	// connection is considered dead and closed.
	//
	// Defaults to PingDelayDisconnect.
	IdleTimeout time.Duration
	// RequestTimeout is function which returns request timeout for given type ID.
	RequestTimeout func(req uint32) time.Duration

	// CompressThreshold is a threshold in bytes to determine that message
	// is large enough to be compressed using GZIP.
	// If < 0, compression will be disabled.
	// If == 0, default value will be used.
	CompressThreshold int
	// MessageID is message id source. Share source between connection to
	// reduce collision probability.
	MessageID MessageIDSource
	// Clock is current time source. Defaults to system time.
	Clock clock.Clock
	// Types map, used in verbose logging of incoming message.
	Types *tmap.Map
	// Key that can be used to restore previous connection.
	Key crypto.AuthKey
	// PermKey is permanent auth key for PFS mode.
	PermKey crypto.AuthKey
	// Salt from server that can be used to restore previous connection.
	Salt int64

	// EnablePFS enables Perfect Forward Secrecy using temporary auth keys.
	EnablePFS bool
	// TempKeyTTL is temporary auth key lifetime in seconds.
	// Default: 86400 (24h). Minimum: 60.
	TempKeyTTL int

	// Tracer for OTEL.
	Tracer trace.Tracer

	// Private options.

	// Cipher defines message crypto.
	Cipher Cipher
	// engine for replacing RPC engine.
	engine *rpc.Engine
}

const (
	// Telegram recommends short-lived temp keys; we keep 24h as practical
	// default to balance reconnect frequency and security.
	defaultTempKeyTTL = 24 * 60 * 60
	// Prevent obviously invalid ttl values that can break renewal math.
	minTempKeyTTL = 60
)

type nopHandler struct{}

func (nopHandler) OnMessage(b *bin.Buffer) error   { return nil }
func (nopHandler) OnSession(session Session) error { return nil }

func (opt *Options) setDefaultPublicKeys() {
	// Using public keys that are included with distribution if not
	// provided.
	//
	// This should never fail and keys should be valid for recent
	// library versions.
	opt.PublicKeys = vendoredKeys()
}

// defaultPingDelayDisconnect derives a disconnect_delay that stays meaningful
// after pingLoop truncates it to whole seconds for the wire.
//
// PingInterval + PingTimeout is the intended value, but it is a Duration while
// ping_delay_disconnect.disconnect_delay is an int of seconds: with sub-second
// ping settings that sum floors to zero, asking the server to drop the
// connection immediately — the exact failure the validation in setDefaults
// exists to prevent, which a plain re-assignment of the same sum would not fix.
// The first whole second strictly above PingInterval is therefore used as a
// floor. With the defaults (60s + 15s) the sum already clears it, so the wire
// value is unchanged at 75s.
func defaultPingDelayDisconnect(interval, timeout time.Duration) time.Duration {
	minWireDelay := (interval/time.Second + 1) * time.Second
	return max(interval+timeout, minWireDelay)
}

func (opt *Options) setDefaults() {
	if opt.DC == 0 {
		opt.DC = 2
	}
	if opt.Random == nil {
		opt.Random = crypto.DefaultRand()
	}
	if opt.Logger == nil {
		opt.Logger = log.Nop
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
	if opt.MaxRetries == 0 {
		opt.MaxRetries = 5
	}
	if opt.DialTimeout == 0 {
		opt.DialTimeout = 35 * time.Second
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
	if opt.PingDelayDisconnect == 0 {
		opt.PingDelayDisconnect = defaultPingDelayDisconnect(opt.PingInterval, opt.PingTimeout)
	}
	// pingLoop truncates PingDelayDisconnect to whole seconds before putting it
	// on the wire (ping_delay_disconnect.disconnect_delay is an int). Validate
	// that truncated value, not the raw Duration: a delay that only clears
	// PingInterval before truncation (e.g. 60.5s delay vs 60s interval, both
	// truncating to 60) still leaves a server window equal to our ping period,
	// which is exactly the endless-reconnect config this check exists to
	// prevent.
	if wireDelay := time.Duration(int(opt.PingDelayDisconnect.Seconds())) * time.Second; wireDelay <= opt.PingInterval {
		// There is no error channel here (New does not return an error), so
		// the value is corrected rather than rejected.
		log.For(opt.Logger).Warn(context.Background(), "PingDelayDisconnect must exceed PingInterval after truncation to whole seconds, using default",
			log.Duration("configured", opt.PingDelayDisconnect),
			log.Duration("configured_wire", wireDelay),
			log.Duration("ping_interval", opt.PingInterval),
		)
		opt.PingDelayDisconnect = defaultPingDelayDisconnect(opt.PingInterval, opt.PingTimeout)
	}
	// The watchdog may only fire once a ping has demonstrably gone unanswered.
	// Merely clearing PingInterval is not enough: with PingInterval=60s and
	// IdleTimeout=61s, a pong arriving 1.1s late — well inside the 15s
	// PingTimeout the connection is otherwise judged healthy by — trips the
	// watchdog and closes a working connection, yielding a self-inflicted
	// disconnect loop. The floor is therefore a full ping period plus the time
	// we are willing to wait for its pong, which is exactly the derived
	// default.
	minIdleTimeout := opt.PingInterval + opt.PingTimeout
	// PingDelayDisconnect can itself be configured below that floor (it is only
	// validated against PingInterval), so clamp rather than trust it blindly.
	defaultIdleTimeout := max(opt.PingDelayDisconnect, minIdleTimeout)
	switch {
	case opt.IdleTimeout == 0:
		opt.IdleTimeout = defaultIdleTimeout
	case opt.IdleTimeout < minIdleTimeout:
		// As with PingDelayDisconnect above, there is no error channel here
		// (New does not return an error), so the value is corrected rather
		// than rejected.
		log.For(opt.Logger).Warn(context.Background(), "IdleTimeout must be at least PingInterval+PingTimeout, using default",
			log.Duration("configured", opt.IdleTimeout),
			log.Duration("minimum", minIdleTimeout),
			log.Duration("ping_interval", opt.PingInterval),
			log.Duration("ping_timeout", opt.PingTimeout),
		)
		opt.IdleTimeout = defaultIdleTimeout
	}
	if opt.RequestTimeout == nil {
		opt.RequestTimeout = func(req uint32) time.Duration {
			return 15 * time.Second
		}
	}
	if opt.CompressThreshold == 0 {
		opt.CompressThreshold = 1024
	}
	if opt.Clock == nil {
		opt.Clock = clock.System
	}
	if opt.MessageID == nil {
		opt.MessageID = proto.NewMessageIDGen(opt.Clock.Now)
	}
	if opt.TempKeyTTL == 0 {
		opt.TempKeyTTL = defaultTempKeyTTL
	}
	if opt.TempKeyTTL < minTempKeyTTL {
		opt.TempKeyTTL = minTempKeyTTL
	}
	// Fallback for callers that still pass restored key via Key.
	if opt.EnablePFS && opt.PermKey.Zero() && !opt.Key.Zero() {
		opt.PermKey = opt.Key
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
