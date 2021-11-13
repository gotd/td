package session

import (
	"context"
	"os"
	"sync"

	"github.com/go-faster/errors"
)

// FileStorage implements SessionStorage for file system as file
// stored in Path.
type FileStorage struct {
	Path string
	mux  sync.Mutex
}

// LoadSession loads session from file.
func (f *FileStorage) LoadSession(_ context.Context) ([]byte, error) {
	if f == nil {
		return nil, errors.New("nil session storage is invalid")
	}

	f.mux.Lock()
	defer f.mux.Unlock()

	data, err := os.ReadFile(f.Path)
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "read")
	}

	return data, nil
}

// StoreSession stores session to file.
func (f *FileStorage) StoreSession(_ context.Context, data []byte) error {
	if f == nil {
		return errors.New("nil session storage is invalid")
	}

	f.mux.Lock()
	defer f.mux.Unlock()
	// TODO(tdakkota): use robustio/renameio?
	return os.WriteFile(f.Path, data, 0600)
}
