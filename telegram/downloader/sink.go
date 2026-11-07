package downloader

import (
	"context"
	"io"

	"github.com/go-faster/errors"
)

// sink tracks download progress.
type sink struct {
	done     int64
	total    int64
	progress ProgressFunc
}

func (s *sink) written(n int) {
	s.done += int64(n)
	if s.progress != nil {
		s.progress(s.done, s.total)
	}
}

func writeAtLoop(w io.WriterAt, s sink, toWrite <-chan block) func(context.Context) error {
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
				s.written(len(part.data))
			}
		}
	}
}

func writeLoop(w io.Writer, s sink, toWrite <-chan block) func(context.Context) error {
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
				s.written(len(part.data))
			}
		}
	}
}
