package downloader

import (
	"context"
	"strings"
	"sync"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// RefreshFunc returns file location with fresh file reference, e.g. by
// fetching the message, story or profile photo that contains the file again.
//
// See https://core.telegram.org/api/file_reference and package fileref.
type RefreshFunc func(ctx context.Context) (tg.InputFileLocationClass, error)

// maxRefreshAttempts limits consecutive refreshes for one request.
const maxRefreshAttempts = 3

// IsFileReferenceError reports whether err means that the file reference
// of the location is expired or invalid, e.g. FILE_REFERENCE_EXPIRED,
// FILE_REFERENCE_INVALID or FILE_REFERENCE_<n>_EXPIRED.
func IsFileReferenceError(err error) bool {
	rpcErr, ok := tgerr.As(err)
	if !ok {
		return false
	}
	return strings.HasPrefix(rpcErr.Type, "FILE_REFERENCE_")
}

// locationRef is a file location shared by download workers, which can be
// refreshed on file reference errors.
type locationRef struct {
	mux      sync.Mutex
	location tg.InputFileLocationClass
	refresh  RefreshFunc
}

func (r *locationRef) get() tg.InputFileLocationClass {
	r.mux.Lock()
	defer r.mux.Unlock()
	return r.location
}

// update refreshes location if it is still the failed one: concurrent
// workers that failed with the same location do one refresh.
//
// Reports whether this call ran the refresh.
func (r *locationRef) update(ctx context.Context, failed tg.InputFileLocationClass) (bool, error) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.location != failed {
		// Already refreshed by another worker.
		return false, nil
	}

	location, err := r.refresh(ctx)
	if err != nil {
		return true, errors.Wrap(err, "refresh file reference")
	}
	if location == nil {
		return true, errors.New("refresh file reference: nil location")
	}
	r.location = location
	return true, nil
}

// WithRefresh sets callback to refresh file location on file reference
// errors (FILE_REFERENCE_*), so download continues from the same offset
// instead of failing.
//
// Concurrent workers share one refresh, and a request is retried at most
// three times. Has no effect for web files.
func (b *Builder) WithRefresh(refresh RefreshFunc) *Builder {
	if m, ok := b.schema.(master); ok && refresh != nil {
		m.ref = &locationRef{location: m.location, refresh: refresh}
		b.schema = m
	}
	return b
}
