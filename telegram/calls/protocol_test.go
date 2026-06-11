package calls

import (
	"strings"
	"testing"

	"github.com/gotd/td/tg"
)

func TestCallProtocol(t *testing.T) {
	p := callProtocol()
	if !p.UDPP2P || !p.UDPReflector {
		t.Fatal("expected p2p and reflector enabled")
	}
	if p.MinLayer != minLayer || p.MaxLayer != maxLayer {
		t.Fatalf("layers = %d/%d", p.MinLayer, p.MaxLayer)
	}
	// Only V2-mapping versions must be advertised.
	for _, v := range p.LibraryVersions {
		if v != "9.0.0" && v != "8.0.0" {
			t.Fatalf("unexpected advertised version %q (risks V3/SCTP)", v)
		}
	}
}

func TestAcceptProtocolIntersectsVersions(t *testing.T) {
	caller := tg.PhoneCallProtocol{
		LibraryVersions: []string{"9.0.0", "7.0.0"},
		MinLayer:        70,
		MaxLayer:        80,
	}
	p := acceptProtocol(caller)
	if len(p.LibraryVersions) != 1 || p.LibraryVersions[0] != "9.0.0" {
		t.Fatalf("intersection = %v, want [9.0.0]", p.LibraryVersions)
	}
	if p.MinLayer != 70 || p.MaxLayer != 80 {
		t.Fatalf("layers not taken from caller: %d/%d", p.MinLayer, p.MaxLayer)
	}

	// No common version → keep our defaults.
	p = acceptProtocol(tg.PhoneCallProtocol{LibraryVersions: []string{"1.0.0"}})
	if len(p.LibraryVersions) == 0 {
		t.Fatal("expected fallback to our versions")
	}
}

func TestIntersect(t *testing.T) {
	got := intersect([]string{"a", "b", "c"}, []string{"c", "a"})
	if len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Fatalf("intersect = %v", got)
	}
	if intersect([]string{"a"}, []string{"b"}) != nil {
		t.Fatal("expected nil for disjoint sets")
	}
}

func TestDiscardReasonTL(t *testing.T) {
	assertType := func(reason DiscardReason, want tg.PhoneCallDiscardReasonClass) {
		if got := reason.tl(); got.TypeID() != want.TypeID() {
			t.Fatalf("reason %d: got %T, want %T", reason, got, want)
		}
	}
	assertType(DiscardHangup, &tg.PhoneCallDiscardReasonHangup{})
	assertType(DiscardBusy, &tg.PhoneCallDiscardReasonBusy{})
	assertType(DiscardMissed, &tg.PhoneCallDiscardReasonMissed{})
	assertType(DiscardDisconnect, &tg.PhoneCallDiscardReasonDisconnect{})
	assertType(DiscardReason(99), &tg.PhoneCallDiscardReasonHangup{}) // default
}

func TestDiscardedError(t *testing.T) {
	for _, tc := range []struct {
		reason tg.PhoneCallDiscardReasonClass
		want   string
	}{
		{&tg.PhoneCallDiscardReasonBusy{}, "busy"},
		{&tg.PhoneCallDiscardReasonMissed{}, "missed"},
		{&tg.PhoneCallDiscardReasonDisconnect{}, "disconnected"},
		{&tg.PhoneCallDiscardReasonHangup{}, "hung up"},
		{&tg.PhoneCallDiscardReasonMigrateConferenceCall{}, "migrated"},
	} {
		err := discardedError(&tg.PhoneCallDiscarded{Reason: tc.reason})
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("reason %T: got %v, want substring %q", tc.reason, err, tc.want)
		}
	}
}
