// Package uploader contains uploading files helpers.
package uploader

import (
	"io"
	"os"
	"path/filepath"

	"go.uber.org/atomic"

	"golang.org/x/xerrors"
)

// Reader is a file upload reader.
type Reader interface {
	io.Reader
	io.ReaderAt
}

// NewUpload creates new Upload struct using given
// name and reader.
func NewUpload(name string, from Reader, total int64) *Upload {
	return &Upload{
		name:       name,
		totalBytes: total,
		from:       from,
		fromAt:     from,
		partSize:   -1,
	}
}

// FromReader creates new Upload struct using
// given io.Reader.
// Note: Upload created with this builder will not be seekable, so upload can't be repeatable.
func FromReader(name string, from io.Reader, total int64) *Upload {
	return &Upload{
		name:       name,
		totalBytes: total,
		from:       from,
		partSize:   -1,
	}
}

// File is file abstraction.
type File interface {
	Name() string
	Stat() (os.FileInfo, error)

	io.Reader
	io.ReaderAt
}

var _ File = (*os.File)(nil)

// FromFile creates new Upload struct using
// given File.
func FromFile(f File) (*Upload, error) {
	info, err := f.Stat()
	if err != nil {
		return nil, xerrors.Errorf("stat: %w", err)
	}

	return NewUpload(f.Name(), f, info.Size()), nil
}

// FromPath creates new Upload struct using
// given path.
func FromPath(path string) (*Upload, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, xerrors.Errorf("open: %w", err)
	}

	return FromFile(f)
}

// Upload represents Telegram file upload.
type Upload struct {
	// Fields which will be set by Uploader.
	// File ID for Telegram.
	id int64
	// Sent parts (in partSize).
	sentParts atomic.Int64
	// Confirmed uploaded parts.
	confirmedParts atomic.Int64
	// Confirmed uploaded bytes.
	confirmedBytes atomic.Int64
	// Total parts.
	totalParts int
	// Part size of uploader.
	partSize int
	// Flag to determine class of size of file.
	big bool

	// Total size (in bytes) of upload.
	totalBytes int64 // immutable
	// Name of file.
	name string // immutable
	// Reader of data.
	from io.Reader // immutable
	// Seekable reader of data.
	fromAt io.ReaderAt // immutable
}

func (u *Upload) confirm(bytes int) (uploaded, parts int) {
	uploaded = int(u.confirmedBytes.Add(int64(bytes)))
	parts = int(u.confirmedParts.Inc())
	return
}
