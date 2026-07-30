package calls

import (
	"testing"

	"github.com/gotd/log"
	"github.com/pion/webrtc/v4"
	"github.com/stretchr/testify/require"
)

func TestConnCallbacksAndState(t *testing.T) {
	c := newConn(true, log.For(log.Nop))
	require.Equal(t, "new", c.State())

	c.addTracks()
	require.NotNil(t, c.AudioTrack())
	require.NotNil(t, c.VideoTrack())
	require.NotZero(t, c.AudioSSRC())
	require.Equal(t, c.AudioSSRC()+1, c.VideoSSRC())

	connected := 0
	c.OnConnected(func() { connected++ })
	c.OnDisconnected(func() {})
	c.OnStateChange(func(string) {})
	require.Equal(t, 0, connected, "OnConnected must not fire before connection")

	// Simulate ICE+DTLS connected.
	c.mu.Lock()
	c.iceConnected = true
	c.dtlsConnected = true
	c.mu.Unlock()
	c.updateState()
	require.Equal(t, 1, connected)
	require.Equal(t, "connected", c.State())

	// Setting OnConnected after the fact fires immediately.
	fired := false
	c.OnConnected(func() { fired = true })
	require.True(t, fired)
}

func TestConnFireDisconnectedOnce(t *testing.T) {
	c := newConn(false, log.For(log.Nop))
	n := 0
	c.OnDisconnected(func() { n++ })
	c.fireDisconnected()
	c.fireDisconnected()
	require.Equal(t, 1, n)
	require.Equal(t, "closed", c.State())
	require.NoError(t, c.Close()) // nil pc
}

// TestConnOnTrackReplaysEarlyTracks covers the callee's normal ordering: the
// peer's media is negotiated before the application, which only gets the Conn
// once Accept returns, can register OnTrack. Dropping those tracks left the
// callee hearing nothing for the entire call.
func TestConnOnTrackReplaysEarlyTracks(t *testing.T) {
	c := newConn(false, log.For(log.Nop))

	// Two tracks arrive with no callback registered yet.
	c.deliverTrack(nil, nil)
	c.deliverTrack(nil, nil)

	n := 0
	c.OnTrack(func(*webrtc.TrackRemote, *webrtc.RTPReceiver) { n++ })
	require.Equal(t, 2, n, "tracks that arrived early must be replayed")

	// Nothing is replayed twice, and later tracks go straight through.
	c.OnTrack(func(*webrtc.TrackRemote, *webrtc.RTPReceiver) { n++ })
	require.Equal(t, 2, n)
	c.deliverTrack(nil, nil)
	require.Equal(t, 3, n)
}

func TestConnEmitJSON(t *testing.T) {
	c := newConn(true, log.For(log.Nop))
	var sent []byte
	c.emit = func(b []byte) { sent = b }
	c.emitJSON(map[string]string{"@type": "X"})
	require.Contains(t, string(sent), "@type")
}

func TestConnOnSignal(t *testing.T) {
	c := newConn(true, log.For(log.Nop))

	require.NoError(t, c.onSignal([]byte(`{"@type":"MediaState","videoState":"active"}`)))
	require.NoError(t, c.onSignal([]byte(`{"@type":"SomethingUnknown"}`)))
	require.Error(t, c.onSignal([]byte(`not json`)))
}

// TestConnHandleNegotiate exercises the negotiation path without a live
// transport: with DTLS not yet started, maybeCreateChannels returns early, so
// no pion transport is needed.
func TestConnHandleNegotiate(t *testing.T) {
	c := newConn(true, log.For(log.Nop))
	c.neg = newContentNegotiation()
	c.addTracks()

	var emitted [][]byte
	c.emit = func(b []byte) { emitted = append(emitted, b) }

	msg, err := jsonMarshal(negotiateChannelsMessage{
		Type:       typeNegotiateChannels,
		ExchangeID: "999",
		Contents:   []mediaContent{{Type: "audio", Ssrc: "5000"}},
	})
	require.NoError(t, err)

	require.NoError(t, c.onSignal(msg))
	require.True(t, c.negotiated)
	require.Equal(t, uint32(5000), c.neg.peerAudioSSRC())
	require.Len(t, emitted, 1, "expected an answer to be emitted")

	// A second negotiate is ignored once negotiated.
	require.NoError(t, c.onSignal(msg))
	require.Len(t, emitted, 1)
}
