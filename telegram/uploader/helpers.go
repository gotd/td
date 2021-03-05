package uploader

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// File is file abstraction.
type File interface {
	Stat() (os.FileInfo, error)
	io.Reader
}

// FromFile uploads given File.
// NB: FromFile does not close given file.
func (u *Uploader) FromFile(ctx context.Context, f File) (tg.InputFileClass, error) {
	info, err := f.Stat()
	if err != nil {
		return nil, xerrors.Errorf("stat: %w", err)
	}

	return u.Upload(ctx, NewUpload(info.Name(), f, info.Size()))
}

// FromPath uploads file from given path.
func (u *Uploader) FromPath(ctx context.Context, path string) (tg.InputFileClass, error) {
	return u.FromFS(ctx, osFS{}, path)
}

type osFS struct{}

func (o osFS) Open(name string) (fs.File, error) {
	return os.Open(filepath.Clean(name))
}

// FromFS uploads file from fs using given path.
func (u *Uploader) FromFS(ctx context.Context, filesystem fs.FS, path string) (tg.InputFileClass, error) {
	f, err := filesystem.Open(path)
	if err != nil {
		return nil, xerrors.Errorf("open: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	return u.FromFile(ctx, f)
}

// FromReader uploads file from given io.Reader.
// NB: totally stream should not exceed the limit for
// small files (10 MB as docs says, may be a bit bigger).
func (u *Uploader) FromReader(ctx context.Context, name string, f io.Reader) (tg.InputFileClass, error) {
	return u.Upload(ctx, NewUpload(name, f, -1))
}

// FromBytes uploads file from given byte slice.
func (u *Uploader) FromBytes(ctx context.Context, name string, b []byte) (tg.InputFileClass, error) {
	return u.Upload(ctx, NewUpload(name, bytes.NewReader(b), int64(len(b))))
}
