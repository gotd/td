package proto

import (
	"bytes"
	"io"
	"sync"

	"github.com/klauspost/compress/gzip"
	"go.uber.org/multierr"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
)

type gzipPool struct {
	writers sync.Pool
	readers sync.Pool
}

func newGzipPool() *gzipPool {
	return &gzipPool{
		writers: sync.Pool{
			New: func() interface{} {
				return gzip.NewWriter(nil)
			},
		},
		readers: sync.Pool{},
	}
}

func (g *gzipPool) GetWriter(w io.Writer) *gzip.Writer {
	writer := g.writers.Get().(*gzip.Writer)
	writer.Reset(w)
	return writer
}

func (g *gzipPool) PutWriter(w *gzip.Writer) {
	g.writers.Put(w)
}

func (g *gzipPool) GetReader(r io.Reader) (*gzip.Reader, error) {
	reader, ok := g.readers.Get().(*gzip.Reader)
	if !ok {
		r, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	if err := reader.Reset(r); err != nil {
		g.readers.Put(reader)
		return nil, err
	}
	return reader, nil
}

func (g *gzipPool) PutReader(w *gzip.Reader) {
	g.readers.Put(w)
}

// GZIP represents a Packed Object.
//
// Used to replace any other object (or rather, a serialization thereof)
// with its archived (gzipped) representation.
type GZIP struct {
	Data []byte
}

// GZIPTypeID is TL type id of GZIP.
const GZIPTypeID = 0x3072cfa1

// nolint:gochecknoglobals
var (
	gzipRWPool  = newGzipPool()
	gzipBufPool = sync.Pool{New: func() interface{} {
		return bytes.NewBuffer(nil)
	}}
)

// Encode implements bin.Encoder.
func (g GZIP) Encode(b *bin.Buffer) (rErr error) {
	b.PutID(GZIPTypeID)

	// Writing compressed data to buf.
	buf := gzipBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer gzipBufPool.Put(buf)

	w := gzipRWPool.GetWriter(buf)
	defer func() {
		if closeErr := w.Close(); closeErr != nil {
			closeErr = xerrors.Errorf("close: %w", closeErr)
			multierr.AppendInto(&rErr, closeErr)
		}
		gzipRWPool.PutWriter(w)
	}()
	if _, err := w.Write(g.Data); err != nil {
		return xerrors.Errorf("compress: %w", err)
	}
	if err := w.Close(); err != nil {
		return xerrors.Errorf("close: %w", err)
	}

	// Writing compressed data as bytes.
	b.PutBytes(buf.Bytes())

	return nil
}

// Decode implements bin.Decoder.
func (g *GZIP) Decode(b *bin.Buffer) (rErr error) {
	if err := b.ConsumeID(GZIPTypeID); err != nil {
		return err
	}
	buf, err := b.Bytes()
	if err != nil {
		return err
	}

	r, err := gzipRWPool.GetReader(bytes.NewReader(buf))
	if err != nil {
		return xerrors.Errorf("gzip error: %w", err)
	}
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			closeErr = xerrors.Errorf("close: %w", closeErr)
			multierr.AppendInto(&rErr, closeErr)
		}
		gzipRWPool.PutReader(r)
	}()

	// Apply mitigation for reading too much data which can result in OOM.
	const maxUncompressedSize = 1024 * 1024 * 10 // 10 mb
	// TODO(ernado): fail explicitly if limit is reached
	// Currently we just return nil, but it is better than failing with OOM.
	if g.Data, err = io.ReadAll(io.LimitReader(r, maxUncompressedSize)); err != nil {
		return xerrors.Errorf("decompress: %w", err)
	}

	if err := r.Close(); err != nil {
		// This will verify checksum only if limit is not reached.
		return xerrors.Errorf("checksum: %w", err)
	}

	return nil
}
