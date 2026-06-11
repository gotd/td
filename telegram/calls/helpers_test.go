package calls

import (
	"testing"

	"github.com/pion/webrtc/v4"

	"github.com/gotd/td/tg"
)

func TestRemoteDTLSRole(t *testing.T) {
	cases := []struct {
		setup      string
		isOutgoing bool
		want       webrtc.DTLSRole
	}{
		{"active", true, webrtc.DTLSRoleClient},
		{"passive", true, webrtc.DTLSRoleServer},  // caller: peer passive => we client
		{"passive", false, webrtc.DTLSRoleServer}, //
		{"actpass", true, webrtc.DTLSRoleServer},  // caller, undecided peer
		{"actpass", false, webrtc.DTLSRoleClient}, // callee, undecided peer => we server
		{"", false, webrtc.DTLSRoleClient},
	}
	for _, tc := range cases {
		if got := remoteDTLSRole(tc.setup, tc.isOutgoing); got != tc.want {
			t.Errorf("remoteDTLSRole(%q,%v) = %v, want %v", tc.setup, tc.isOutgoing, got, tc.want)
		}
	}
}

func TestPeerSetup(t *testing.T) {
	if got := peerSetup(&initialSetupMessage{}); got != "actpass" {
		t.Fatalf("empty fingerprints => %q, want actpass", got)
	}
	msg := &initialSetupMessage{Fingerprints: []sigFingerprint{{Setup: "passive"}}}
	if got := peerSetup(msg); got != "passive" {
		t.Fatalf("got %q, want passive", got)
	}
}

func TestCandidateAddress(t *testing.T) {
	cases := map[string]string{
		"candidate:1 1 udp 2122260223 192.168.1.5 54321 typ host":     "192.168.1.5",
		"a=candidate:1 1 udp 1 reflector-1-9.reflector 443 typ relay": "reflector-1-9.reflector",
		"garbage": "",
	}
	for line, want := range cases {
		if got := candidateAddress(line); got != want {
			t.Errorf("candidateAddress(%q) = %q, want %q", line, got, want)
		}
	}
}

func TestCandidateTypeMapping(t *testing.T) {
	if _, err := candidateType(0xff); err == nil {
		t.Fatal("expected error for unknown candidate type")
	}
}

func TestParseCandidateTypes(t *testing.T) {
	for _, line := range []string{
		"candidate:1 1 udp 2122260223 192.168.1.5 54321 typ host generation 0",
		"candidate:2 1 udp 1686052607 1.2.3.4 50000 typ srflx raddr 192.168.1.5 rport 54321 generation 0",
	} {
		if _, err := parseCandidate(line); err != nil {
			t.Errorf("parseCandidate(%q): %v", line, err)
		}
	}
	if _, err := parseCandidate("not a candidate"); err == nil {
		t.Fatal("expected error for malformed candidate")
	}
}

func TestICEServers(t *testing.T) {
	conns := []tg.PhoneConnectionClass{
		&tg.PhoneConnectionWebrtc{
			Turn: true, Stun: true,
			IP: "1.2.3.4", Ipv6: "2001:db8::1", Port: 443,
			Username: "u", Password: "p",
		},
		&tg.PhoneConnection{}, // non-webrtc, ignored
	}
	servers := iceServers(conns)
	var turn, stun, google bool
	for _, s := range servers {
		for _, u := range s.URLs {
			switch {
			case len(u) > 5 && u[:5] == "turn:":
				turn = true
			case len(u) > 5 && u[:5] == "stun:" && s.Username == "":
				stun = true
			}
			if u == "stun:stun.l.google.com:19302" {
				google = true
			}
		}
	}
	if !turn || !stun || !google {
		t.Fatalf("missing servers: turn=%v stun=%v google=%v (%d total)", turn, stun, google, len(servers))
	}
}

func TestBuildMediaEngineAndSettingEngine(t *testing.T) {
	if _, err := buildMediaEngine(); err != nil {
		t.Fatalf("buildMediaEngine: %v", err)
	}
	_ = buildSettingEngine()
	if audioCodec().MimeType != webrtc.MimeTypeOpus {
		t.Fatal("audio codec is not opus")
	}
	if videoCodec().MimeType != webrtc.MimeTypeVP8 {
		t.Fatal("video codec is not vp8")
	}
}

func TestRandomSSRC(t *testing.T) {
	for range 100 {
		if s := randomSSRC(); s == 0 || s > 0x7fffffff {
			t.Fatalf("ssrc out of range: %d", s)
		}
	}
}
