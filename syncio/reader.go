package syncio

import (
	"io"
	"sync"
)

// Reader is synchronized io.Reader.
type Reader struct {
	r   io.Reader
	mux sync.Mutex
}

// NewReader creates new Reader.
func NewReader(r io.Reader) *Reader {
	return &Reader{r: r}
}

// Read implements io.Reader.
func (s *Reader) Read(p []byte) (n int, err error) {
	s.mux.Lock()
	n, err = s.r.Read(p)
	s.mux.Unlock()

	return
}
