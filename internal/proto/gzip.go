package proto

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/go-faster/errors"
	"github.com/klauspost/compress/gzip"
	"go.uber.org/multierr"

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
			closeErr = errors.Wrap(closeErr, "close")
			multierr.AppendInto(&rErr, closeErr)
		}
		gzipRWPool.PutWriter(w)
	}()
	if _, err := w.Write(g.Data); err != nil {
		return errors.Wrap(err, "compress")
	}
	if err := w.Close(); err != nil {
		return errors.Wrap(err, "close")
	}

	// Writing compressed data as bytes.
	b.PutBytes(buf.Bytes())

	return nil
}

type countReader struct {
	reader io.Reader
	read   int64
}

func (c *countReader) Total() int64 {
	return atomic.LoadInt64(&c.read)
}

func (c *countReader) Read(p []byte) (n int, err error) {
	n, err = c.reader.Read(p)
	atomic.AddInt64(&c.read, int64(n))
	return n, err
}

// DecompressionBombErr means that GZIP decode detected decompression bomb
// which decompressed payload is significantly higher than initial compressed
// size and stopped decompression to prevent OOM.
type DecompressionBombErr struct {
	Compressed   int
	Decompressed int
}

func (d *DecompressionBombErr) Error() string {
	return fmt.Sprintf("payload too big (expanded %d bytes to greater than %d)",
		d.Compressed, d.Decompressed,
	)
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
		return errors.Wrap(err, "gzip error")
	}
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			closeErr = errors.Wrap(closeErr, "close")
			multierr.AppendInto(&rErr, closeErr)
		}
		gzipRWPool.PutReader(r)
	}()

	// Apply mitigation for reading too much data which can result in OOM.
	const maxUncompressedSize = 1024 * 1024 * 10 // 10 mb
	reader := &countReader{
		reader: io.LimitReader(r, maxUncompressedSize),
	}
	if g.Data, err = io.ReadAll(reader); err != nil {
		return errors.Wrap(err, "decompress")
	}
	if reader.Total() >= maxUncompressedSize {
		// Read limit reached, possible decompression bomb detected.
		return errors.Wrap(&DecompressionBombErr{
			Compressed:   maxUncompressedSize,
			Decompressed: int(reader.Total()),
		}, "decompress")
	}

	if err := r.Close(); err != nil {
		return errors.Wrap(err, "checksum")
	}

	return nil
}
