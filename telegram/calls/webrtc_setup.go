package calls

import (
	"encoding/json"
	"time"

	"github.com/go-faster/errors"
	"github.com/pion/webrtc/v4"
)

func audioCodec() webrtc.RTPCodecCapability {
	return webrtc.RTPCodecCapability{
		MimeType:     webrtc.MimeTypeOpus,
		ClockRate:    48000,
		Channels:     2,
		SDPFmtpLine:  "minptime=10;useinbandfec=1",
		RTCPFeedback: []webrtc.RTCPFeedback{{Type: "transport-cc"}},
	}
}

func videoCodec() webrtc.RTPCodecCapability {
	return webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeVP8,
		ClockRate: 90000,
		RTCPFeedback: []webrtc.RTCPFeedback{
			{Type: "goog-remb"},
			{Type: "transport-cc"},
			{Type: "ccm", Parameter: "fir"},
			{Type: "nack"},
			{Type: "nack", Parameter: "pli"},
		},
	}
}

// buildMediaEngine registers the Opus and VP8 codecs and the RTP header
// extensions used by Telegram calls.
func buildMediaEngine() (*webrtc.MediaEngine, error) {
	m := &webrtc.MediaEngine{}
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: audioCodec(),
		PayloadType:        111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		return nil, errors.Wrap(err, "register opus")
	}
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: videoCodec(),
		PayloadType:        100,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		return nil, errors.Wrap(err, "register vp8")
	}

	exts := []struct {
		uri  string
		kind webrtc.RTPCodecType
	}{
		{"http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time", webrtc.RTPCodecTypeAudio},
		{"http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01", webrtc.RTPCodecTypeAudio},
		{"http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time", webrtc.RTPCodecTypeVideo},
		{"http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01", webrtc.RTPCodecTypeVideo},
	}
	for _, e := range exts {
		if err := m.RegisterHeaderExtension(
			webrtc.RTPHeaderExtensionCapability{URI: e.uri}, e.kind,
		); err != nil {
			return nil, errors.Wrapf(err, "register header extension %s", e.uri)
		}
	}
	return m, nil
}

// buildSettingEngine tunes ICE timeouts and network selection for direct calls.
func buildSettingEngine() webrtc.SettingEngine {
	se := webrtc.SettingEngine{}
	se.SetICETimeouts(30*time.Second, 60*time.Second, 2*time.Second)
	se.SetSrflxAcceptanceMinWait(0)
	se.SetNetworkTypes([]webrtc.NetworkType{
		webrtc.NetworkTypeUDP4,
		webrtc.NetworkTypeUDP6,
	})
	return se
}

func jsonMarshal(v any) ([]byte, error)      { return json.Marshal(v) }
func jsonUnmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
