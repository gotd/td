package calls

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// Library versions we advertise in phone.requestCall/acceptCall. The server
// intersects these with the peer's and returns the best common value, which
// both sides feed to tgcalls' signalingProtocolVersion().
//
// We must pin a version that selects the "V2" signaling protocol — sequenced,
// uncompressed, sent directly over phone.sendSignalingData. "8.0.0" and "9.0.0"
// map to V2; newer strings (e.g. "12.0.0"+) select "V3", which gzips messages
// and tunnels them over SCTP, a framing this package does not implement.
// Advertising unknown/high versions risks the peer choosing V3 and producing
// packets we cannot decrypt (the bytes are SCTP, not a raw encrypted packet).
var libraryVersions = []string{"9.0.0", "8.0.0"}

const (
	minLayer = 65
	maxLayer = 92
)

// callProtocol builds the protocol descriptor advertised by the caller.
func callProtocol() tg.PhoneCallProtocol {
	return tg.PhoneCallProtocol{
		UDPP2P:          true,
		UDPReflector:    true,
		MinLayer:        minLayer,
		MaxLayer:        maxLayer,
		LibraryVersions: libraryVersions,
	}
}

// acceptProtocol builds the callee's protocol descriptor, narrowing the
// advertised library versions to those also offered by the caller.
func acceptProtocol(caller tg.PhoneCallProtocol) tg.PhoneCallProtocol {
	p := callProtocol()
	if shared := intersect(caller.LibraryVersions, p.LibraryVersions); len(shared) > 0 {
		p.LibraryVersions = shared
	}
	if caller.MinLayer != 0 {
		p.MinLayer = caller.MinLayer
	}
	if caller.MaxLayer != 0 {
		p.MaxLayer = caller.MaxLayer
	}
	return p
}

// intersect returns the elements of a that also appear in b, preserving a's order.
func intersect(a, b []string) []string {
	set := make(map[string]struct{}, len(b))
	for _, v := range b {
		set[v] = struct{}{}
	}
	var out []string
	for _, v := range a {
		if _, ok := set[v]; ok {
			out = append(out, v)
		}
	}
	return out
}

// DiscardReason describes why a call ended, mapping to the
// phoneCallDiscardReason* constructors.
type DiscardReason int

const (
	// DiscardHangup means the call was ended normally by a participant.
	DiscardHangup DiscardReason = iota
	// DiscardBusy means the callee was busy.
	DiscardBusy
	// DiscardMissed means the call was not answered.
	DiscardMissed
	// DiscardDisconnect means the connection was lost.
	DiscardDisconnect
)

func (r DiscardReason) tl() tg.PhoneCallDiscardReasonClass {
	switch r {
	case DiscardBusy:
		return &tg.PhoneCallDiscardReasonBusy{}
	case DiscardMissed:
		return &tg.PhoneCallDiscardReasonMissed{}
	case DiscardDisconnect:
		return &tg.PhoneCallDiscardReasonDisconnect{}
	case DiscardHangup:
		return &tg.PhoneCallDiscardReasonHangup{}
	default:
		return &tg.PhoneCallDiscardReasonHangup{}
	}
}

// discardedError turns a phoneCallDiscarded update into an error.
func discardedError(d *tg.PhoneCallDiscarded) error {
	reason := "ended"
	switch d.Reason.(type) {
	case *tg.PhoneCallDiscardReasonBusy:
		reason = "busy"
	case *tg.PhoneCallDiscardReasonMissed:
		reason = "missed"
	case *tg.PhoneCallDiscardReasonDisconnect:
		reason = "disconnected"
	case *tg.PhoneCallDiscardReasonHangup:
		reason = "hung up"
	case *tg.PhoneCallDiscardReasonMigrateConferenceCall:
		reason = "migrated to conference call"
	}
	return errors.Errorf("call discarded: %s", reason)
}
