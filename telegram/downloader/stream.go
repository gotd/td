package downloader

import (
	"context"
	"io"
	"sync"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/tg"
)

func (d *Downloader) stream(ctx context.Context, r *reader, w io.Writer) (tg.StorageFileTypeClass, error) {
	var typ tg.StorageFileTypeClass
	typOnce := &sync.Once{}

	grp := tdsync.NewCancellableGroup(ctx)
	toWrite := make(chan block, 1)
	// Download loop
	grp.Go(func(groupCtx context.Context) error {
		for {
			b, err := r.Next(ctx)
			if err != nil {
				return xerrors.Errorf("get file: %w", err)
			}

			n := len(b.data)
			if n < 1 {
				typOnce.Do(func() {
					typ = b.tag
					close(toWrite)
				})
				return nil
			}

			select {
			case <-groupCtx.Done():
				return groupCtx.Err()
			case toWrite <- b:
			}
		}
	})

	// Write loop
	grp.Go(writeLoop(w, toWrite))

	return typ, grp.Wait()
}
