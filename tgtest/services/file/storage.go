package file

import (
	"io"
	"sync"

	"go.uber.org/atomic"

	"github.com/gotd/td/internal/syncio"
)

// File represents Telegram file.
type File interface {
	io.ReaderAt
	io.WriterAt
	io.Closer
	PartSize() int
	SetPartSize(v int)
	Size() int
}

// Storage is a abstraction for Telegram file storage.
type Storage interface {
	Open(name string) (File, error)
}

type memFile struct {
	syncio.BufWriterAt
	partSize atomic.Int64
}

func (m *memFile) Size() int {
	return m.Len()
}

func (m *memFile) Close() error {
	return nil
}

func (m *memFile) PartSize() int {
	return int(m.partSize.Load())
}

func (m *memFile) SetPartSize(v int) {
	m.partSize.Store(int64(v))
}

// InMemory is a inmemory implementation of file storage.
type InMemory struct {
	files    map[string]*memFile
	filesMux sync.Mutex
}

// NewInMemory creates new InMemory.
func NewInMemory() *InMemory {
	return &InMemory{
		files: map[string]*memFile{},
	}
}

// Open implement Storage.
func (i *InMemory) Open(name string) (File, error) {
	i.filesMux.Lock()
	defer i.filesMux.Unlock()
	file, ok := i.files[name]
	if !ok {
		file = &memFile{}
		i.files[name] = file
	}

	return file, nil
}
