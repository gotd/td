package uploader

import (
	"context"
	"io"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func (u *Uploader) smallLoop(ctx context.Context, h io.Writer, upload *Upload) error {
	buf := u.pool.GetSize(u.partSize)
	defer u.pool.Put(buf)

	last := false

	r := io.TeeReader(upload.from, h)
	for {
		n, err := io.ReadFull(r, buf.Buf)
		switch {
		case errors.Is(err, io.ErrUnexpectedEOF):
			last = true
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return errors.Wrap(err, "read source")
		}
		read := buf.Buf[:n]

		// Upload loop.
		for {
			r, err := u.rpc.UploadSaveFilePart(ctx, &tg.UploadSaveFilePartRequest{
				FileID:   upload.id,
				FilePart: int(upload.sentParts.Load()) % partsLimit,
				Bytes:    read,
			})

			if flood, err := tgerr.FloodWait(ctx, err); err != nil {
				if flood {
					continue
				}
				return errors.Wrap(err, "send upload RPC")
			}

			// If Telegram returned false, it seems save is not successful, so we retry to send.
			if !r {
				continue
			}

			break
		}

		upload.sentParts.Inc()
		if err := u.callback(ctx, upload.confirmSmall(n)); err != nil {
			return errors.Wrap(err, "progress callback")
		}

		if last {
			break
		}
	}

	return nil
}
