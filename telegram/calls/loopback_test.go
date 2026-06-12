package calls

import (
	"sync"
	"testing"
	"time"

	"github.com/pion/logging"
	"github.com/pion/rtp"
	"github.com/pion/transport/v4/vnet"
	"github.com/pion/webrtc/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log"
	"github.com/gotd/log/logzap"
)

// TestConnLoopback runs a full 1:1 call handshake between two Conn instances
// over an in-memory pion virtual network: ICE gathering, the InitialSetup /
// Candidates / NegotiateChannels signaling exchange, the DTLS handshake and the
// SRTP media path — with no real network and no Telegram server.
func TestConnLoopback(t *testing.T) {
	if testing.Short() {
		t.Skip("loopback handshake is slow")
	}

	// Two virtual NICs on a shared subnet: direct host-candidate connectivity.
	router, err := vnet.NewRouter(&vnet.RouterConfig{
		CIDR:          "10.0.0.0/24",
		LoggerFactory: logging.NewDefaultLoggerFactory(),
	})
	require.NoError(t, err)
	callerNet, err := vnet.NewNet(&vnet.NetConfig{StaticIPs: []string{"10.0.0.1"}})
	require.NoError(t, err)
	calleeNet, err := vnet.NewNet(&vnet.NetConfig{StaticIPs: []string{"10.0.0.2"}})
	require.NoError(t, err)
	require.NoError(t, router.AddNet(callerNet))
	require.NoError(t, router.AddNet(calleeNet))
	require.NoError(t, router.Start())
	defer func() { _ = router.Stop() }()

	zapLog := zaptest.NewLogger(t)
	caller := newConn(true, log.For(logzap.New(zapLog.Named("caller"))))
	caller.net = callerNet
	callee := newConn(false, log.For(logzap.New(zapLog.Named("callee"))))
	callee.net = calleeNet
	defer func() { _ = caller.Close(); _ = callee.Close() }()

	// In-memory signaling bridge: each side's emitted JSON is delivered to the
	// other's onSignal asynchronously (as it would be over the wire).
	bridge := func(from, to *Conn) {
		from.emit = func(payload []byte) {
			data := append([]byte(nil), payload...)
			go func() { _ = to.onSignal(data) }()
		}
	}
	bridge(caller, callee)
	bridge(callee, caller)

	callerUp := make(chan struct{})
	calleeUp := make(chan struct{})
	caller.OnConnected(func() { close(callerUp) })
	callee.OnConnected(func() { close(calleeUp) })

	// The callee reports the caller's forwarded audio track.
	gotRTP := make(chan struct{})
	var once sync.Once
	callee.OnTrack(func(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		if track.Kind() != webrtc.RTPCodecTypeAudio {
			return
		}
		go func() {
			if _, _, err := track.ReadRTP(); err == nil {
				once.Do(func() { close(gotRTP) })
			}
		}()
	})

	require.NoError(t, caller.open(nil))
	require.NoError(t, callee.open(nil))
	require.NoError(t, caller.start()) // caller offers first
	require.NoError(t, callee.start()) // callee waits

	waitClosed(t, callerUp, 30*time.Second, "caller never connected")
	waitClosed(t, calleeUp, 30*time.Second, "callee never connected")

	// Push audio from the caller until the callee receives a packet.
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
		var seq uint16
		var ts uint32
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
			}
			seq++
			ts += 960
			_ = caller.AudioTrack().WriteRTP(&rtp.Packet{
				Header:  rtp.Header{Version: 2, PayloadType: 111, SequenceNumber: seq, Timestamp: ts},
				Payload: make([]byte, 60),
			})
		}
	}()

	waitClosed(t, gotRTP, 30*time.Second, "callee never received audio RTP")
}

func waitClosed(t *testing.T, ch <-chan struct{}, d time.Duration, msg string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(d):
		t.Fatal(msg)
	}
}
