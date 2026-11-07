package dialogs

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// Cursor is a position of Iterator, which can be saved to resume iteration
// later, e.g. after restart, from the dialog after the last returned one.
type Cursor struct {
	OffsetID   int
	OffsetDate int
	// OffsetPeer is the peer of the last dialog. Nil means inputPeerEmpty.
	OffsetPeer tg.InputPeerClass
}

// Cursor returns position after the last element returned by Value, or the
// initial offsets if no element was returned yet.
func (m *Iterator) Cursor() Cursor {
	if m.bufCur < 0 || m.bufCur >= len(m.buf) {
		return Cursor{
			OffsetID:   m.offsetID,
			OffsetDate: m.offsetDate,
			OffsetPeer: m.offsetPeer,
		}
	}
	elem := m.buf[m.bufCur]
	c := Cursor{OffsetPeer: elem.Peer}
	if elem.Last != nil {
		c.OffsetID = elem.Last.GetID()
		c.OffsetDate = elem.Last.GetDate()
	}
	return c
}

// Resume sets offsets from cursor, to continue iteration from it.
func (m *Iterator) Resume(c Cursor) *Iterator {
	m.offsetID = c.OffsetID
	m.offsetDate = c.OffsetDate
	m.offsetPeer = c.OffsetPeer
	if m.offsetPeer == nil {
		m.offsetPeer = &tg.InputPeerEmpty{}
	}
	return m
}

// Folders are dialog folders: the main list and the archive.
var Folders = []int{0, 1}

// ForEachFolder calls cb on each dialog of given folders, in order (all
// folders, i.e. the main list and the archive, if none given), with given
// batch size.
func (q *QueryBuilder) ForEachFolder(
	ctx context.Context,
	batchSize int,
	cb func(ctx context.Context, folderID int, e Elem) error,
	folders ...int,
) error {
	if len(folders) == 0 {
		folders = Folders
	}
	for _, folder := range folders {
		if err := q.GetDialogs().FolderID(folder).BatchSize(batchSize).ForEach(ctx, func(ctx context.Context, e Elem) error {
			return cb(ctx, folder, e)
		}); err != nil {
			return errors.Wrapf(err, "folder %d", folder)
		}
	}
	return nil
}
