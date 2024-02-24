package syncio

import (
	"io"
	"sync"

	"github.com/go-faster/errors"
)

// WriterAt is synchronized io.WriterAt.
type WriterAt struct {
	w   io.WriterAt
	mux sync.Mutex
}

// NewWriterAt creates new WriterAt.
func NewWriterAt(w io.WriterAt) *WriterAt {
	return &WriterAt{w: w}
}

// WriteAt implements io.WriterAt.
func (s *WriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	s.mux.Lock()
	n, err = s.w.WriteAt(p, off)
	s.mux.Unlock()

	return
}

// BufWriterAt is synchronized buffer which implements io.WriterAt.
type BufWriterAt struct {
	buf []byte
	mux sync.RWMutex
}

// Bytes returns copy of data from buffer.
func (b *BufWriterAt) Bytes() (r []byte) {
	b.mux.RLock()
	defer b.mux.RUnlock()

	return append(make([]byte, 0, len(b.buf)), b.buf...)
}

// Len returns buffer available data size.
func (b *BufWriterAt) Len() int {
	b.mux.RLock()
	defer b.mux.RUnlock()

	return len(b.buf)
}

// ReadAt implements io.ReaderAt.
func (b *BufWriterAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, errors.Errorf("invalid offset %d", off)
	}

	b.mux.RLock()
	defer b.mux.RUnlock()

	l := int64(len(b.buf))
	switch {
	case off >= l:
		return 0, nil
	case off+int64(len(p)) >= l:
		r := b.buf[off:]
		copy(p, r)
		return len(r), nil
	}

	from := off
	to := off + int64(len(p))
	copy(p, b.buf[from:to])
	return len(p), nil
}

// WriteAt implements io.WriterAt.
func (b *BufWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, errors.Errorf("invalid offset %d", off)
	}

	b.mux.Lock()
	defer b.mux.Unlock()

	ends := len(p) + int(off)
	if len(b.buf) < ends {
		newBuf := make([]byte, ends)
		copy(newBuf, b.buf)
		b.buf = newBuf
	}

	from := off
	to := off + int64(len(p))
	copy(b.buf[from:to], p)
	return len(p), nil
}
