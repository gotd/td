package proto

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

// Message identifiers are coupled to message creation time.
//
// https://core.telegram.org/mtproto/description#message-identifier-msg-id

const (
	yieldClient         = 0
	yieldServerResponse = 1
	yieldFromServer     = 3

	messageIDModulo = 4
)

func newMessageID(nowNano int64, yield int) int64 {
	return newMessageIDJitter(nowNano, yield, 0)
}

// newMessageIDJitter returns a message id for the time point nowNano: unix
// seconds in the upper 32 bits and the fraction of the second in units of
// 2^-32 s in the lower 32 bits (msg_id approximately equals unixtime * 2^32),
// with the lower bits XORed with jitter (TDLib XORs 22 random bits, about
// 1 ms, for clocks of low precision) and the two lowest bits set to yield.
func newMessageIDJitter(nowNano int64, yield int, jitter uint32) int64 {
	const nano = 1e9

	// Important: to counter replay-attacks the lower 32 bits of msg_id
	// passed by the client must not be empty and must present a
	// fractional part of the time point when the message was created.
	intPart := nowNano / nano
	frac := uint64(nowNano%nano) << 32 / nano
	frac ^= uint64(jitter)

	// Ensure that frac % 4 == 0 and add modulo 4 yield to ensure message
	// type.
	frac = frac&^(messageIDModulo-1) | uint64(yield)

	return intPart<<32 | int64(frac)
}

// MessageID represents 64-bit message id.
type MessageID int64

func (id MessageID) String() string {
	return fmt.Sprintf("%x (%s, %s)",
		int64(id), id.Type(), id.Time().Format(time.RFC3339),
	)
}

// MessageType is type of message determined by message id.
//
// A message is rejected over 300 seconds after it is created or
// 30 seconds before it is created (this is needed to protect from replay attacks).
//
// The identifier of a message container must be strictly greater than those of
// its nested messages.
type MessageType byte

const (
	// MessageUnknown reports that message id has unknown time and probably
	// should be ignored.
	MessageUnknown MessageType = iota
	// MessageFromClient is client message identifiers.
	MessageFromClient
	// MessageServerResponse is a response to a client message.
	MessageServerResponse
	// MessageFromServer is a message from the server.
	MessageFromServer
)

func (m MessageType) String() string {
	switch m {
	case MessageFromClient:
		return "FromClient"
	case MessageServerResponse:
		return "ServerResponse"
	case MessageFromServer:
		return "FromServer"
	default:
		return "Unknown"
	}
}

// Time returns approximate time when MessageID were generated.
func (id MessageID) Time() time.Time {
	intPart := int64(id) >> 32
	// The lower 32 bits are the fraction of the second in units of 2^-32 s.
	nanos := int64(uint64(uint32(id)) * 1e9 >> 32)
	return time.Unix(intPart, nanos).UTC()
}

// Type returns message type.
func (id MessageID) Type() MessageType {
	switch id % messageIDModulo {
	case yieldClient:
		return MessageFromClient
	case yieldServerResponse:
		return MessageServerResponse
	case yieldFromServer:
		return MessageFromServer
	default:
		return MessageUnknown
	}
}

// NewMessageID returns new message id for provided time and type.
func NewMessageID(now time.Time, typ MessageType) MessageID {
	return NewMessageIDNano(now.UnixNano(), typ)
}

// NewMessageIDNano returns new message id for provided current unix
// nanoseconds and type.
func NewMessageIDNano(nano int64, typ MessageType) MessageID {
	return MessageID(newMessageID(nano, messageYield(typ)))
}

// messageYield returns the two lowest bits of message ids of type typ.
func messageYield(typ MessageType) int {
	switch typ {
	case MessageFromServer:
		return yieldFromServer
	case MessageServerResponse:
		return yieldServerResponse
	default:
		return yieldClient
	}
}

// MessageIDGen is message id generator that provides collision prevention.
//
// Ids follow TDLib (AuthData::next_message_id): the time of now() with 22
// random bits XORed into the fraction, and if that is not above the last id
// (now() returned the same or an earlier time, e.g. after a correction of the
// server time offset), the last id plus a random multiple of 4. So ids always
// increase, and their lower bits don't reveal the clock resolution.
type MessageIDGen struct {
	mux  sync.Mutex
	last int64
	now  func() time.Time
}

// jitterBits is the number of random lower bits of generated ids (TDLib).
const jitterBits = 22

// New generates new message id for provided type, protecting from collisions
// that are caused by low system time resolution.
func (g *MessageIDGen) New(t MessageType) int64 {
	g.mux.Lock()
	defer g.mux.Unlock()

	r := rand.Uint32() // #nosec G404: jitter, not a secret
	id := newMessageIDJitter(g.now().UnixNano(), messageYield(t), r&(1<<jitterBits-1))
	if id <= g.last {
		step := int64(r>>jitterBits&1023) + 1
		id = (g.last&^(messageIDModulo-1) + messageIDModulo*step) | int64(messageYield(t))
	}
	g.last = id
	return id
}

// NewMessageIDGen creates new message id generator.
//
// Current time will be provided by now() function, e.g. the server time of
// mtproto.ServerTime.
func NewMessageIDGen(now func() time.Time) *MessageIDGen {
	return &MessageIDGen{
		now: now,
	}
}

// MessageIDBuf stores last N message ids and is used in replay attack mitigation.
type MessageIDBuf struct {
	mux sync.Mutex
	buf []int64
}

// NewMessageIDBuf initializes new message id buffer for last N stored values.
func NewMessageIDBuf(n int) *MessageIDBuf {
	return &MessageIDBuf{
		buf: make([]int64, n),
	}
}

// Reset forgets all stored message ids, e.g. when a new session is started.
func (b *MessageIDBuf) Reset() {
	b.mux.Lock()
	defer b.mux.Unlock()

	clear(b.buf)
}

// Consume returns false if message should be discarded.
func (b *MessageIDBuf) Consume(newID int64) bool {
	// In addition, the identifiers (msg_id) of the last N messages received
	// from the other side must be stored, and if a message comes in with an
	// msg_id lower than all or equal to any of the stored values, that message
	// is to be ignored. Otherwise, the new message msg_id is added to the set,
	// and, if the number of stored msg_id values is greater than N, the oldest
	// (i. e. the lowest) is discarded.
	//
	// https://core.telegram.org/mtproto/security_guidelines#checking-msg-id

	b.mux.Lock()
	defer b.mux.Unlock()

	var (
		minIDx int
		minID  int64
	)
	for i, id := range b.buf {
		if id == newID {
			// Equal to stored value.
			return false
		}
		// Searching for minimum value.
		if id < minID {
			minIDx = i
			minID = id
		}
	}
	if newID < minID {
		// Lower than all stored values.
		return false
	}

	// Message is accepted. Replacing lowest message id with new id.
	b.buf[minIDx] = newID
	return true
}
