package calls

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestNegotiateContentFeedbackSubtype guards against regressing the omitempty
// bug: tgcalls' FeedbackType parser rejects the whole NegotiateChannels message
// if any feedbackType lacks a "subtype" key, which silently breaks media setup.
func TestNegotiateContentFeedbackSubtype(t *testing.T) {
	for _, content := range []mediaContent{audioContent(1000), videoContent(1002)} {
		data, err := json.Marshal(content)
		if err != nil {
			t.Fatal(err)
		}
		// Every feedbackType object must carry a "subtype" field.
		fbCount := 0
		for _, pt := range content.PayloadTypes {
			fbCount += len(pt.FeedbackTypes)
		}
		if got := strings.Count(string(data), `"subtype"`); got != fbCount {
			t.Fatalf("%s: %d subtype keys, want %d in %s", content.Type, got, fbCount, data)
		}
	}
}

func TestSignalingType(t *testing.T) {
	typ, err := signalingType([]byte(`{"@type":"InitialSetup","ufrag":"x"}`))
	if err != nil {
		t.Fatal(err)
	}
	if typ != typeInitialSetup {
		t.Fatalf("type = %q, want %q", typ, typeInitialSetup)
	}
	if _, err := signalingType([]byte(`{"ufrag":"x"}`)); err == nil {
		t.Fatal("missing @type accepted")
	}
	if _, err := signalingType([]byte(`not json`)); err == nil {
		t.Fatal("invalid json accepted")
	}
}

// TestNegotiatorOfferAnswer checks that when our offer is echoed back with the
// same exchange ID, we do not reply again but mark negotiation ready and learn
// the peer's SSRCs.
func TestNegotiatorOfferAnswer(t *testing.T) {
	n := newContentNegotiation()
	offer := n.proposeChannels(1000, 1001)
	if offer == nil {
		t.Fatal("first localOffer returned nil")
	}
	if again := n.proposeChannels(1000, 1001); again != nil {
		t.Fatal("second localOffer should return nil")
	}

	answer := &negotiateChannelsMessage{
		Type:       typeNegotiateChannels,
		ExchangeID: offer.ExchangeID,
		Contents: []mediaContent{
			{Type: "audio", Ssrc: "2000"},
			{Type: "video", Ssrc: "2002"},
		},
	}
	reply, ready := n.applyRemoteChannels(answer, 1000, 1001)
	if reply != nil {
		t.Fatal("answer to our own offer should not be replied to")
	}
	if !ready {
		t.Fatal("negotiation should be ready")
	}
	if n.peerAudioSSRC() != 2000 || n.peerVideoSSRC() != 2002 {
		t.Fatalf("peer ssrcs = %d/%d, want 2000/2002", n.peerAudioSSRC(), n.peerVideoSSRC())
	}
}

// TestNegotiatorRemoteOffer checks that a peer offer with a different exchange
// ID produces an answer carrying our SSRCs.
func TestNegotiatorRemoteOffer(t *testing.T) {
	n := newContentNegotiation()
	remote := &negotiateChannelsMessage{
		Type:       typeNegotiateChannels,
		ExchangeID: "999",
		Contents:   []mediaContent{{Type: "audio", Ssrc: "5000"}},
	}
	reply, ready := n.applyRemoteChannels(remote, 1000, 1001)
	if !ready {
		t.Fatal("negotiation should be ready")
	}
	if reply == nil {
		t.Fatal("expected an answer to remote offer")
	}
	if reply.ExchangeID != "999" {
		t.Fatalf("answer exchange id = %q, want 999", reply.ExchangeID)
	}
	if got := reply.Contents[0].Ssrc; got != "1000" {
		t.Fatalf("answer audio ssrc = %q, want 1000", got)
	}
}

func TestParseCandidate(t *testing.T) {
	line := "candidate:1 1 udp 2122260223 192.168.1.5 54321 typ host generation 0"
	cand, err := parseCandidate(line)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cand.Address != "192.168.1.5" {
		t.Fatalf("address = %q", cand.Address)
	}
	if cand.Port != 54321 {
		t.Fatalf("port = %d", cand.Port)
	}
}
