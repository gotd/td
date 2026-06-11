package calls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/go-faster/errors"
	pion "github.com/pion/ice/v4"
	"github.com/pion/interceptor"
	"github.com/pion/transport/v4"
	"github.com/pion/webrtc/v4"
	"go.uber.org/zap"

	"github.com/gotd/td/tg"
)

// Conn is the media transport of an established call. It bridges the Telegram
// signaling channel to a pion/webrtc ICE/DTLS/SRTP connection.
//
// Outgoing media is written to the tracks returned by AudioTrack/VideoTrack;
// incoming media is delivered to the OnTrack callback.
type Conn struct {
	log        *zap.Logger
	isOutgoing bool

	// emit sends an encoded JSON signaling message to the peer.
	emit func(payload []byte)

	// net, when set, routes ICE/DTLS over a custom transport instead of the
	// host network and disables the public STUN fallback. Used by tests to run
	// a full handshake over a pion virtual network; nil in production.
	net transport.Net

	api      *webrtc.API
	gatherer *webrtc.ICEGatherer
	ice      *webrtc.ICETransport
	dtls     *webrtc.DTLSTransport

	audioTrack *webrtc.TrackLocalStaticRTP
	videoTrack *webrtc.TrackLocalStaticRTP
	audioSSRC  uint32
	videoSSRC  uint32

	mu              sync.Mutex
	neg             *contentNegotiation
	remoteReady     bool
	dtlsStarted     bool
	negotiated      bool
	channelsCreated bool
	pendingRemote   []candidateDescription
	iceConnected    bool
	dtlsConnected   bool
	firedConnect    bool
	firedDisconnect bool
	state           string

	onConnected    func()
	onDisconnected func()
	onStateChange  func(state string)
	onTrack        func(*webrtc.TrackRemote, *webrtc.RTPReceiver)
}

func newConn(isOutgoing bool, log *zap.Logger) *Conn {
	return &Conn{isOutgoing: isOutgoing, log: log}
}

// AudioTrack returns the local audio track; write Opus RTP packets to it.
func (c *Conn) AudioTrack() *webrtc.TrackLocalStaticRTP { return c.audioTrack }

// VideoTrack returns the local video track; write VP8 RTP packets to it.
func (c *Conn) VideoTrack() *webrtc.TrackLocalStaticRTP { return c.videoTrack }

// AudioSSRC returns the SSRC chosen for the local audio track.
func (c *Conn) AudioSSRC() uint32 { return c.audioSSRC }

// VideoSSRC returns the SSRC chosen for the local video track.
func (c *Conn) VideoSSRC() uint32 { return c.videoSSRC }

// OnTrack registers a callback invoked for each remote media track.
func (c *Conn) OnTrack(fn func(*webrtc.TrackRemote, *webrtc.RTPReceiver)) { c.onTrack = fn }

// OnConnected registers a callback invoked once the call's media transport is
// connected. If the connection is already up, fn is called immediately.
func (c *Conn) OnConnected(fn func()) {
	c.mu.Lock()
	c.onConnected = fn
	already := c.firedConnect
	c.mu.Unlock()
	if already && fn != nil {
		fn()
	}
}

// OnDisconnected registers a callback invoked when the transport is lost.
func (c *Conn) OnDisconnected(fn func()) {
	c.mu.Lock()
	c.onDisconnected = fn
	c.mu.Unlock()
}

// OnStateChange registers a callback invoked on aggregate state changes.
func (c *Conn) OnStateChange(fn func(state string)) {
	c.mu.Lock()
	c.onStateChange = fn
	c.mu.Unlock()
}

// State returns the current aggregate connection state.
func (c *Conn) State() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state == "" {
		return "new"
	}
	return c.state
}

