package mtproto

import (
	"time"

	"go.uber.org/atomic"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/rpc"
)

// WireTracer observes the plaintext MTProto traffic and the lifecycle of a
// connection, e.g. to record wire traces for debugging or for comparing the
// client behavior with other implementations.
//
// Callbacks are called synchronously from the connection goroutines, so
// implementations must be safe for concurrent use and must return quickly.
//
// Calls are not ordered by wire order: incoming messages are decrypted
// concurrently, and outgoing messages are reported by concurrent writers
// just before they are written, so a reported outgoing message may never
// reach the wire if the write fails. Order by message id where needed.
//
// The tracer never receives auth key material: only auth key ids.
type WireTracer interface {
	// OnMessage is called for every encrypted message: outgoing messages
	// just before encryption and incoming messages after decryption and
	// msg_key check, before session and msg_id checks (rejected messages
	// are reported too, followed by a WireRejected event).
	//
	// The body is the serialized message body as sent or received
	// (including gzip_packed and msg_container). It is only valid during
	// the call: the tracer must not modify or retain it, and must copy it
	// if needed.
	OnMessage(m WireMessage, body []byte)
	// OnEvent is called on connection, key, session and request lifecycle
	// events.
	OnEvent(e WireEvent)
}

// WireDirection is the direction of a traced message.
type WireDirection uint8

// Message directions.
const (
	// WireOut is a message sent by the client.
	WireOut WireDirection = iota + 1
	// WireIn is a message received from the server.
	WireIn
)

// String implements fmt.Stringer.
func (d WireDirection) String() string {
	switch d {
	case WireOut:
		return "out"
	case WireIn:
		return "in"
	default:
		return "unknown"
	}
}

// WireConn identifies the connection of a traced message or event.
type WireConn struct {
	// ID is unique for every Conn of the process, starting from 1.
	ID int64
	// DC is the DC id of the connection used for key exchange
	// (Options.DC): 10000 is added for test DCs.
	DC int
}

// WireMessage is the header of a traced message.
type WireMessage struct {
	Conn      WireConn
	Direction WireDirection
	// AuthKeyID is the id of the auth key the message is encrypted with.
	AuthKeyID [8]byte
	// TempKey is true if the message is encrypted with a temporary
	// (PFS) key.
	TempKey   bool
	Salt      int64
	SessionID int64
	MessageID int64
	SeqNo     int32
}

// WireEventType is the type of a WireEvent.
type WireEventType uint8

// Event types.
const (
	// WireConnStart is reported before dialing.
	WireConnStart WireEventType = iota + 1
	// WireConnUp is reported when the transport connection is established.
	WireConnUp
	// WireConnDown is reported when the connection is closed, Err is the
	// reason (nil on graceful close).
	WireConnDown
	// WireKeyCreated is reported when a new auth key is created by key
	// exchange: AuthKeyID, TempKey and ExpiresAt (temporary keys).
	WireKeyCreated
	// WireKeySet is reported when the connection starts using an auth key
	// for messages, including after the key is re-created because the server
	// did not find it: AuthKeyID, TempKey and ExpiresAt (temporary keys).
	WireKeySet
	// WireSessionNew is reported when the client session id changes:
	// Old and New.
	WireSessionNew
	// WireSaltChange is reported when the server salt changes: Old and New.
	// Salts set by key exchange are implied by WireKeyCreated.
	WireSaltChange
	// WireDecryptFailed is reported when an incoming message can't be
	// decrypted or fails the msg_key check: Err.
	WireDecryptFailed
	// WireRejected is reported when a decrypted incoming message is
	// rejected: MessageID and Reason ("bad_session" or "bad_msg_id").
	WireRejected
	// WireRequest is reported when a request is going to be sent:
	// MessageID, SeqNo and TypeID (of the encoded request, e.g. a
	// wrapper like invokeWithLayer).
	WireRequest
	// WireRequestResend is reported when a request is re-encoded and sent
	// again after bad_server_salt: MessageID and Reason ("bad_server_salt").
	// Retransmits after an ack timeout reuse the message id and are only
	// visible as repeated outgoing messages.
	WireRequestResend
	// WireRequestDone is reported when a request is finished: MessageID,
	// Err (nil on success, RPC or context error otherwise) and Latency.
	WireRequestDone
)

