package proto

import (
	"bytes"
	"compress/gzip"
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

	if g.Data, err = ioutil.ReadAll(r); err != nil {
		return err
	}

	if err := r.Close(); err != nil {
		// This will verify checksum.
		return xerrors.Errorf("gzip error: %w", err)
	}

	return nil
}