// open builds the pion transport stack from the call's reflector endpoints and
// starts ICE candidate gathering.
func (c *Conn) open(conns []tg.PhoneConnectionClass) error {
	me, err := buildMediaEngine()
	if err != nil {
		return err
	}
	ir := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(me, ir); err != nil {
		return errors.Wrap(err, "register interceptors")
	}

	se := buildSettingEngine()
	servers := iceServers(conns)
	if c.net != nil {
		// Test transport: route over the virtual network and gather only its
		// host candidates (no reachable STUN).
		se.SetNet(c.net)
		servers = nil
	}

	c.api = webrtc.NewAPI(
		webrtc.WithMediaEngine(me),
		webrtc.WithInterceptorRegistry(ir),
		webrtc.WithSettingEngine(se),
	)

	c.gatherer, err = c.api.NewICEGatherer(webrtc.ICEGatherOptions{ICEServers: servers})
	if err != nil {
		return errors.Wrap(err, "new ice gatherer")
	}
	c.ice = c.api.NewICETransport(c.gatherer)

	sk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return errors.Wrap(err, "generate key")
	}
	cert, err := webrtc.GenerateCertificate(sk)
	if err != nil {
		return errors.Wrap(err, "generate certificate")
	}
	c.dtls, err = c.api.NewDTLSTransport(c.ice, []webrtc.Certificate{*cert})
	if err != nil {
		return errors.Wrap(err, "new dtls transport")
	}

	c.gatherer.OnLocalCandidate(func(cand *webrtc.ICECandidate) {
		if cand == nil {
			return // Gathering complete.
		}
		c.sendCandidate(cand)
	})
	c.ice.OnConnectionStateChange(func(s webrtc.ICETransportState) {
		c.log.Debug("ICE state", zap.Stringer("state", s))
		c.mu.Lock()
		c.iceConnected = s == webrtc.ICETransportStateConnected || s == webrtc.ICETransportStateCompleted
		c.mu.Unlock()
		c.updateState()
		if s == webrtc.ICETransportStateFailed || s == webrtc.ICETransportStateClosed {
			c.fireDisconnected()
		}
	})
	c.dtls.OnStateChange(func(s webrtc.DTLSTransportState) {
		c.log.Debug("DTLS state", zap.Stringer("state", s))
		c.mu.Lock()
		c.dtlsConnected = s == webrtc.DTLSTransportStateConnected
		c.mu.Unlock()
		c.updateState()
		if s == webrtc.DTLSTransportStateFailed || s == webrtc.DTLSTransportStateClosed {
			c.fireDisconnected()
		}
	})

	c.neg = newContentNegotiation()
	c.addTracks()

	if err := c.gatherer.Gather(); err != nil {
		return errors.Wrap(err, "gather")
	}
	return nil
}

func (c *Conn) addTracks() {
	c.audioSSRC = randomSSRC()
	c.videoSSRC = c.audioSSRC + 1
	c.audioTrack, _ = webrtc.NewTrackLocalStaticRTP(audioCodec(), "audio", "gotd-audio")
	c.videoTrack, _ = webrtc.NewTrackLocalStaticRTP(videoCodec(), "video", "gotd-video")
}

// start begins the handshake. The caller (offerer) sends the first
// InitialSetup; the callee waits for the peer's.
func (c *Conn) start() error {
	if c.isOutgoing {
		return c.sendInitialSetup("actpass")
	}
	return nil
}

func (c *Conn) sendInitialSetup(setup string) error {
	iceParams, err := c.gatherer.GetLocalParameters()
	if err != nil {
		return errors.Wrap(err, "local ice params")
	}
	dtlsParams, err := c.dtls.GetLocalParameters()
	if err != nil {
		return errors.Wrap(err, "local dtls params")
	}
	var fps []sigFingerprint
	if len(dtlsParams.Fingerprints) > 0 {
		fps = append(fps, sigFingerprint{
			Hash:        dtlsParams.Fingerprints[0].Algorithm,
			Fingerprint: dtlsParams.Fingerprints[0].Value,
			Setup:       setup,
		})
	}
	c.emitJSON(initialSetupMessage{
		Type:         typeInitialSetup,
		Ufrag:        iceParams.UsernameFragment,
		Pwd:          iceParams.Password,
		Fingerprints: fps,
	})
	return nil
}

// onSignal dispatches a decrypted JSON control message.
func (c *Conn) onSignal(data []byte) error {
	typ, err := signalingType(data)
	if err != nil {
		return err
	}
	switch typ {
	case typeInitialSetup:
		return c.handleInitialSetup(data)
	case typeCandidates:
		return c.handleCandidates(data)
	case typeNegotiateChannels:
		return c.handleNegotiate(data)
	case typeMediaState:
		return nil
	default:
		c.log.Debug("Ignoring signaling type", zap.String("type", typ))
		return nil
	}
}

