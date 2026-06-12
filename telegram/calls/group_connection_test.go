package calls

import (
	"encoding/json"
	"testing"

	"github.com/pion/logging"
	"github.com/pion/rtcp"
	"github.com/pion/transport/v4/vnet"
	"github.com/pion/webrtc/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log"
	"github.com/gotd/log/logzap"
)

// TestGroupConnJoinPayload exercises the SFU-side transport setup that does not
// require a live server: building the PeerConnection, adding the audio track,
// gathering host candidates over a virtual network and producing the join
// payload, plus the connection lifecycle helpers.
func TestGroupConnJoinPayload(t *testing.T) {
	if testing.Short() {
		t.Skip("gathers ICE candidates")
	}

	router, err := vnet.NewRouter(&vnet.RouterConfig{
		CIDR:          "10.0.0.0/24",
		LoggerFactory: logging.NewDefaultLoggerFactory(),
	})
	require.NoError(t, err)
	nw, err := vnet.NewNet(&vnet.NetConfig{StaticIPs: []string{"10.0.0.1"}})
	require.NoError(t, err)
	require.NoError(t, router.AddNet(nw))
	require.NoError(t, router.Start())
	defer func() { _ = router.Stop() }()

	g := newGroupConn(log.For(logzap.New(zaptest.NewLogger(t))))
	g.net = nw
	defer func() { _ = g.close() }()

	require.Equal(t, webrtc.PeerConnectionStateClosed, g.connectionState(), "no pc before open")
	require.Error(t, g.writeRTCP([]rtcp.Packet{&rtcp.SenderReport{}}), "no pc before open")

	require.NoError(t, g.open())
	require.NoError(t, g.addAudioTrack())
	require.NotNil(t, g.audioTrack)
	require.NotZero(t, g.audioSSRC)

	payload, err := g.buildJoinPayload()
	require.NoError(t, err)

	var jp groupJoinPayload
	require.NoError(t, json.Unmarshal([]byte(payload), &jp))
	require.NotEmpty(t, jp.Ufrag)
	require.NotEmpty(t, jp.Pwd)
	require.Len(t, jp.Fingerprints, 1)
	require.Equal(t, "sha-256", jp.Fingerprints[0].Hash)
	require.Equal(t, "passive", jp.Fingerprints[0].Setup)
	require.Equal(t, int32(g.audioSSRC), jp.Ssrc)

	require.NotEqual(t, webrtc.PeerConnectionStateClosed, g.connectionState())
}

func TestGroupConnFireOnce(t *testing.T) {
	g := newGroupConn(log.For(logzap.New(zaptest.NewLogger(t))))

	connected := 0
	g.onConnected = func() { connected++ }
	g.fireConnected()
	g.fireConnected()
	require.Equal(t, 1, connected)

	disconnected := 0
	g.onDisconnected = func() { disconnected++ }
	g.fireDisconnected()
	g.fireDisconnected()
	require.Equal(t, 1, disconnected)
}
