package downloader

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

var errHashesNotSupported = xerrors.New("this schema does not support hashes fetch")

// web is a web file download schema.
// See https://core.telegram.org/api/files#downloading-webfiles.
type web struct {
	client Client

	location tg.InputWebFileLocationClass
}

var _ schema = web{}

func (w web) Part(ctx context.Context, offset, limit int) (part, error) {
	file, err := w.client.UploadGetWebFile(ctx, &tg.UploadGetWebFileRequest{
		Location: w.location,
		Offset:   offset,
		Limit:    limit,
	})
	if err != nil {
		return part{}, err
	}

	return part{data: file.Bytes, tag: file.FileType}, nil
}

func (w web) Hashes(ctx context.Context, offset int) ([]tg.FileHash, error) {
	return nil, errHashesNotSupported
}
