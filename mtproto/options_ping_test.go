package mtproto

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/proto"
)

func TestOptionsPingDefaults(t *testing.T) {
	var opt Options
	opt.setDefaults()

	require.Equal(t, time.Minute, opt.PingInterval)
	require.Equal(t, 15*time.Second, opt.PingTimeout)
	require.Equal(t, 75*time.Second, opt.PingDelayDisconnect,
		"default must stay PingInterval+PingTimeout for wire compatibility")
	require.Equal(t, opt.PingDelayDisconnect, opt.IdleTimeout)
}

func TestOptionsPingExplicit(t *testing.T) {
	// tdesktop profile.
	opt := Options{
		PingInterval:        30 * time.Second,
		PingTimeout:         15 * time.Second,
		PingDelayDisconnect: 60 * time.Second,
	}
	opt.setDefaults()

	require.Equal(t, 60*time.Second, opt.PingDelayDisconnect)
	require.Equal(t, 60*time.Second, opt.IdleTimeout)
}

func TestOptionsPingDelayMustExceedInterval(t *testing.T) {
	// A delay below the interval makes the server drop us between pings,
	// producing an endless reconnect loop. Such config is corrected.
	opt := Options{
		PingInterval:        60 * time.Second,
		PingTimeout:         15 * time.Second,
		PingDelayDisconnect: 10 * time.Second,
	}
	opt.setDefaults()

	require.Greater(t, opt.PingDelayDisconnect, opt.PingInterval)
}

func TestOptionsPingDelayWireTruncation(t *testing.T) {
	// DisconnectDelay is sent to the server as int(seconds), which floors.
	// 60500ms truncates to 60s on the wire, exactly equal to a 60s
	// PingInterval: a naive Duration comparison (60.5s > 60s) would pass this
	// through unchanged and ship a server window equal to our ping period.
	opt := Options{
		PingInterval:        60 * time.Second,
		PingDelayDisconnect: 60500 * time.Millisecond,
	}
	opt.setDefaults()

	wireDelay := time.Duration(int(opt.PingDelayDisconnect.Seconds())) * time.Second
	require.Greater(t, wireDelay, opt.PingInterval,
		"corrected PingDelayDisconnect must clear PingInterval after truncation to whole seconds")
}

func TestOptionsPingDelaySubSecondFallback(t *testing.T) {
	// disconnect_delay goes on the wire as an int of seconds, so with
	// sub-second ping settings PingInterval+PingTimeout floors to zero.
	// Re-applying that same sum as the correction would ship
	// disconnect_delay=0, asking the server to drop the connection at once —
	// precisely the outcome the validation exists to prevent.
	for _, tt := range []struct {
		name string
		opt  Options
	}{
		{
			name: "derived default",
			opt:  Options{PingInterval: 500 * time.Millisecond, PingTimeout: 100 * time.Millisecond},
		},
		{
			name: "explicitly configured below the floor",
			opt: Options{
				PingInterval:        500 * time.Millisecond,
				PingTimeout:         100 * time.Millisecond,
				PingDelayDisconnect: 600 * time.Millisecond,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			opt := tt.opt
			opt.setDefaults()

			require.NotZero(t, int(opt.PingDelayDisconnect.Seconds()),
				"disconnect_delay must never reach the server as 0")

			wireDelay := time.Duration(int(opt.PingDelayDisconnect.Seconds())) * time.Second
			require.Greater(t, wireDelay, opt.PingInterval,
				"corrected PingDelayDisconnect must clear PingInterval after truncation to whole seconds")
		})
	}
}

func TestOptionsIdleTimeoutMustExceedInterval(t *testing.T) {
	// An IdleTimeout no longer than PingInterval would tear down a healthy
	// connection between pings.
	opt := Options{
		PingInterval: 60 * time.Second,
		PingTimeout:  15 * time.Second,
		IdleTimeout:  10 * time.Second,
	}
	opt.setDefaults()

	require.Greater(t, opt.IdleTimeout, opt.PingInterval)
	require.Equal(t, opt.PingDelayDisconnect, opt.IdleTimeout,
		"corrected IdleTimeout must fall back to the derived default")
}

