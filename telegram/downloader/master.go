package downloader

import (
	"context"
	"strconv"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// RedirectError error is returned when Downloader get CDN redirect.
// See https://core.telegram.org/constructor/upload.fileCdnRedirect.
type RedirectError struct {
	Redirect *tg.UploadFileCDNRedirect
}

// Error implements error interface.
func (r *RedirectError) Error() string {
	return "redirect to CDN DC " + strconv.Itoa(r.Redirect.DCID)
}

// master is a master DC download schema.
// See https://core.telegram.org/api/files#downloading-files.
type master struct {
	client Client

	precise  bool
	allowCDN bool
	location tg.InputFileLocationClass
}

var _ schema = master{}

func (c master) Chunk(ctx context.Context, offset int64, limit int) (chunk, error) {
	req := &tg.UploadGetFileRequest{
		Offset:   offset,
		Limit:    limit,
		Location: c.location,
	}
	req.SetCDNSupported(c.allowCDN)
	req.SetPrecise(c.precise)

	r, err := c.client.UploadGetFile(ctx, req)
	if err != nil {
		return chunk{}, err
	}

	switch result := r.(type) {
	case *tg.UploadFile:
		return chunk{data: result.Bytes, tag: result.Type}, nil
	case *tg.UploadFileCDNRedirect:
		return chunk{}, &RedirectError{Redirect: result}
	default:
		return chunk{}, errors.Errorf("unexpected type %T", r)
	}
}

func (c master) Hashes(ctx context.Context, offset int64) ([]tg.FileHash, error) {
	return c.client.UploadGetFileHashes(ctx, &tg.UploadGetFileHashesRequest{
		Location: c.location,
		Offset:   offset,
	})
}
