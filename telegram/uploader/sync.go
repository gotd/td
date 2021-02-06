package uploader

import (
	"io"
	"sync"
)

type syncReader struct {
	r   io.Reader
	mux sync.Mutex
}

func (s *syncReader) Read(p []byte) (n int, err error) {
	s.mux.Lock()
	n, err = s.r.Read(p)
	s.mux.Unlock()

	return
}
