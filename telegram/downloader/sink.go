package downloader

import (
	"context"
	"io"

	"github.com/go-faster/errors"
)

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
					return errors.Wrap(err, "write output")
				}
			}
		}
	}
}

func writeLoop(w io.Writer, toWrite <-chan block) func(context.Context) error {
	return func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case part, ok := <-toWrite:
				if !ok {
					return nil
				}

				_, err := w.Write(part.data)
				if err != nil {
					return errors.Wrap(err, "write output")
				}
			}
		}
	}
}
