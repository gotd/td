package calls

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"strconv"

	"github.com/go-faster/errors"
)

// tgcalls JSON signaling message types (the "@type" discriminator).
const (
	typeInitialSetup      = "InitialSetup"
	typeCandidates        = "Candidates"
	typeMediaState        = "MediaState"
	typeNegotiateChannels = "NegotiateChannels"
)

type sigFingerprint struct {
	Hash        string `json:"hash"`
	Setup       string `json:"setup"`
	Fingerprint string `json:"fingerprint"`
}

type sigFeedback struct {
	Type string `json:"type"`
	// Subtype must always be serialized (even empty): tgcalls' FeedbackType
	// parser rejects the whole message if the "subtype" key is missing, which
	// would make the peer ignore our channel negotiation entirely.
	Subtype string `json:"subtype"`
}

type sigPayloadType struct {
	ID            int               `json:"id"`
	Name          string            `json:"name"`
	Clockrate     int               `json:"clockrate"`
	Channels      int               `json:"channels,omitempty"`
	FeedbackTypes []sigFeedback     `json:"feedbackTypes,omitempty"`
	Parameters    map[string]string `json:"parameters,omitempty"`
}

type sigSsrcGroup struct {
	Semantics string   `json:"semantics"`
	Ssrcs     []string `json:"ssrcs"`
}

type sigExtension struct {
	ID  int    `json:"id"`
	URI string `json:"uri"`
}

// initialSetupMessage carries the local ICE credentials and DTLS fingerprint.
type initialSetupMessage struct {
	Type         string           `json:"@type"`
	Ufrag        string           `json:"ufrag"`
	Pwd          string           `json:"pwd"`
	Renomination bool             `json:"renomination"`
	Fingerprints []sigFingerprint `json:"fingerprints"`
}

type candidateDescription struct {
	SdpString string `json:"sdpString"`
}

// candidatesMessage carries trickled ICE candidates.
type candidatesMessage struct {
	Type       string                 `json:"@type"`
	Candidates []candidateDescription `json:"candidates"`
}

// mediaStateMessage reports mute/video state to the peer.
type mediaStateMessage struct {
	Type            string `json:"@type"`
	Muted           bool   `json:"muted"`
	LowBattery      bool   `json:"lowBattery"`
	VideoState      string `json:"videoState"`
	VideoRotation   int    `json:"videoRotation"`
	ScreencastState string `json:"screencastState"`
}

type mediaContent struct {
	Type          string           `json:"type"`
	Ssrc          string           `json:"ssrc"`
	SsrcGroups    []sigSsrcGroup   `json:"ssrcGroups,omitempty"`
	PayloadTypes  []sigPayloadType `json:"payloadTypes"`
	RtpExtensions []sigExtension   `json:"rtpExtensions,omitempty"`
}

// negotiateChannelsMessage exchanges SSRCs and codec parameters for the media tracks.
type negotiateChannelsMessage struct {
	Type       string         `json:"@type"`
	ExchangeID string         `json:"exchangeId"`
	Contents   []mediaContent `json:"contents"`
}

// envelope is used to peek at the message type before full decoding.
type envelope struct {
	Type string `json:"@type"`
}

// signalingType extracts the "@type" of a raw JSON control message.
func signalingType(data []byte) (string, error) {
	var e envelope
	if err := json.Unmarshal(data, &e); err != nil {
		return "", errors.Wrap(err, "decode envelope")
	}
	if e.Type == "" {
		return "", errors.New("signaling message missing @type")
	}
	return e.Type, nil
}

// contentNegotiation tracks the NegotiateChannels handshake and the peer's
// SSRCs, mirroring tgcalls' ContentNegotiationContext.
type contentNegotiation struct {
	localExchangeID string
	offered         bool

	peerAudio uint32
	peerVideo uint32
}

func newContentNegotiation() *contentNegotiation { return &contentNegotiation{} }

func (n *contentNegotiation) peerAudioSSRC() uint32 { return n.peerAudio }
func (n *contentNegotiation) peerVideoSSRC() uint32 { return n.peerVideo }

// proposeChannels builds our NegotiateChannels offer once.
func (n *contentNegotiation) proposeChannels(audioSSRC, videoSSRC uint32) *negotiateChannelsMessage {
	if n.offered {
		return nil
	}
	n.offered = true
	n.localExchangeID = randomExchangeID()
	return &negotiateChannelsMessage{
		Type:       typeNegotiateChannels,
		ExchangeID: n.localExchangeID,
		Contents:   []mediaContent{audioContent(audioSSRC), videoContent(videoSSRC)},
	}
}

// applyRemoteChannels processes a peer NegotiateChannels message, returning a
// reply to send (nil if the message answered our own offer) and whether
// negotiation is complete.
func (n *contentNegotiation) applyRemoteChannels(msg *negotiateChannelsMessage, audioSSRC, videoSSRC uint32) (reply *negotiateChannelsMessage, ready bool) {
	n.captureSSRCs(msg)
	if n.offered && msg.ExchangeID == n.localExchangeID {
		return nil, true
	}
	return &negotiateChannelsMessage{
		Type:       typeNegotiateChannels,
		ExchangeID: msg.ExchangeID,
		Contents:   []mediaContent{audioContent(audioSSRC), videoContent(videoSSRC)},
	}, true
}

func (n *contentNegotiation) captureSSRCs(msg *negotiateChannelsMessage) {
	for _, c := range msg.Contents {
		ssrc := parseSSRC(c.Ssrc)
		if ssrc == 0 {
			continue
		}
		switch c.Type {
		case "audio":
			if n.peerAudio == 0 {
				n.peerAudio = ssrc
			}
		case "video":
			if n.peerVideo == 0 {
				n.peerVideo = ssrc
			}
		}
	}
}

func audioContent(ssrc uint32) mediaContent {
	return mediaContent{
		Type: "audio",
		Ssrc: ssrcString(ssrc),
		PayloadTypes: []sigPayloadType{{
			ID: 111, Name: "opus", Clockrate: 48000, Channels: 2,
			FeedbackTypes: []sigFeedback{{Type: "transport-cc"}},
			Parameters:    map[string]string{"minptime": "10", "useinbandfec": "1"},
		}},
		RtpExtensions: rtpExtensions(),
	}
}

func videoContent(ssrc uint32) mediaContent {
	return mediaContent{
		Type: "video",
		Ssrc: ssrcString(ssrc),
		SsrcGroups: []sigSsrcGroup{{
			Semantics: "FID",
			Ssrcs:     []string{ssrcString(ssrc), ssrcString(ssrc + 1)},
		}},
		PayloadTypes: []sigPayloadType{{
			ID: 100, Name: "VP8", Clockrate: 90000,
			FeedbackTypes: []sigFeedback{
				{Type: "goog-remb"}, {Type: "transport-cc"},
				{Type: "ccm", Subtype: "fir"},
				{Type: "nack"}, {Type: "nack", Subtype: "pli"},
			},
		}},
		RtpExtensions: rtpExtensions(),
	}
}

func rtpExtensions() []sigExtension {
	return []sigExtension{
		{ID: 2, URI: "http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time"},
		{ID: 3, URI: "http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01"},
	}
}

func ssrcString(ssrc uint32) string { return strconv.FormatUint(uint64(ssrc), 10) }

func parseSSRC(s string) uint32 {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(v)
}

func randomExchangeID() string {
	var buf [4]byte
	_, _ = rand.Read(buf[:])
	id := binary.BigEndian.Uint32(buf[:]) & 0x7fffffff
	if id == 0 {
		id = 1
	}
	return strconv.FormatUint(uint64(id), 10)
}
