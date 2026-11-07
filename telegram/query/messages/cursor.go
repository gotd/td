package messages

import "github.com/gotd/td/tg"

// Cursor is a position of Iterator, which can be saved to resume iteration
// later, e.g. after restart, from the element after the last returned one.
type Cursor struct {
	OffsetID   int
	OffsetDate int
	OffsetRate int
	// OffsetPeer is used by global search. Nil means inputPeerEmpty.
	OffsetPeer tg.InputPeerClass
}

// Cursor returns position after the last element returned by Value, or the
// initial offsets if no element was returned yet.
//
// OffsetRate (global search) is the date of the element, as tdlib does, or
// next_rate returned by the server for the last element of a batch.
func (m *Iterator) Cursor() Cursor {
	if m.bufCur < 0 || m.bufCur >= len(m.buf) {
		return Cursor{
			OffsetID:   m.offsetID,
			OffsetDate: m.offsetDate,
			OffsetRate: m.offsetRate,
			OffsetPeer: m.offsetPeer,
		}
	}
	elem := m.buf[m.bufCur]
	rate := elem.Msg.GetDate()
	if m.bufCur == len(m.buf)-1 && m.offsetRate > 0 {
		// Same offset as the next request.
		rate = m.offsetRate
	}
	return Cursor{
		OffsetID:   elem.Msg.GetID(),
		OffsetDate: elem.Msg.GetDate(),
		OffsetRate: rate,
		OffsetPeer: elem.Entities.OffsetPeer(elem.Msg.GetPeerID()),
	}
}

// Resume sets offsets from cursor, to continue iteration from it.
func (m *Iterator) Resume(c Cursor) *Iterator {
	m.offsetID = c.OffsetID
	m.offsetDate = c.OffsetDate
	m.offsetRate = c.OffsetRate
	m.offsetPeer = c.OffsetPeer
	if m.offsetPeer == nil {
		m.offsetPeer = &tg.InputPeerEmpty{}
	}
	return m
}
