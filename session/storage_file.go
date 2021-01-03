package session

import (
	"context"
	"io/ioutil"
	"os"

	"golang.org/x/xerrors"
)

// FileStorage implements SessionStorage for file system as file
// stored in Path.
type FileStorage struct {
	Path string
}

// LoadSession loads session from file.
func (f *FileStorage) LoadSession(_ context.Context) ([]byte, error) {
	if f == nil {
		return nil, xerrors.New("nil session storage is invalid")
	}
	data, err := ioutil.ReadFile(f.Path)
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, xerrors.Errorf("read: %w", err)
	}
	return data, nil
}

// StoreSession stores session to file.
func (f *FileStorage) StoreSession(_ context.Context, data []byte) error {
	if f == nil {
		return xerrors.New("nil session storage is invalid")
	}
	return ioutil.WriteFile(f.Path, data, 0600)
}