func (c *Conn) handleInitialSetup(data []byte) error {
	var msg initialSetupMessage
	if err := jsonUnmarshal(data, &msg); err != nil {
		return errors.Wrap(err, "decode InitialSetup")
	}

	c.mu.Lock()
	if c.remoteReady {
		c.mu.Unlock()
		return nil
	}
	c.remoteReady = true
	for _, ic := range c.pendingRemote {
		c.addRemoteCandidate(ic)
	}
	c.pendingRemote = nil
	c.mu.Unlock()

	// pion's DTLSTransport.Start takes the remote peer's parameters and derives
	// our own role as the inverse, so we pass the peer's role here.
	remoteRole := remoteDTLSRole(peerSetup(&msg), c.isOutgoing)
	iceRole := webrtc.ICERoleControlled
	if c.isOutgoing {
		iceRole = webrtc.ICERoleControlling
	}
	remoteICE := webrtc.ICEParameters{UsernameFragment: msg.Ufrag, Password: msg.Pwd}
	var fps []webrtc.DTLSFingerprint
	if len(msg.Fingerprints) > 0 {
		fps = append(fps, webrtc.DTLSFingerprint{
			Algorithm: msg.Fingerprints[0].Hash,
			Value:     msg.Fingerprints[0].Fingerprint,
		})
	}

	go func() {
		if err := c.ice.Start(c.gatherer, remoteICE, &iceRole); err != nil {
			c.log.Warn("ICE start", zap.Error(err))
			return
		}
		if err := c.dtls.Start(webrtc.DTLSParameters{Role: remoteRole, Fingerprints: fps}); err != nil {
			c.log.Warn("DTLS start", zap.Error(err))
			return
		}
		c.onDTLSStarted()
	}()

	if !c.isOutgoing {
		return c.sendInitialSetup("passive")
	}
	return nil
}

func (c *Conn) onDTLSStarted() {
	c.mu.Lock()
	c.dtlsStarted = true
	offer := c.neg.proposeChannels(c.audioSSRC, c.videoSSRC)
	c.mu.Unlock()
	if offer != nil {
		c.emitJSON(offer)
	}
	c.maybeCreateChannels()
}

