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

// TestNegotiatorOfferAnswer checks that the answer to our own offer is not
// replied to and — crucially — is not mistaken for the peer's SSRCs.
//
// The answer echoes the OFFER's contents, i.e. our own SSRCs (verified against
// the official Android client: answering our offer of ssrc 1893540410 it
// replied with ssrc 1893540410). Reading those back as "the peer's" binds the
// receiver to our own SSRC and silently drops all inbound media.
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
			{Type: "audio", Ssrc: "1000"},
			{Type: "video", Ssrc: "1001"},
		},
	}
	reply, ready := n.applyRemoteChannels(answer, 1000, 1001)
	if reply != nil {
		t.Fatal("answer to our own offer should not be replied to")
	}
	if ready {
		t.Fatal("not ready yet: the peer has not offered its own SSRCs")
	}
	if n.peerAudioSSRC() != 0 || n.peerVideoSSRC() != 0 {
		t.Fatalf("peer ssrcs = %d/%d, want 0/0 — the answer mirrors our own",
			n.peerAudioSSRC(), n.peerVideoSSRC())
	}
}

// TestNegotiatorPeerOfferAfterOwnRound checks the sequence a real call actually
// produces: our offer is answered, and only then does the peer make its own
// offer carrying its real SSRCs. That late offer must still be processed and
// answered — dropping it leaves the peer unanswered and it never sends media.
func TestNegotiatorPeerOfferAfterOwnRound(t *testing.T) {
	n := newContentNegotiation()
	offer := n.proposeChannels(1000, 1001)

	// Peer answers our offer, mirroring our SSRCs.
	n.applyRemoteChannels(&negotiateChannelsMessage{
		ExchangeID: offer.ExchangeID,
		Contents:   []mediaContent{{Type: "audio", Ssrc: "1000"}},
	}, 1000, 1001)

	// Peer now offers its own audio channel (audio-only call: no video content).
	reply, ready := n.applyRemoteChannels(&negotiateChannelsMessage{
		ExchangeID: "2212439328",
		Contents:   []mediaContent{{Type: "audio", Ssrc: "1864316852"}},
	}, 1000, 1001)
	if reply == nil {
		t.Fatal("the peer's own offer must be answered")
	}
	if !ready {
		t.Fatal("negotiation should be ready once the peer's SSRC is known")
	}
	if n.peerAudioSSRC() != 1864316852 {
		t.Fatalf("peer audio ssrc = %d, want 1864316852", n.peerAudioSSRC())
	}
}

// TestNegotiatorRemoteOffer checks that a peer offer with a different exchange
// ID produces an answer echoing the OFFERED contents, which is what the
// reference implementation does and what the official client expects.
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
	if got := reply.Contents[0].Ssrc; got != "5000" {
		t.Fatalf("answer audio ssrc = %q, want 5000 (the offer is echoed back)", got)
	}
	if n.peerAudioSSRC() != 5000 {
		t.Fatalf("peer audio ssrc = %d, want 5000", n.peerAudioSSRC())
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
