package downloader

import (
	"io"
	"sync"
)

type syncWriterAt struct {
	w   io.WriterAt
	mux sync.Mutex
}

func (s *syncWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	s.mux.Lock()
	n, err = s.w.WriteAt(p, off)
	s.mux.Unlock()

	return
}
