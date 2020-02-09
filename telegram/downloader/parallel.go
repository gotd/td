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

func (d *Downloader) parallel(ctx context.Context, rpc schema, threads int, w io.WriterAt) (tg.StorageFileTypeClass, error) {
	var typ tg.StorageFileTypeClass
	typOnce := &sync.Once{}

	grp := tdsync.NewCancellableGroup(ctx)
	toWrite := make(chan block, threads)
	ready := tdsync.NewReady()

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

					n := len(p.data)
					if n < 1 {
						typOnce.Do(func() {
							typ = p.tag
						})
						ready.Signal()
						return nil
					}

					select {
					case <-downloadCtx.Done():
						return downloadCtx.Err()
					case toWrite <- block{data: p.data, offset: partOffset}:
					}

					if n < d.partSize {
						typOnce.Do(func() {
							typ = p.tag
						})
						ready.Signal()
						return nil
					}
				}
			})
		}

		return downloads.Wait()
	})

	w = &syncWriterAt{w: w}
	// Write loop
	grp.Go(func(groupCtx context.Context) error {
		for {
			select {
			case <-groupCtx.Done():
				return groupCtx.Err()
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
	})

	return typ, grp.Wait()
}
