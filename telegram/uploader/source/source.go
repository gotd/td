package source

import (
	"context"
	"io"
	"net/url"
)

// RemoteFile is abstraction for remote file.
type RemoteFile interface {
	io.ReadCloser
	// Name returns filename. Should not be empty.
	Name() string
	// Size returns size of file. If size is unknown, -1 should be returned.
	Size() int64
}

// Source is abstraction for remote upload source.
type Source interface {
	Open(ctx context.Context, u *url.URL) (RemoteFile, error)
}
