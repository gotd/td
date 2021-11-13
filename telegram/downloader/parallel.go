package downloader

import (
	"context"
	"io"
	"sync"

	"github.com/go-faster/errors"

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
	g := tdsync.NewCancellableGroup(ctx)
	toWrite := make(chan block, threads)

	stop := func(t tg.StorageFileTypeClass) {
		typOnce.Do(func() {
			typ = t
		})
		ready.Signal()
	}

	// Download loop
	g.Go(func(ctx context.Context) error {
		downloads := tdsync.NewCancellableGroup(ctx)
		defer close(toWrite)

		for i := 0; i < threads; i++ {
			downloads.Go(func(ctx context.Context) error {
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-ready.Ready():
						return nil
					default:
					}

					b, err := r.Next(ctx)
					if err != nil {
						return errors.Wrap(err, "get file")
					}

					// If returned chunk is zero, that means we read all file.
					n := len(b.data)
					if n < 1 {
						stop(b.tag)
						return nil
					}

					select {
					case <-ctx.Done():
						return ctx.Err()
					case toWrite <- b:
					}

					if b.last() {
						stop(b.tag)
						return nil
					}
				}
			})
		}

		return downloads.Wait()
	})

	// Write loop
	g.Go(writeAtLoop(syncio.NewWriterAt(w), toWrite))

	return typ, g.Wait()
}
