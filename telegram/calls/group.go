package calls

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-faster/errors"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"

	"github.com/gotd/log"

	"github.com/gotd/td/tg"
)

// GroupCall joins a Telegram group call (voice chat) and streams audio into it.
//
// Unlike a 1:1 [Conn], a group call connects to Telegram's SFU rather than to a
// peer: a single WebRTC PeerConnection carries our outgoing audio and any
// incoming tracks the server forwards. Outgoing audio is written with
// [GroupCall.WriteAudio] (which also feeds RTCP sender reports) or, for raw
// access, via [GroupCall.AudioTrack].
type GroupCall struct {
	api  *tg.Client
	log  log.Helper
	conn *groupConn

	mu     sync.Mutex
	call   *tg.InputGroupCall
	joined bool

	// RTCP sender-report counters for the outgoing audio stream.
	packets  atomic.Uint32
	octets   atomic.Uint32
	lastTS   atomic.Uint32
	rtcpStop chan struct{}

	onParticipants func([]tg.GroupCallParticipant)
}

// NewGroupCall returns a group call bound to the given invoker.
func NewGroupCall(api *tg.Client, opts Options) *GroupCall {
	opts.setDefaults()
	return &GroupCall{api: api, log: log.For(opts.Logger)}
}

// OnConnected registers a callback fired when the media transport connects.
func (g *GroupCall) OnConnected(fn func()) { g.connOnce().onConnected = fn }

// OnDisconnected registers a callback fired when the transport is lost.
func (g *GroupCall) OnDisconnected(fn func()) { g.connOnce().onDisconnected = fn }

// OnTrack registers a callback fired for each incoming forwarded track.
func (g *GroupCall) OnTrack(fn func(*webrtc.TrackRemote, *webrtc.RTPReceiver)) {
	g.connOnce().onTrack = fn
}

// OnParticipants registers a callback fired on participant-list updates (only
// after Register has wired the dispatcher).
func (g *GroupCall) OnParticipants(fn func([]tg.GroupCallParticipant)) { g.onParticipants = fn }

// connOnce lazily creates the transport so callbacks can be set before Join.
func (g *GroupCall) connOnce() *groupConn {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.conn == nil {
		g.conn = newGroupConn(g.log)
	}
	return g.conn
}

// Register installs the group-call participant handler on the dispatcher.
func (g *GroupCall) Register(d tg.UpdateDispatcher) {
	d.OnGroupCallParticipants(func(_ context.Context, _ tg.Entities, u *tg.UpdateGroupCallParticipants) error {
		g.mu.Lock()
		active := g.call != nil && sameCall(g.call, u.Call)
		g.mu.Unlock()
		if active && g.onParticipants != nil {
			g.onParticipants(u.Participants)
		}
		return nil
	})
}

// Join joins the group call identified by call, presenting joinAs as the
// speaking peer (typically the logged-in user). It blocks until the media
// transport connects or ctx is done.
func (g *GroupCall) Join(ctx context.Context, call *tg.InputGroupCall, joinAs tg.InputPeerClass) error {
	conn := g.connOnce()

	connected := make(chan struct{})
	var once sync.Once
	prevConnected := conn.onConnected
	conn.onConnected = func() {
		once.Do(func() { close(connected) })
		if prevConnected != nil {
			prevConnected()
		}
	}

	if err := conn.open(); err != nil {
		return errors.Wrap(err, "open transport")
	}
	if err := conn.addAudioTrack(); err != nil {
		return errors.Wrap(err, "add audio track")
	}

	payload, err := conn.buildJoinPayload()
	if err != nil {
		return errors.Wrap(err, "build join payload")
	}

	updates, err := g.api.PhoneJoinGroupCall(ctx, &tg.PhoneJoinGroupCallRequest{
		Call:   call,
		JoinAs: joinAs,
		Params: tg.DataJSON{Data: payload},
	})
	if err != nil {
		return errors.Wrap(err, "join group call")
	}
	params, err := connectionParams(updates)
	if err != nil {
		return err
	}

	g.mu.Lock()
	g.call = call
	g.joined = true
	g.mu.Unlock()

	if err := conn.connect(params); err != nil {
		return errors.Wrap(err, "apply server response")
	}

	select {
	case <-connected:
		g.startRTCP()
		return nil
	case <-time.After(30 * time.Second):
		_ = g.Leave(ctx)
		return errors.New("timed out waiting for group call to connect")
	case <-ctx.Done():
		_ = g.Leave(ctx)
		return ctx.Err()
	}
}

