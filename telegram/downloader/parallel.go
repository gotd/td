package downloader

import (
	"context"
	"io"
	"sync"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/syncio"
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/tg"
)

// nolint:gocognit
func (d *Downloader) parallel(
	ctx context.Context, r *reader,
	threads int, w io.WriterAt,
) (tg.StorageFileTypeClass, error) {
	var typ tg.StorageFileTypeClass
	typOnce := &sync.Once{}

	ready := tdsync.NewReady()
	grp := tdsync.NewCancellableGroup(ctx)
	toWrite := make(chan block, threads)

	stop := func(t tg.StorageFileTypeClass) {
		typOnce.Do(func() {
			typ = t
		})
		ready.Signal()
	}

	// Download loop
	grp.Go(func(groupCtx context.Context) error {
		downloads := tdsync.NewCancellableGroup(groupCtx)
		defer close(toWrite)

		for i := 0; i < threads; i++ {
			downloads.Go(func(downloadCtx context.Context) error {
				for {
					select {
					case <-downloadCtx.Done():
						return downloadCtx.Err()
					case <-ready.Ready():
						return nil
					default:
					}

					b, err := r.Next(downloadCtx)
					if err != nil {
						return xerrors.Errorf("get file: %w", err)
					}

					// If returned chunk is zero, that means we read all file.
					n := len(b.data)
					if n < 1 {
						stop(b.tag)
						return nil
					}

					select {
					case <-downloadCtx.Done():
						return downloadCtx.Err()
					case toWrite <- b:
					}
				}
			})
		}

		return downloads.Wait()
	})

	// Write loop
	grp.Go(writeAtLoop(syncio.NewWriterAt(w), toWrite))

	return typ, grp.Wait()
}
