package proto

import (
	"fmt"
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
	const nano = 1e9
	// Must approximately equal unixtime*2^32.

	// Important: to counter replay-attacks the lower 32 bits of msg_id
	// passed by the client must not be empty and must present a
	// fractional part of the time point when the message was created.
	intPart := nowNano / nano
	fracPart := nowNano % nano

	// Ensure that fracPart % 4 == 0.
	fracPart &= -messageIDModulo
	// Adding modulo 4 yield to ensure message type.
	fracPart += int64(yield)

	return (intPart << 32) | fracPart
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
	fracPart := int64(int32(id))
	return time.Unix(intPart, fracPart).UTC()
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
	var yield int
	switch typ {
	case MessageFromClient:
		yield = yieldClient
	case MessageFromServer:
		yield = yieldFromServer
	case MessageServerResponse:
		yield = yieldServerResponse
	default:
		yield = yieldClient
	}
	return MessageID(newMessageID(nano, yield))
}

// MessageIDGen is message id generator that provides collision prevention.
//
// The main reason of such structure is that now() can return same time during
// multiple calls and that leads to duplicate message id.
type MessageIDGen struct {
	mux  sync.Mutex
	nano int64
	now  func() time.Time
}

// New generates new message id for provided type, protecting from collisions
// that are caused by low system time resolution.
func (g *MessageIDGen) New(t MessageType) int64 {
	g.mux.Lock()
	defer g.mux.Unlock()

	// Minimum resolution is required because id is only approximately
	// equal to unix nano time, some part is replaced by message type.
	const minResolutionNanos = 10

	nano := g.now().UnixNano()
	if nano > g.nano {
		g.nano = nano
	} else {
		g.nano += minResolutionNanos
	}

	return int64(NewMessageIDNano(g.nano, t))
}

// NewMessageIDGen creates new message id generator.
//
// Current time will be provided by now() function.
//
// This generator compensates time resolution problem removing
// probability of id collision.
//
// Such problem can be observed for relatively high RPS, sequential calls to
// time.Now() will return same time which leads to equal ids.
func NewMessageIDGen(now func() time.Time) *MessageIDGen {
	return &MessageIDGen{
		now: now,
	}
}

// MessageIDBuf stores the last N received message ids and is used in replay
// attack mitigation.
//
// It works like TDLib's MessageIdDuplicateChecker: ids are kept sorted in a
// buffer of 2N, so an id is checked by binary search, an increasing id (the
// common case) is appended, and the oldest N are dropped at once when the
// buffer is full.
type MessageIDBuf struct {
	mux sync.Mutex
	// buf[:end] holds the stored ids in increasing order.
	buf []int64
	end int
	n   int
}

// NewMessageIDBuf initializes new message id buffer for last N stored values.
func NewMessageIDBuf(n int) *MessageIDBuf {
	if n < 1 {
		n = 1
	}
	return &MessageIDBuf{
		buf: make([]int64, 2*n),
		n:   n,
	}
}

// Reset forgets all stored message ids, e.g. when a new session is started.
func (b *MessageIDBuf) Reset() {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.end = 0
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
	if newID <= 0 {
		return false
	}

	b.mux.Lock()
	defer b.mux.Unlock()

	if b.end == len(b.buf) {
		// Forget the oldest N ids.
		copy(b.buf[:b.n], b.buf[b.n:])
		b.end = b.n
	}
	if b.end == 0 || newID > b.buf[b.end-1] {
		// Fast path: ids mostly increase.
		b.buf[b.end] = newID
		b.end++
		return true
	}
	if b.end >= b.n && newID < b.buf[0] {
		// Lower than all of the last N stored ids.
		return false
	}

	// Binary search for the first stored id >= newID.
	lo, hi := 0, b.end
	for lo < hi {
		mid := int(uint(lo+hi) >> 1)
		if b.buf[mid] < newID {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if b.buf[lo] == newID {
		// Equal to a stored id.
		return false
	}
	copy(b.buf[lo+1:b.end+1], b.buf[lo:b.end])
	b.buf[lo] = newID
	b.end++
	return true
}