// WriteAudio writes one Opus RTP packet to the call and updates the RTCP
// sender-report counters. The packet's SSRC and payload type are rewritten to
// the negotiated values, so only sequence number and timestamp matter.
func (g *GroupCall) WriteAudio(pkt *rtp.Packet) error {
	g.mu.Lock()
	conn := g.conn
	g.mu.Unlock()
	if conn == nil || conn.audioTrack == nil {
		return errors.New("not joined")
	}
	g.packets.Add(1)
	g.octets.Add(uint32(len(pkt.Payload)))
	g.lastTS.Store(pkt.Timestamp)
	return conn.audioTrack.WriteRTP(pkt)
}

// AudioTrack returns the raw outgoing audio track for advanced use. Prefer
// WriteAudio, which also drives RTCP sender reports.
func (g *GroupCall) AudioTrack() *webrtc.TrackLocalStaticRTP {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.conn == nil {
		return nil
	}
	return g.conn.audioTrack
}

// AudioSSRC returns the SSRC of the outgoing audio stream.
func (g *GroupCall) AudioSSRC() uint32 {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.conn == nil {
		return 0
	}
	return g.conn.audioSSRC
}

// Leave leaves the call and tears down the transport.
func (g *GroupCall) Leave(ctx context.Context) error {
	g.mu.Lock()
	conn, call, joined := g.conn, g.call, g.joined
	g.call = nil
	g.joined = false
	if g.rtcpStop != nil {
		close(g.rtcpStop)
		g.rtcpStop = nil
	}
	g.mu.Unlock()

	var firstErr error
	if joined && call != nil {
		source := 0
		if conn != nil {
			source = int(int32(conn.audioSSRC))
		}
		if _, err := g.api.PhoneLeaveGroupCall(ctx, &tg.PhoneLeaveGroupCallRequest{
			Call:   call,
			Source: source,
		}); err != nil {
			firstErr = errors.Wrap(err, "leave group call")
		}
	}
	if conn != nil {
		if err := conn.close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// startRTCP periodically sends an RTCP sender report for the outgoing audio so
// the SFU keeps the stream active and can synchronise it for other listeners.
func (g *GroupCall) startRTCP() {
	g.mu.Lock()
	conn := g.conn
	stop := make(chan struct{})
	g.rtcpStop = stop
	g.mu.Unlock()
	if conn == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
			}
			switch conn.connectionState() {
			case webrtc.PeerConnectionStateClosed, webrtc.PeerConnectionStateFailed:
				return
			}
			sr := &rtcp.SenderReport{
				SSRC:        conn.audioSSRC,
				NTPTime:     ntpTime(time.Now()),
				RTPTime:     g.lastTS.Load(),
				PacketCount: g.packets.Load(),
				OctetCount:  g.octets.Load(),
			}
			if err := conn.writeRTCP([]rtcp.Packet{sr}); err != nil {
				g.log.Debug(context.Background(), "Write RTCP SR", log.Error(err))
			}
		}
	}()
}

// connectionParams extracts the SFU connection JSON from a joinGroupCall result.
func connectionParams(updates tg.UpdatesClass) (string, error) {
	var list []tg.UpdateClass
	switch u := updates.(type) {
	case *tg.Updates:
		list = u.Updates
	case *tg.UpdatesCombined:
		list = u.Updates
	}
	for _, upd := range list {
		if c, ok := upd.(*tg.UpdateGroupCallConnection); ok {
			return c.Params.Data, nil
		}
	}
	return "", errors.New("no updateGroupCallConnection in join response")
}

func sameCall(a *tg.InputGroupCall, b tg.InputGroupCallClass) bool {
	other, ok := b.(*tg.InputGroupCall)
	return ok && a.ID == other.ID
}

// ntpTime converts a wall-clock time to a 64-bit NTP timestamp (seconds since
// 1900 in the high 32 bits, fractional seconds in the low 32).
func ntpTime(t time.Time) uint64 {
	const epochOffset = 2208988800 // seconds between 1900 and 1970
	secs := uint64(t.Unix()) + epochOffset
	frac := uint64(t.Nanosecond()) << 32 / 1_000_000_000
	return secs<<32 | frac
}
