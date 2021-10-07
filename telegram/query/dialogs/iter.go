// Package dialogs contains dialog iteration helper.
package dialogs

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram/message/peer"
	"github.com/nnqq/td/tg"
)

// Elem is a dialog iterator element.
type Elem struct {
	Dialog   tg.DialogClass
	Peer     tg.InputPeerClass
	Last     tg.NotEmptyMessage
	Entities peer.Entities
}

// Iterator is a dialog stream iterator.
type Iterator struct {
	// Current state.
	lastErr error
	// Buffer state.
	buf    []Elem
	bufCur int
	// Request state.
	limit     int
	lastBatch bool
	// Offset parameters state.
	offsetID   int
	offsetDate int
	offsetPeer tg.InputPeerClass
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

// OffsetPeer sets OffsetPeer request parameter.
func (m *Iterator) OffsetPeer(offsetPeer tg.InputPeerClass) *Iterator {
	m.offsetPeer = offsetPeer
	return m
}

// messageMap is a helper to store messages for multiple peers.
type messageMap map[peer.DialogKey]tg.NotEmptyMessage

func (m messageMap) collect(messages tg.MessageClassArray) error {
	for _, msg := range messages {
		nonEmpty, ok := msg.AsNotEmpty()
		if !ok {
			// TODO(tdakkota): Maybe should I return error here?
			continue
		}

		var key peer.DialogKey
		if err := key.FromPeer(nonEmpty.GetPeerID()); err != nil {
			return err
		}

		m[key] = nonEmpty
	}

	return nil
}

func (m *Iterator) apply(r tg.MessagesDialogsClass) error {
	if m.lastBatch {
		return nil
	}

	var (
		messages tg.MessageClassArray
		dialogs  tg.DialogClassArray
		entities peer.Entities
	)

	switch dlgs := r.(type) {
	case *tg.MessagesDialogs: // messages.dialogs#15ba6c40
		dialogs = dlgs.Dialogs
		messages = dlgs.Messages
		entities = peer.EntitiesFromResult(dlgs)

		m.count = len(messages)
		m.lastBatch = true
	case *tg.MessagesDialogsSlice: // messages.dialogsSlice#71e094f3
		dialogs = dlgs.Dialogs
		messages = dlgs.Messages
		entities = peer.EntitiesFromResult(dlgs)

		m.count = dlgs.Count
		m.lastBatch = len(dlgs.Dialogs) < m.limit
	default: // messages.dialogsNotModified#f0e3e596
		return xerrors.Errorf("unexpected type %T", r)
	}
	m.totalGot = true

	msgMap := make(messageMap, len(messages))
	if err := msgMap.collect(messages); err != nil {
		return xerrors.Errorf("collect last messages: %w", err)
	}

	m.bufCur = -1
	m.buf = m.buf[:0]

	var last tg.NotEmptyMessage
	for _, dlg := range dialogs {
		var key peer.DialogKey
		if err := key.FromPeer(dlg.GetPeer()); err == nil {
			last = msgMap[key]
		}

		p, err := entities.ExtractPeer(dlg.GetPeer())
		if err != nil {
			p = &tg.InputPeerEmpty{}
		}

		m.buf = append(m.buf, Elem{
			Dialog:   dlg,
			Peer:     p,
			Last:     last,
			Entities: entities,
		})
	}

	if !m.lastBatch && len(m.buf) > 0 {
		if last != nil {
			m.offsetID = last.GetID()
			m.offsetDate = last.GetDate()
		}

		p, err := entities.ExtractPeer(dialogs[len(m.buf)-1].GetPeer())
		if err != nil {
			return xerrors.Errorf("get offset peer: %w", err)
		}
		m.offsetPeer = p
	}

	return nil
}

func (m *Iterator) requestNext(ctx context.Context) error {
	r, err := m.query.Query(ctx, Request{
		OffsetID:   m.offsetID,
		OffsetDate: m.offsetDate,
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

	switch dlgs := r.(type) {
	case *tg.MessagesDialogs: // messages.dialogs#15ba6c40
		m.count = len(dlgs.Dialogs)
	case *tg.MessagesDialogsSlice: // messages.dialogsSlice#71e094f3
		m.count = dlgs.Count
	default: // messages.dialogsNotModified#f0e3e596
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