var wireTypeNames = [...]string{
	WireConnStart:     "conn_start",
	WireConnUp:        "conn_up",
	WireConnDown:      "conn_down",
	WireKeyCreated:    "key_created",
	WireKeySet:        "key_set",
	WireSessionNew:    "session_new",
	WireSaltChange:    "salt_change",
	WireDecryptFailed: "decrypt_failed",
	WireRejected:      "rejected",
	WireRequest:       "request",
	WireRequestResend: "request_resend",
	WireRequestDone:   "request_done",
}

// String implements fmt.Stringer.
func (t WireEventType) String() string {
	if int(t) < len(wireTypeNames) && wireTypeNames[t] != "" {
		return wireTypeNames[t]
	}
	return "unknown"
}

// WireEvent is a connection lifecycle event. Fields are set depending on
// Type.
type WireEvent struct {
	Type WireEventType
	Conn WireConn

	// AuthKeyID, TempKey and ExpiresAt (unix seconds, temporary keys only)
	// describe the auth key of key events.
	AuthKeyID [8]byte
	TempKey   bool
	ExpiresAt int64

	// Old and New are session ids or salts.
	Old, New int64

	MessageID int64
	SeqNo     int32
	TypeID    uint32
	Latency   time.Duration

	Reason string
	Err    error
}

var wireConnID atomic.Int64

func (c *Conn) wireConn() WireConn {
	return WireConn{ID: c.wireID, DC: c.dcID}
}

// traceEvent reports e if tracing is enabled. Callers on hot paths should
// check c.wire before building the event.
func (c *Conn) traceEvent(e WireEvent) {
	if c.wire == nil {
		return
	}
	e.Conn = c.wireConn()
	c.wire.OnEvent(e)
}

func (c *Conn) traceMessage(dir WireDirection, key crypto.AuthKey, salt, session, msgID int64, seqNo int32, body []byte) {
	c.wire.OnMessage(WireMessage{
		Conn:      c.wireConn(),
		Direction: dir,
		AuthKeyID: key.ID,
		TempKey:   c.pfs,
		Salt:      salt,
		SessionID: session,
		MessageID: msgID,
		SeqNo:     seqNo,
	}, body)
}

// traceKey reports a key event for the current auth key.
func (c *Conn) traceKey(t WireEventType, key crypto.AuthKey, temp bool) {
	if c.wire == nil {
		return
	}
	var expiresAt int64
	if temp {
		c.sessionMux.RLock()
		expiresAt = c.tempKeyExpiry
		c.sessionMux.RUnlock()
	}
	c.traceEvent(WireEvent{Type: t, AuthKeyID: key.ID, TempKey: temp, ExpiresAt: expiresAt})
}

func (c *Conn) traceSession(old, current int64) {
	if c.wire == nil || old == current {
		return
	}
	c.traceEvent(WireEvent{Type: WireSessionNew, Old: old, New: current})
}

// traceRequestStart reports a request that is going to be sent and returns
// the start time for traceRequestDone.
func (c *Conn) traceRequestStart(req rpc.Request) time.Time {
	start := c.clock.Now()
	c.traceEvent(WireEvent{Type: WireRequest, MessageID: req.MsgID, SeqNo: req.SeqNo, TypeID: peekID(req.Input)})
	return start
}

// traceRequestDone reports the result of a request.
func (c *Conn) traceRequestDone(req rpc.Request, start time.Time, err error) {
	c.traceEvent(WireEvent{Type: WireRequestDone, MessageID: req.MsgID, Latency: c.clock.Now().Sub(start), Err: err})
}

// traceResend reports a request sent again with the same message id.
func (c *Conn) traceResend(req rpc.Request, reason string) {
	c.traceEvent(WireEvent{Type: WireRequestResend, MessageID: req.MsgID, SeqNo: req.SeqNo, Reason: reason})
}

// traceRejected reports a rejected incoming message.
func (c *Conn) traceRejected(msgID int64, reason string) {
	c.traceEvent(WireEvent{Type: WireRejected, MessageID: msgID, Reason: reason})
}
