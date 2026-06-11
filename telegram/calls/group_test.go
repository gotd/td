package calls

import (
	"strings"
	"testing"
	"time"

	"github.com/gotd/td/tg"
)

func TestBuildAnswerSDP(t *testing.T) {
	resp := groupJoinResponse{
		Transport: groupTransportDescription{
			Ufrag:        "abcd",
			Pwd:          "secretpassword",
			Fingerprints: []groupFingerprint{{Hash: "sha-256", Fingerprint: "AA:BB"}},
			Candidates: []groupCandidate{{
				Foundation: "1", Component: "1", Protocol: "udp", Priority: "1000",
				IP: "149.154.1.2", Port: "443", Type: "host", Generation: "0",
			}},
		},
		Audio: &groupMediaDescription{
			PayloadTypes: []groupPayloadType{{
				ID: 111, Name: "opus", Clockrate: 48000, Channels: 2,
				FeedbackTypes: []groupFeedback{{Type: "transport-cc"}},
			}},
			RTPExtensions: []groupRTPExtension{{ID: 1, URI: "urn:ietf:params:rtp-hdrext:ssrc-audio-level"}},
		},
	}

	sdp := buildAnswerSDP(resp)
	for _, want := range []string{
		"v=0",
		"a=ice-lite",
		"m=audio 443 RTP/SAVPF 111",
		"c=IN IP4 149.154.1.2",
		"a=ice-ufrag:abcd",
		"a=ice-pwd:secretpassword",
		"a=fingerprint:sha-256 AA:BB",
		"a=setup:active",
		"a=candidate:1 1 udp 1000 149.154.1.2 443 typ host generation 0",
		"a=rtpmap:111 opus/48000/2",
		"a=rtcp-fb:111 transport-cc",
		"a=extmap:1 urn:ietf:params:rtp-hdrext:ssrc-audio-level",
		"a=rtcp-mux",
	} {
		if !strings.Contains(sdp, want) {
			t.Errorf("answer SDP missing %q\n---\n%s", want, sdp)
		}
	}
}

func TestExtractSDPParams(t *testing.T) {
	sdp := strings.Join([]string{
		"v=0",
		"a=ice-ufrag:myufrag",
		"a=ice-pwd:mypwd",
		"a=fingerprint:sha-256 DE:AD:BE:EF",
		"a=setup:actpass",
	}, "\r\n")

	ufrag, pwd, fp, hash := extractSDPParams(sdp)
	if ufrag != "myufrag" || pwd != "mypwd" || hash != "sha-256" || fp != "DE:AD:BE:EF" {
		t.Fatalf("got ufrag=%q pwd=%q hash=%q fp=%q", ufrag, pwd, hash, fp)
	}
}

func TestConnectionParams(t *testing.T) {
	const want = `{"transport":{"ufrag":"x"}}`
	updates := &tg.Updates{
		Updates: []tg.UpdateClass{
			&tg.UpdateGroupCall{},
			&tg.UpdateGroupCallConnection{Params: tg.DataJSON{Data: want}},
		},
	}
	got, err := connectionParams(updates)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	if _, err := connectionParams(&tg.Updates{}); err == nil {
		t.Fatal("expected error when no connection update present")
	}
}

func TestNTPTime(t *testing.T) {
	// 1970-01-01T00:00:00Z is NTP seconds 2208988800, zero fraction.
	got := ntpTime(time.Unix(0, 0))
	if high := got >> 32; high != 2208988800 {
		t.Fatalf("ntp seconds = %d, want 2208988800", high)
	}
	if low := got & 0xffffffff; low != 0 {
		t.Fatalf("ntp fraction = %d, want 0", low)
	}
}
