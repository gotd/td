package downloader

import (
	"context"
	"io"

	"golang.org/x/xerrors"
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
					return xerrors.Errorf("write output: %w", err)
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
					return xerrors.Errorf("write output: %w", err)
				}
			}
		}
	}
}
