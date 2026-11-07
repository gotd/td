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
	// retryHandler observes retried transient downloader errors.
	retryHandler RetryHandler
	location     tg.InputFileLocationClass
	// ref is refreshable location, nil if refresh is not set.
	ref *locationRef
}

func (c master) currentLocation() tg.InputFileLocationClass {
	if c.ref != nil {
		return c.ref.get()
	}
	return c.location
}

// withRefresh calls f with current location, refreshing it on file reference
// errors if refresh is set.
func withRefresh[T any](ctx context.Context, c master, f func(loc tg.InputFileLocationClass) (T, error)) (T, error) {
	for attempt := 0; ; attempt++ {
		loc := c.currentLocation()
		r, err := f(loc)
		if err == nil || c.ref == nil || attempt >= maxRefreshAttempts || !IsFileReferenceError(err) {
			return r, err
		}
		c.reportRetry(RetryOperationFileReference, attempt+1, err)
		if err := c.ref.update(ctx, loc); err != nil {
			var zero T
			return zero, err
		}
	}
}

var _ schema = master{}

func (c master) reportRetry(operation string, attempt int, err error) {
	if attempt < 1 || err == nil || c.retryHandler == nil {
		return
	}
	c.retryHandler(RetryEvent{
		Operation: operation,
		Attempt:   attempt,
		Err:       err,
	})
}

func (c master) Chunk(ctx context.Context, offset int64, limit int) (chunk, error) {
	r, err := withRefresh(ctx, c, func(loc tg.InputFileLocationClass) (tg.UploadFileClass, error) {
		req := &tg.UploadGetFileRequest{
			Offset:   offset,
			Limit:    limit,
			Location: loc,
		}
		req.SetCDNSupported(c.allowCDN)
		req.SetPrecise(c.precise)
		return c.client.UploadGetFile(ctx, req)
	})
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
	return withRefresh(ctx, c, func(loc tg.InputFileLocationClass) ([]tg.FileHash, error) {
		return c.client.UploadGetFileHashes(ctx, &tg.UploadGetFileHashesRequest{
			Location: loc,
			Offset:   offset,
		})
	})
}
