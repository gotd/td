package downloader

import (
	"context"
	"io"
	"sync"

	"go.uber.org/atomic"
	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/tg"
)

type block struct {
	data   []byte
	offset int64
}

func writeAtLoop(w io.WriterAt, toWrite <-chan block) func(context.Context) error {
	return func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case part, ok := <-toWrite:
				if !ok {
					return nil
				}

				_, err := w.WriteAt(part.data, part.offset)
				if err != nil {
					return xerrors.Errorf("write output: %w", err)
				}
			}
		}
	}
}

// nolint:gocognit
func (d *Downloader) parallel(
	ctx context.Context,
	rpc schema,
	threads int,
	w io.WriterAt,
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
		offset := atomic.NewInt64(int64(-d.partSize))
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

					partOffset := offset.Add(int64(d.partSize))
					p, err := rpc.Part(downloadCtx, int(partOffset), d.partSize)
					if err != nil {
						return xerrors.Errorf("get file: %w", err)
					}

					// If returned part is zero, that means we read all file.
					n := len(p.data)
					if n < 1 {
						stop(p.tag)
						return nil
					}

					select {
					case <-downloadCtx.Done():
						return downloadCtx.Err()
					case toWrite <- block{data: p.data, offset: partOffset}:
					}

					// If returned part is less than requested, that means it is the last part.
					if n < d.partSize {
						stop(p.tag)
						return nil
					}
				}
			})
		}

		return downloads.Wait()
	})

	// Write loop
	grp.Go(writeAtLoop(&syncWriterAt{w: w}, toWrite))

	return typ, grp.Wait()
}
