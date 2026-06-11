package uploader

import (
	"context"
	"io"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/syncio"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

type part struct {
	id     int
	buf    *bin.Buffer
	upload *Upload
}

func (u *Uploader) uploadBigFilePart(ctx context.Context, p part) (int, error) {
	defer p.upload.pool.Put(p.buf)

	// Upload loop.
	for {
		r, err := u.rpc.UploadSaveBigFilePart(ctx, &tg.UploadSaveBigFilePartRequest{
			FileID:         p.upload.id,
			FilePart:       p.id,
			FileTotalParts: p.upload.totalParts,
			Bytes:          p.buf.Buf,
		})

		if flood, err := tgerr.FloodWait(ctx, err); err != nil {
			if flood {
				continue
			}
			return 0, errors.Wrapf(err, "send upload part %d RPC", p.id)
		}

		// If Telegram returned false, it seems save is not successful, so we retry to send.
		if r {
			return p.buf.Len(), nil
		}
	}
}

func (u *Uploader) bigLoop(ctx context.Context, threads int, upload *Upload) error { // nolint:gocognit
	g := tdsync.NewCancellableGroup(ctx)
	toSend := make(chan part, threads)

	// Run read loop
	r := syncio.NewReader(upload.from)
	g.Go(func(ctx context.Context) error {
		last := false
		totalStreamSize := 0

		for {
			buf := upload.pool.GetSize(upload.partSize)

			n, err := io.ReadFull(r, buf.Buf)
			if n > 0 {
				totalStreamSize += n
			}
			switch {
			case errors.Is(err, io.ErrUnexpectedEOF):
				last = true
				if upload.totalParts == -1 {
					totalParts := (totalStreamSize + upload.partSize - 1) / upload.partSize
					upload.totalParts = int(totalParts)
				}
			case errors.Is(err, io.EOF):
				upload.pool.Put(buf)

				close(toSend)
				return nil
			case err != nil:
				upload.pool.Put(buf)

				return errors.Wrap(err, "read source")
			}

			buf.Buf = buf.Buf[:n]
			nextPart := part{
				id:     int(upload.sentParts.Load()),
				buf:    buf,
				upload: upload,
			}
			select {
			case toSend <- nextPart:
				upload.sentParts.Inc()
				if last {
					close(toSend)
					return nil
				}
			case <-ctx.Done():
				upload.pool.Put(buf)

				return ctx.Err()
			}
		}
	})

	for i := 0; i < threads; i++ {
		g.Go(func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case part, ok := <-toSend:
					if !ok {
						return nil
					}

					n, err := u.uploadBigFilePart(ctx, part)
					if err != nil {
						return errors.Wrap(err, "upload part")
					}

					if err := u.callback(ctx, upload.confirm(part.id, n)); err != nil {
						return errors.Wrap(err, "progress callback")
					}
				}
			}
		})
	}

	return g.Wait()
}
