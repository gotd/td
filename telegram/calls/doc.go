// Package calls implements Telegram 1:1 phone calls (the tgcalls "v2"
// protocol) on top of the gotd/td MTProto client.
//
// A phone call has two halves:
//
//   - Signaling, carried over MTProto: the Diffie-Hellman key exchange
//     (phone.requestCall / acceptCall / confirmCall), reflector/STUN/TURN
//     endpoint discovery and the encrypted control channel
//     (phone.sendSignalingData / updatePhoneCallSignalingData).
//   - Media, carried over a direct peer-to-peer ICE/DTLS/SRTP connection
//     negotiated through the signaling channel.
//
// This package owns all of the signaling half and drives a
// [github.com/pion/webrtc/v4] transport for the media half. The media engine
// (audio/video encoding) is intentionally left to the caller: outgoing media
// is written as RTP to the local tracks exposed by [Conn.AudioTrack] and
// [Conn.VideoTrack], and incoming media is delivered as
// [github.com/pion/webrtc/v4.TrackRemote] values through [Conn.OnTrack].
//
// # Connectivity
//
// The transport connects over direct ICE candidates (host and STUN/server-
// reflexive). Telegram also offers relay candidates served by its custom
// "reflector" protocol (tgcalls ReflectorPort), which is not RFC TURN; those
// are not implemented here and are skipped. Calls therefore require a usable
// direct path between the peers — they will not fall back to relay when both
// sides are behind restrictive NATs.
//
// # Signaling protocol
//
// This package speaks the tgcalls "V2" signaling protocol (sequenced,
// uncompressed). It advertises only the library versions that select V2; it
// does not implement "V3" (gzip + SCTP-tunneled signaling).
//
// # Usage
//
// Create a [Client] bound to a *tg.Client invoker, register it on the update
// dispatcher, then place or answer calls:
//
//	calls := calls.NewClient(tg.NewClient(invoker), calls.Options{})
//	calls.Register(dispatcher)
//
//	calls.OnIncoming(func(in *calls.IncomingCall) {
//		conn, err := in.Accept(ctx)
//		// ... use conn ...
//	})
//
//	conn, err := calls.Request(ctx, inputUser)
//
// The returned [Conn] reports connectivity through OnConnected/OnDisconnected
// and exposes the negotiated media tracks.
//
// White-box note: the wire formats (DH key derivation, the signaling
// EncryptedConnection AES-CTR scheme and the JSON control messages) are
// reimplemented from the official tgcalls reference and verified against it;
// no third-party Go implementation is reused.
package calls
