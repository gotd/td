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
//
// Pinned dialogs come first, out of date order, and dialogs without a last
// message or peer have no position. For such elements the cursor is the
// position of the last preceding regular dialog of the batch, or the offsets
// the batch was requested with, so resuming may return some dialogs again
// but never skips any.
func (m *Iterator) Cursor() Cursor {
	if m.bufCur < 0 || m.bufCur >= len(m.buf) {
		return Cursor{
			OffsetID:   m.offsetID,
			OffsetDate: m.offsetDate,
			OffsetPeer: m.offsetPeer,
		}
	}
	for i := m.bufCur; i >= 0; i-- {
		elem := m.buf[i]
		if !elem.hasPosition() {
			continue
		}
		return Cursor{
			OffsetID:   elem.Last.GetID(),
			OffsetDate: elem.Last.GetDate(),
			OffsetPeer: elem.Peer,
		}
	}
	return Cursor{
		OffsetID:   m.batchOffsetID,
		OffsetDate: m.batchOffsetDate,
		OffsetPeer: m.batchOffsetPeer,
	}
}

// hasPosition reports whether element can be used as iteration offset.
func (e Elem) hasPosition() bool {
	if e.Last == nil || e.Peer == nil {
		return false
	}
	if _, ok := e.Peer.(*tg.InputPeerEmpty); ok {
		return false
	}
	if p, ok := e.Dialog.(interface{ GetPinned() bool }); ok && p.GetPinned() {
		return false
	}
	return true
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

// folderQuery requests dialogs of one folder.
//
// Unlike GetDialogs().FolderID(0), it always sends folder_id: without it
// the server returns dialogs of all folders.
type folderQuery struct {
	raw    *tg.Client
	folder int
}

func (q folderQuery) Query(ctx context.Context, req Request) (tg.MessagesDialogsClass, error) {
	r := &tg.MessagesGetDialogsRequest{
		OffsetDate: req.OffsetDate,
		OffsetID:   req.OffsetID,
		OffsetPeer: req.OffsetPeer,
		Limit:      req.Limit,
	}
	r.SetFolderID(q.folder)
	return q.raw.MessagesGetDialogs(ctx, r)
}

// ForEachFolder calls cb on each dialog of given folders, in order (all
// folders, i.e. the main list and the archive, if none given), with given
// batch size.
//
// Folder entries (dialogFolder) are skipped: their dialogs are returned
// with their own folder.
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
		iter := NewIterator(folderQuery{raw: q.raw, folder: folder}, batchSize)
		for iter.Next(ctx) {
			e := iter.Value()
			if _, ok := e.Dialog.(*tg.DialogFolder); ok {
				continue
			}
			if err := cb(ctx, folder, e); err != nil {
				return errors.Wrapf(err, "folder %d", folder)
			}
		}
		if err := iter.Err(); err != nil {
			return errors.Wrapf(err, "folder %d", folder)
		}
	}
	return nil
}