// TestOptionsIdleTimeoutMustCoverPingTimeout covers the gap left by merely
// requiring IdleTimeout > PingInterval: 61s clears a 60s ping period but
// leaves only a 1s margin, so a pong arriving 1.1s late — still well inside
// the 15s PingTimeout — trips the watchdog and closes a healthy connection.
// The floor must be a full ping period plus its pong wait.
func TestOptionsIdleTimeoutMustCoverPingTimeout(t *testing.T) {
	opt := Options{
		PingInterval: 60 * time.Second,
		PingTimeout:  15 * time.Second,
		IdleTimeout:  61 * time.Second,
	}
	opt.setDefaults()

	require.GreaterOrEqual(t, opt.IdleTimeout, opt.PingInterval+opt.PingTimeout,
		"corrected IdleTimeout must survive a merely-late pong")
	require.Equal(t, 75*time.Second, opt.IdleTimeout)
}

// TestOptionsIdleTimeoutClampsLowDisconnectDelay covers the fallback itself:
// PingDelayDisconnect is validated only against PingInterval, so it can sit
// below the IdleTimeout floor. Falling back to it blindly would reintroduce
// the very margin this validation exists to prevent.
func TestOptionsIdleTimeoutClampsLowDisconnectDelay(t *testing.T) {
	opt := Options{
		PingInterval:        60 * time.Second,
		PingTimeout:         15 * time.Second,
		PingDelayDisconnect: 61 * time.Second,
	}
	opt.setDefaults()

	require.Equal(t, 61*time.Second, opt.PingDelayDisconnect,
		"PingDelayDisconnect itself must be left alone: it clears PingInterval on the wire")
	require.GreaterOrEqual(t, opt.IdleTimeout, opt.PingInterval+opt.PingTimeout)
}

// identityCipher writes the plaintext service message straight to the wire
// buffer instead of encrypting it, so a test can decode exactly what would
// have been sent to the server.
type identityCipher struct{}

func (identityCipher) Encrypt(_ crypto.AuthKey, d crypto.EncryptedMessageData, b *bin.Buffer) error {
	return d.Message.Encode(b)
}

func (identityCipher) DecryptFromBuffer(crypto.AuthKey, *bin.Buffer) (*crypto.EncryptedMessageData, error) {
	return nil, errors.New("identityCipher: decrypt not implemented")
}

// captureSendConn records the first buffer handed to Send and then fails the
// write, so pingLoop does not block forever waiting for a pong that will
// never arrive.
type captureSendConn struct {
	sent chan []byte
}

func (c *captureSendConn) Send(_ context.Context, b *bin.Buffer) error {
	buf := make([]byte, len(b.Buf))
	copy(buf, b.Buf)
	select {
	case c.sent <- buf:
	default:
	}
	return context.Canceled
}

func (c *captureSendConn) Recv(ctx context.Context, _ *bin.Buffer) error {
	<-ctx.Done()
	return ctx.Err()
}

func (c *captureSendConn) Close() error { return nil }

// TestPingLoopSendsConfiguredDisconnectDelay pins the wire value: it must
// track Options.PingDelayDisconnect (default or explicit), not be
// recomputed from PingInterval+PingTimeout inside pingLoop. A test that only
// checks setDefaults would not catch a regression where pingLoop reverts to
// the old formula while the option itself goes unused.
func TestPingLoopSendsConfiguredDisconnectDelay(t *testing.T) {
	tests := []struct {
		name string
		opt  Options
		want int
	}{
		{
			name: "default",
			opt:  Options{},
			want: 75,
		},
		{
			name: "explicit",
			opt: Options{
				PingInterval:        30 * time.Second,
				PingTimeout:         15 * time.Second,
				PingDelayDisconnect: 60 * time.Second,
			},
			want: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := tt.opt
			opt.setDefaults()
			// Fire the ping ticker almost immediately; only
			// PingDelayDisconnect (not PingInterval) is under test here.
			opt.PingInterval = time.Millisecond

			sent := make(chan []byte, 1)
			c := &Conn{
				conn:            &captureSendConn{sent: sent},
				clock:           clock.System,
				rand:            rand.Reader,
				cipher:          identityCipher{},
				log:             log.For(log.Nop),
				messageID:       proto.NewMessageIDGen(clock.System.Now),
				ping:            map[int64]chan struct{}{},
				pingTimeout:     opt.PingTimeout,
				pingInterval:    opt.PingInterval,
				disconnectDelay: opt.PingDelayDisconnect,
			}

			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			// Returns as soon as captureSendConn rejects the write.
			_ = c.pingLoop(ctx)

			select {
			case buf := <-sent:
				var req mt.PingDelayDisconnectRequest
				require.NoError(t, req.Decode(&bin.Buffer{Buf: buf}))
				require.Equal(t, tt.want, req.DisconnectDelay)
			default:
				t.Fatal("ping_delay_disconnect was not sent")
			}
		})
	}
}
