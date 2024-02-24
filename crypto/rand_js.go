package crypto

import (
	"bufio"
	"crypto/rand"
	"io"
	"sync"
)

var defaultRand struct {
	sync.Once
	reader io.Reader
}

// DefaultRand returns default entropy source.
func DefaultRand() io.Reader {
	defaultRand.Do(func() {
		defaultRand.reader = bufio.NewReaderSize(rand.Reader, 1024)
	})
	return rand.Reader
}
