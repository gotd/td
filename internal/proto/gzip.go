package proto

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"

	"golang.org/x/xerrors"

	"github.com/ernado/td/bin"
)

// GZIP represents a Packed Object.
//
// Used to replace any other object (or rather, a serialization thereof)
// with its archived (gzipped) representation.
type GZIP struct {
	Data []byte
}

const GZIPTypeID = 0x3072cfa1

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
