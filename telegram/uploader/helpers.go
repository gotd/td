package uploader

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// File is file abstraction.
type File interface {
	Name() string
	Stat() (os.FileInfo, error)
	io.Reader
}

var _ File = (*os.File)(nil)

// FromFile uploads given File.
// NB: UploadFromFile does not close given file.
func (u *Uploader) FromFile(ctx context.Context, f File) (tg.InputFileClass, error) {
	info, err := f.Stat()
	if err != nil {
		return nil, xerrors.Errorf("stat: %w", err)
	}

	return u.Upload(ctx, NewUpload(f.Name(), f, info.Size()))
}

// FromPath uploads file from given path.
func (u *Uploader) FromPath(ctx context.Context, path string) (tg.InputFileClass, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, xerrors.Errorf("open: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	return u.FromFile(ctx, f)
}