func (c *Conn) handleCandidates(data []byte) error {
	var msg candidatesMessage
	if err := jsonUnmarshal(data, &msg); err != nil {
		return errors.Wrap(err, "decode Candidates")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, ic := range msg.Candidates {
		if !c.remoteReady {
			c.pendingRemote = append(c.pendingRemote, ic)
			continue
		}
		c.addRemoteCandidate(ic)
	}
	return nil
}

// addRemoteCandidate parses and adds an SDP candidate line. Caller holds c.mu.
func (c *Conn) addRemoteCandidate(ic candidateDescription) {
	// Telegram advertises relay candidates as "reflector-<id>-<tag>.reflector"
	// hostnames served by its custom reflector protocol (tgcalls ReflectorPort),
	// which is not RFC TURN and which pion cannot use. Skip any non-IP candidate
	// quietly; connectivity then relies on the direct host/srflx candidates.
	if addr := candidateAddress(ic.SdpString); addr != "" && net.ParseIP(addr) == nil {
		c.log.Debug("Skipping non-IP relay (reflector) candidate", zap.String("address", addr))
		return
	}
	cand, err := parseCandidate(ic.SdpString)
	if err != nil {
		c.log.Debug("Skip remote candidate", zap.Error(err))
		return
	}
	if err := c.ice.AddRemoteCandidate(cand); err != nil {
		c.log.Debug("Add remote candidate", zap.Error(err))
	}
}

// candidateAddress returns the connection-address field of an SDP candidate
// line, or "" if it cannot be located.
func candidateAddress(line string) string {
	s := strings.TrimPrefix(strings.TrimSpace(line), "a=")
	s = strings.TrimPrefix(s, "candidate:")
	// foundation component transport priority address port typ ...
	if f := strings.Fields(s); len(f) >= 5 {
		return f[4]
	}
	return ""
}

func (c *Conn) handleNegotiate(data []byte) error {
	var msg negotiateChannelsMessage
	if err := jsonUnmarshal(data, &msg); err != nil {
		return errors.Wrap(err, "decode NegotiateChannels")
	}
	c.mu.Lock()
	if c.negotiated {
		c.mu.Unlock()
		return nil
	}
	reply, ready := c.neg.applyRemoteChannels(&msg, c.audioSSRC, c.videoSSRC)
	if ready {
		c.negotiated = true
	}
	c.mu.Unlock()
	if reply != nil {
		c.emitJSON(reply)
	}
	if ready {
		c.maybeCreateChannels()
	}
	return nil
}

// maybeCreateChannels wires up RTP senders/receivers once DTLS is up and SSRCs
// have been negotiated.
func (c *Conn) maybeCreateChannels() {
	c.mu.Lock()
	if c.channelsCreated || !c.dtlsStarted || !c.negotiated {
		c.mu.Unlock()
		return
	}
	c.channelsCreated = true
	peerAudio := c.neg.peerAudioSSRC()
	peerVideo := c.neg.peerVideoSSRC()
	c.mu.Unlock()

	c.sendTrack(c.audioTrack, c.audioSSRC, 111)
	c.recvTrack(webrtc.RTPCodecTypeAudio, peerAudio)
	c.sendTrack(c.videoTrack, c.videoSSRC, 100)
	c.recvTrack(webrtc.RTPCodecTypeVideo, peerVideo)

	c.sendMediaState()
}

func (c *Conn) sendTrack(track *webrtc.TrackLocalStaticRTP, ssrc uint32, pt webrtc.PayloadType) {
	sender, err := c.api.NewRTPSender(track, c.dtls)
	if err != nil {
		c.log.Warn("New RTP sender", zap.Error(err))
		return
	}
	if err := sender.Send(webrtc.RTPSendParameters{
		Encodings: []webrtc.RTPEncodingParameters{{
			RTPCodingParameters: webrtc.RTPCodingParameters{SSRC: webrtc.SSRC(ssrc), PayloadType: pt},
		}},
	}); err != nil {
		c.log.Warn("RTP sender send", zap.Error(err))
		return
	}
	go drainSenderRTCP(sender)
}

func (c *Conn) recvTrack(kind webrtc.RTPCodecType, ssrc uint32) {
	if ssrc == 0 {
		return
	}
	receiver, err := c.api.NewRTPReceiver(kind, c.dtls)
	if err != nil {
		c.log.Warn("New RTP receiver", zap.Error(err))
		return
	}
	if err := receiver.Receive(webrtc.RTPReceiveParameters{
		Encodings: []webrtc.RTPDecodingParameters{{
			RTPCodingParameters: webrtc.RTPCodingParameters{SSRC: webrtc.SSRC(ssrc)},
		}},
	}); err != nil {
		c.log.Warn("RTP receiver receive", zap.Error(err))
		return
	}
	go drainReceiverRTCP(receiver)
	if c.onTrack != nil {
		c.onTrack(receiver.Track(), receiver)
	}
}

func (c *Conn) sendMediaState() {
	c.emitJSON(mediaStateMessage{
		Type:            typeMediaState,
		VideoState:      "active",
		ScreencastState: "inactive",
	})
}

func (c *Conn) sendCandidate(cand *webrtc.ICECandidate) {
	line := cand.ToJSON().Candidate
	if line == "" {
		return
	}
	c.emitJSON(candidatesMessage{
		Type:       typeCandidates,
		Candidates: []candidateDescription{{SdpString: line}},
	})
}

func (c *Conn) emitJSON(v any) {
	data, err := jsonMarshal(v)
	if err != nil {
		c.log.Warn("Encode signaling", zap.Error(err))
		return
	}
	if c.emit != nil {
		c.emit(data)
	}
}

func (c *Conn) updateState() {
	c.mu.Lock()
	connected := c.iceConnected && c.dtlsConnected && !c.firedConnect
	if connected {
		c.firedConnect = true
		c.state = "connected"
	}
	onConnected, onState := c.onConnected, c.onStateChange
	c.mu.Unlock()

	if connected {
		if onState != nil {
			onState("connected")
		}
		if onConnected != nil {
			onConnected()
		}
	}
}

func (c *Conn) fireDisconnected() {
	c.mu.Lock()
	if c.firedDisconnect {
		c.mu.Unlock()
		return
	}
	c.firedDisconnect = true
	c.state = "closed"
	fn := c.onDisconnected
	c.mu.Unlock()
	if fn != nil {
		fn()
	}
}

// Close tears down the transport.
//
// It uses pion's graceful shutdown variants, which block until the ICE agent's
// asynchronous state-change notifier goroutines have drained. Otherwise the
// queued "closed" callback could fire (and log) after the caller — e.g. a test
// — has already returned.
func (c *Conn) Close() error {
	c.mu.Lock()
	dtls, ice, gatherer := c.dtls, c.ice, c.gatherer
	c.mu.Unlock()
	if dtls != nil {
		_ = dtls.Stop()
	}
	if ice != nil {
		_ = ice.GracefulStop()
	}
	if gatherer != nil {
		_ = gatherer.GracefulClose()
	}
	return nil
}

func drainSenderRTCP(s *webrtc.RTPSender) {
	for {
		if _, _, err := s.ReadRTCP(); err != nil {
			return
		}
	}
}

func drainReceiverRTCP(r *webrtc.RTPReceiver) {
	for {
		if _, _, err := r.ReadRTCP(); err != nil {
			return
		}
	}
}

// remoteDTLSRole returns the peer's DTLS role to hand to pion's
// DTLSTransport.Start, which then takes the inverse for us. tgcalls advertises
// setup=actpass for the caller (which becomes the DTLS client) and
// setup=passive for the callee (the DTLS server). So: a "passive" peer is the
// server (we become client); an "active" peer is the client (we become
// server); and for an undecided "actpass" peer we fall back to call direction —
// the caller is the client, hence the remote callee is the server.
func remoteDTLSRole(peerSetup string, isOutgoing bool) webrtc.DTLSRole {
	switch peerSetup {
	case "active":
		return webrtc.DTLSRoleClient
	case "passive":
		return webrtc.DTLSRoleServer
	default: // actpass
		if isOutgoing {
			return webrtc.DTLSRoleServer
		}
		return webrtc.DTLSRoleClient
	}
}

func peerSetup(msg *initialSetupMessage) string {
	if len(msg.Fingerprints) > 0 && msg.Fingerprints[0].Setup != "" {
		return msg.Fingerprints[0].Setup
	}
	return "actpass"
}

// parseCandidate converts a tgcalls SDP candidate line into a pion ICE
// candidate. pion does not export a constructor from an SDP string, so we
// parse via the ice package and map the fields onto webrtc.ICECandidate.
func parseCandidate(line string) (*webrtc.ICECandidate, error) {
	raw := strings.TrimPrefix(strings.TrimSpace(line), "a=")
	ic, err := pion.UnmarshalCandidate(raw)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal candidate")
	}
	typ, err := candidateType(ic.Type())
	if err != nil {
		return nil, err
	}
	proto, err := webrtc.NewICEProtocol(ic.NetworkType().NetworkShort())
	if err != nil {
		return nil, errors.Wrap(err, "ice protocol")
	}
	out := &webrtc.ICECandidate{
		Foundation: ic.Foundation(),
		Priority:   ic.Priority(),
		Address:    ic.Address(),
		Protocol:   proto,
		Port:       uint16(ic.Port()),
		Component:  ic.Component(),
		Typ:        typ,
		TCPType:    ic.TCPType().String(),
	}
	if ra := ic.RelatedAddress(); ra != nil {
		out.RelatedAddress = ra.Address
		out.RelatedPort = uint16(ra.Port)
	}
	return out, nil
}

