package proto

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
)

// GZIP represents a Packed Object.
//
// Used to replace any other object (or rather, a serialization thereof)
// with its archived (gzipped) representation.
type GZIP struct {
	Data []byte
}

// GZIPTypeID is TL type id of GZIP.
const GZIPTypeID = 0x3072cfa1

// Encode implements bin.Encoder.
func (g GZIP) Encode(b *bin.Buffer) error {
	b.PutID(GZIPTypeID)

	// Writing compressed data to buf.
	buf := new(bytes.Buffer)
	w := gzip.NewWriter(buf)
	if _, err := io.Copy(w, bytes.NewReader(g.Data)); err != nil {
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
func (g *GZIP) Decode(b *bin.Buffer) error {
	if err := b.ConsumeID(GZIPTypeID); err != nil {
		return err
	}
	buf, err := b.Bytes()
	if err != nil {
		return err
	}

	r, err := gzip.NewReader(bytes.NewReader(buf))
	if err != nil {
		return xerrors.Errorf("gzip error: %w", err)
	}
	defer func() { _ = r.Close() }()

	// Apply mitigation for DOS via gzip bomp.
	const maxUncompressedSize = 1024 * 1024 * 10 // 10 mb
	// TODO(ernado): fail explicitly if limit is reached
	// Currently we just return nil, but it is better than failing with OOM.
	if g.Data, err = ioutil.ReadAll(io.LimitReader(r, maxUncompressedSize)); err != nil {
		return xerrors.Errorf("decompress: %w", err)
	}

	if err := r.Close(); err != nil {
		// This will verify checksum only if limit is not reached.
		return xerrors.Errorf("checksum: %w", err)
	}

	return nil
}
