package calls

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/go-faster/errors"
	"github.com/pion/interceptor"
	"github.com/pion/rtcp"
	"github.com/pion/transport/v4"
	"github.com/pion/webrtc/v4"

	"github.com/gotd/log"
)

// groupConn is the media transport for a group call: a single pion
// PeerConnection that performs SDP offer/answer with the Telegram SFU.
type groupConn struct {
	log log.Helper

	mu          sync.Mutex
	pc          *webrtc.PeerConnection
	audioTrack  *webrtc.TrackLocalStaticRTP
	audioSender *webrtc.RTPSender
	audioSSRC   uint32
	state       string

	srflxReady chan struct{}
	srflxOnce  sync.Once

	onConnected     func()
	onDisconnected  func()
	onStateChange   func(string)
	onTrack         func(*webrtc.TrackRemote, *webrtc.RTPReceiver)
	firedConnect    bool
	firedDisconnect bool

	// net, when set, routes the PeerConnection over a custom transport (used by
	// tests); nil in production.
	net transport.Net
}

func newGroupConn(logger log.Helper) *groupConn { return &groupConn{log: logger} }

// open builds the PeerConnection and installs its state handlers.
func (g *groupConn) open() error {
	g.audioSSRC = randomSSRC()
	g.srflxReady = make(chan struct{})

	me, err := buildMediaEngine()
	if err != nil {
		return err
	}
	ir := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(me, ir); err != nil {
		return errors.Wrap(err, "register interceptors")
	}
	se := buildSettingEngine()
	cfg := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
			{URLs: []string{"stun:stun1.l.google.com:19302"}},
		},
	}
	if g.net != nil {
		se.SetNet(g.net)
		cfg.ICEServers = nil
	}
	api := webrtc.NewAPI(
		webrtc.WithMediaEngine(me),
		webrtc.WithInterceptorRegistry(ir),
		webrtc.WithSettingEngine(se),
	)

	pc, err := api.NewPeerConnection(cfg)
	if err != nil {
		return errors.Wrap(err, "create peer connection")
	}
	g.pc = pc

	pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		g.log.Debug(context.Background(), "Group call connection state", log.Stringer("state", s))
		g.mu.Lock()
		g.state = s.String()
		onState := g.onStateChange
		g.mu.Unlock()
		if onState != nil {
			onState(s.String())
		}
		switch s {
		case webrtc.PeerConnectionStateConnected:
			g.fireConnected()
		case webrtc.PeerConnectionStateFailed, webrtc.PeerConnectionStateClosed,
			webrtc.PeerConnectionStateDisconnected:
			g.fireDisconnected()
		}
	})
	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		// The first server-reflexive candidate tells us our public address is
		// known; the offer is most useful to the SFU once we have it.
		if c.Typ == webrtc.ICECandidateTypeSrflx {
			g.srflxOnce.Do(func() { close(g.srflxReady) })
		}
	})
	pc.OnTrack(func(t *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		if g.onTrack != nil {
			g.onTrack(t, r)
		}
	})
	return nil
}

// addAudioTrack adds the outgoing audio track with the pinned SSRC.
func (g *groupConn) addAudioTrack() error {
	track, err := webrtc.NewTrackLocalStaticRTP(audioCodec(), "audio", "gotd-audio")
	if err != nil {
		return errors.Wrap(err, "create audio track")
	}
	tr, err := g.pc.AddTransceiverFromTrack(track, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendrecv,
		SendEncodings: []webrtc.RTPEncodingParameters{{
			RTPCodingParameters: webrtc.RTPCodingParameters{SSRC: webrtc.SSRC(g.audioSSRC)},
		}},
	})
	if err != nil {
		return errors.Wrap(err, "add audio transceiver")
	}
	g.audioTrack = track
	g.audioSender = tr.Sender()
	return nil
}

// buildJoinPayload creates the SDP offer, gathers candidates and returns the
// JSON join payload describing our transport and audio SSRC.
func (g *groupConn) buildJoinPayload() (string, error) {
	offer, err := g.pc.CreateOffer(nil)
	if err != nil {
		return "", errors.Wrap(err, "create offer")
	}
	if err := g.pc.SetLocalDescription(offer); err != nil {
		return "", errors.Wrap(err, "set local description")
	}

	// Prefer to send the offer once we have a server-reflexive candidate, but
	// don't block forever if STUN is slow/unavailable.
	select {
	case <-g.srflxReady:
	case <-time.After(3 * time.Second):
		g.log.Warn(context.Background(), "No server-reflexive candidate gathered; sending host-only offer")
	}

	ufrag, pwd, fingerprint, hash := extractSDPParams(g.pc.LocalDescription().SDP)

	// Pin the advertised SSRC to whatever pion actually bound to the sender.
	if g.audioSender != nil {
		if enc := g.audioSender.GetParameters().Encodings; len(enc) > 0 && enc[0].SSRC != 0 {
			g.audioSSRC = uint32(enc[0].SSRC)
		}
	}

	data, err := json.Marshal(groupJoinPayload{
		Ufrag:        ufrag,
		Pwd:          pwd,
		Fingerprints: []groupFingerprint{{Hash: hash, Setup: "passive", Fingerprint: fingerprint}},
		Ssrc:         int32(g.audioSSRC),
	})
	if err != nil {
		return "", errors.Wrap(err, "marshal join payload")
	}
	return string(data), nil
}

// connect applies the SFU's response as the SDP answer.
func (g *groupConn) connect(responseJSON string) error {
	var resp groupJoinResponse
	if err := json.Unmarshal([]byte(responseJSON), &resp); err != nil {
		return errors.Wrap(err, "parse server response")
	}
	if st := g.pc.SignalingState(); st != webrtc.SignalingStateHaveLocalOffer {
		return errors.Errorf("signaling state is %s, expected have-local-offer", st)
	}
	if err := g.pc.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  buildAnswerSDP(resp),
	}); err != nil {
		return errors.Wrap(err, "set remote description")
	}
	return nil
}

func (g *groupConn) fireConnected() {
	g.mu.Lock()
	first := !g.firedConnect
	g.firedConnect = true
	fn := g.onConnected
	g.mu.Unlock()
	if first && fn != nil {
		fn()
	}
}

func (g *groupConn) fireDisconnected() {
	g.mu.Lock()
	if g.firedDisconnect {
		g.mu.Unlock()
		return
	}
	g.firedDisconnect = true
	fn := g.onDisconnected
	g.mu.Unlock()
	if fn != nil {
		fn()
	}
}

func (g *groupConn) writeRTCP(pkts []rtcp.Packet) error {
	g.mu.Lock()
	pc := g.pc
	g.mu.Unlock()
	if pc == nil {
		return errors.New("connection closed")
	}
	return pc.WriteRTCP(pkts)
}

func (g *groupConn) connectionState() webrtc.PeerConnectionState {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.pc == nil {
		return webrtc.PeerConnectionStateClosed
	}
	return g.pc.ConnectionState()
}

func (g *groupConn) close() error {
	g.mu.Lock()
	pc := g.pc
	g.mu.Unlock()
	if pc != nil {
		return pc.Close()
	}
	return nil
}
