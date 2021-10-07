// Package messages contains message iteration helper.
package messages

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram/message/peer"
	"github.com/nnqq/td/tg"
)

// Elem is a message iterator element.
type Elem struct {
	Msg      tg.NotEmptyMessage
	Peer     tg.InputPeerClass
	Entities peer.Entities
}

// Iterator is a message stream iterator.
type Iterator struct {
	// Current state.
	lastErr error
	// Buffer state.
	buf    []Elem
	bufCur int
	// Request state.
	addOffset int
	limit     int
	lastBatch bool
	// Offset parameters state.
	offsetID   int
	offsetDate int
	offsetPeer tg.InputPeerClass
	offsetRate int
	// Remote state.
	count    int
	totalGot bool

	// Query builder.
	query Query
}

// NewIterator creates new iterator.
func NewIterator(query Query, limit int) *Iterator {
	return &Iterator{
		buf:        make([]Elem, 0, limit),
		bufCur:     -1,
		limit:      limit,
		query:      query,
		offsetPeer: &tg.InputPeerEmpty{},
	}
}

// OffsetID sets OffsetID request parameter.
func (m *Iterator) OffsetID(offsetID int) *Iterator {
	m.offsetID = offsetID
	return m
}

// OffsetDate sets OffsetDate request parameter.
func (m *Iterator) OffsetDate(offsetDate int) *Iterator {
	m.offsetDate = offsetDate
	return m
}

// OffsetRate sets OffsetRate request parameter.
func (m *Iterator) OffsetRate(offsetRate int) *Iterator {
	m.offsetRate = offsetRate
	return m
}

// OffsetPeer sets OffsetPeer request parameter.
func (m *Iterator) OffsetPeer(offsetPeer tg.InputPeerClass) *Iterator {
	m.offsetPeer = offsetPeer
	return m
}

func (m *Iterator) apply(r tg.MessagesMessagesClass) error {
	if m.lastBatch {
		return nil
	}

	var (
		messages tg.MessageClassArray
		entities peer.Entities
	)
	switch msgs := r.(type) {
	case *tg.MessagesMessages: // messages.messages#8c718e87
		messages = msgs.Messages
		entities = peer.EntitiesFromResult(msgs)

		m.count = len(messages)
		m.lastBatch = true
	case *tg.MessagesMessagesSlice: // messages.messagesSlice#3a54685e
		messages = msgs.Messages
		entities = peer.EntitiesFromResult(msgs)

		m.offsetRate = msgs.NextRate
		m.count = msgs.Count
		m.lastBatch = len(msgs.Messages) < m.limit
	case *tg.MessagesChannelMessages: // messages.channelMessages#64479808
		messages = msgs.Messages
		entities = peer.EntitiesFromResult(msgs)

		m.count = msgs.Count
		m.lastBatch = len(msgs.Messages) < m.limit
	default: // messages.messagesNotModified#74535f21
		return xerrors.Errorf("unexpected type %T", r)
	}
	m.totalGot = true

	// Sort messages to guarantee order and find the last message.
	messages = messages.SortStable(func(a, b tg.MessageClass) bool {
		return a.GetID() > b.GetID()
	})

	// Get the last message (with smallest ID).
	msg, ok := messages.Last()
	if !ok {
		// If Last() returned false, result is empty, so we this is a last batch.
		m.lastBatch = true
		return nil
	}

	// Update offsetID and offsetDate, if can to prevent duplication in case
	// when there a lot new messages in a chat/channel between previous and current request.
	//
	// Illustration of problem:
	//
	//	Remote state:
	//  [10, 9, 8, 7, 6, 5, 4, 3, 2, 1]
	//   ^ offset = 0
	//
	//  First request(offset = 0, limit = 5):
	// 	[10, 9, 8, 7, 6]
	//  offset = 5
	//
	//	Remote state:
	//  [15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1]
	//                       ^ offset = 5
	//
	//  Second request(offset = 5, limit = 5):
	// 	[10, 9, 8, 7, 6]
	//  offset = 10
	//
	m.offsetID = msg.GetID()
	if nonEmpty, ok := msg.AsNotEmpty(); ok {
		m.offsetDate = nonEmpty.GetDate()

		p, err := entities.ExtractPeer(nonEmpty.GetPeerID())
		if err == nil {
			m.offsetPeer = p
		}
	}

	m.bufCur = -1
	m.buf = m.buf[:0]
	for _, msg := range messages {
		nonEmpty, ok := msg.AsNotEmpty()
		if !ok {
			continue
		}

		msgPeer, err := entities.ExtractPeer(nonEmpty.GetPeerID())
		if err != nil {
			msgPeer = &tg.InputPeerEmpty{}
		}

		m.buf = append(m.buf, Elem{
			Msg:      nonEmpty,
			Peer:     msgPeer,
			Entities: entities,
		})
	}

	return nil
}

func (m *Iterator) requestNext(ctx context.Context) error {
	r, err := m.query.Query(ctx, Request{
		OffsetID:   m.offsetID,
		AddOffset:  m.addOffset,
		OffsetDate: m.offsetDate,
		OffsetRate: m.offsetRate,
		OffsetPeer: m.offsetPeer,
		Limit:      m.limit,
	})
	if err != nil {
		return err
	}

	return m.apply(r)
}

func (m *Iterator) bufNext() bool {
	if len(m.buf)-1 <= m.bufCur {
		return false
	}

	m.bufCur++
	return true
}

// Total returns last fetched count of elements.
// If count was not fetched before, it requests server using FetchTotal.
func (m *Iterator) Total(ctx context.Context) (int, error) {
	if m.totalGot {
		return m.count, nil
	}

	return m.FetchTotal(ctx)
}

// FetchTotal fetches and returns count of elements.
func (m *Iterator) FetchTotal(ctx context.Context) (int, error) {
	r, err := m.query.Query(ctx, Request{
		Limit:      1,
		OffsetPeer: &tg.InputPeerEmpty{},
	})
	if err != nil {
		return 0, xerrors.Errorf("fetch total: %w", err)
	}

	switch msgs := r.(type) {
	case *tg.MessagesMessages: // messages.messages#8c718e87
		m.count = len(msgs.Messages)
	case *tg.MessagesMessagesSlice: // messages.messagesSlice#3a54685e
		m.count = msgs.Count
	case *tg.MessagesChannelMessages: // messages.channelMessages#64479808
		m.count = msgs.Count
	default: // messages.messagesNotModified#74535f21
		return 0, xerrors.Errorf("unexpected type %T", r)
	}

	m.totalGot = true
	return m.count, nil
}

// Next prepares the next message for reading with the Value method.
// It returns true on success, or false if there is no next message or an error happened while preparing it.
// Err should be consulted to distinguish between the two cases.
func (m *Iterator) Next(ctx context.Context) bool {
	if m.lastErr != nil {
		return false
	}

	if !m.bufNext() {
		// If buffer is empty, we should fetch next batch.
		if err := m.requestNext(ctx); err != nil {
			m.lastErr = err
			return false
		}
		// Try again with new buffer.
		return m.bufNext()
	}

	return true
}

// Value returns current message.
func (m *Iterator) Value() Elem {
	return m.buf[m.bufCur]
}

// Err returns the error, if any, that was encountered during iteration.
func (m *Iterator) Err() error {
	return m.lastErr
}