func candidateType(t pion.CandidateType) (webrtc.ICECandidateType, error) {
	switch t {
	case pion.CandidateTypeHost:
		return webrtc.ICECandidateTypeHost, nil
	case pion.CandidateTypeServerReflexive:
		return webrtc.ICECandidateTypeSrflx, nil
	case pion.CandidateTypePeerReflexive:
		return webrtc.ICECandidateTypePrflx, nil
	case pion.CandidateTypeRelay:
		return webrtc.ICECandidateTypeRelay, nil
	default:
		return webrtc.ICECandidateType(0), errors.Errorf("unknown candidate type %q", t)
	}
}

// iceServers builds the STUN/TURN server list from the call's reflector
// endpoints, plus public STUN fallbacks.
func iceServers(conns []tg.PhoneConnectionClass) []webrtc.ICEServer {
	var servers []webrtc.ICEServer
	for _, conn := range conns {
		w, ok := conn.(*tg.PhoneConnectionWebrtc)
		if !ok {
			continue
		}
		var hosts []string
		if w.IP != "" {
			hosts = append(hosts, w.IP)
		}
		if w.Ipv6 != "" {
			hosts = append(hosts, "["+w.Ipv6+"]")
		}
		for _, host := range hosts {
			addr := net.JoinHostPort(strings.Trim(host, "[]"), strconv.Itoa(w.Port))
			if w.Turn {
				servers = append(servers, webrtc.ICEServer{
					URLs:       []string{"turn:" + addr + "?transport=udp"},
					Username:   w.Username,
					Credential: w.Password,
				})
			}
			if w.Stun {
				servers = append(servers, webrtc.ICEServer{URLs: []string{"stun:" + addr}})
			}
		}
	}
	servers = append(servers,
		webrtc.ICEServer{URLs: []string{"stun:stun.l.google.com:19302"}},
		webrtc.ICEServer{URLs: []string{"stun:stun1.l.google.com:19302"}},
	)
	return servers
}

func randomSSRC() uint32 {
	var buf [4]byte
	_, _ = rand.Read(buf[:])
	ssrc := uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])
	ssrc &= 0x7fffffff
	if ssrc == 0 {
		ssrc = 1
	}
	return ssrc
}
