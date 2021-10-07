package file

import (
	"io"
	"sync"
	"sync/atomic"

	"github.com/nnqq/td/internal/syncio"
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

// Storage is an abstraction for Telegram file storage.
type Storage interface {
	Open(name string) (File, error)
}

type memFile struct {
	syncio.BufWriterAt
	partSize int32
	_        [4]byte
}

func (m *memFile) Size() int {
	return m.Len()
}

func (m *memFile) Close() error {
	return nil
}

func (m *memFile) PartSize() int {
	return int(atomic.LoadInt32(&m.partSize))
}

func (m *memFile) SetPartSize(v int) {
	atomic.StoreInt32(&m.partSize, int32(v))
}

// InMemory is an inmemory implementation of file storage.
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
