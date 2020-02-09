package downloader

import (
	"context"
	"strconv"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// RedirectError error is returned when Downloader get CDN redirect.
// See https://core.telegram.org/constructor/upload.fileCdnRedirect.
type RedirectError struct {
	Redirect *tg.UploadFileCdnRedirect
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

func (c master) Part(ctx context.Context, offset, limit int) (part, error) {
	req := &tg.UploadGetFileRequest{
		Offset:   offset,
		Limit:    limit,
		Location: c.location,
	}
	req.SetCDNSupported(c.allowCDN)
	req.SetPrecise(c.precise)

	r, err := c.client.UploadGetFile(ctx, req)
	if err != nil {
		return part{}, err
	}

	switch result := r.(type) {
	case *tg.UploadFile:
		return part{data: result.Bytes, tag: result.Type}, nil
	case *tg.UploadFileCdnRedirect:
		return part{}, &RedirectError{Redirect: result}
	default:
		return part{}, xerrors.Errorf("unexpected type %T", r)
	}
}

func (c master) Hashes(ctx context.Context, offset int) ([]tg.FileHash, error) {
	return c.client.UploadGetFileHashes(ctx, &tg.UploadGetFileHashesRequest{
		Location: c.location,
		Offset:   offset,
	})
}
